package main

import (
	"flag"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/coreos/kapprover/pkg/inspectors"
	_ "github.com/coreos/kapprover/pkg/inspectors/group"
	_ "github.com/coreos/kapprover/pkg/inspectors/username"
)

var (
	kubeconfigPath = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	filters        inspectors.Inspectors
	deniers        inspectors.Inspectors
	warners        inspectors.Inspectors
)

func init() {
	flag.Var(&filters, "filter", "comma-separated list of inspectors to filter the set of requests to handle")
	flag.Var(&deniers, "denier", "comma-separated list of inspectors to deny requests")
	flag.Var(&warners, "warner", "comma-separated list of inspectors to log warnings (but not block approval)")
}

func main() {
	flag.Parse()
	if len(filters) == 0 {
		filters.Set("username,group")
	}

	// Create a Kubernetes client.
	client, err := newClient(*kubeconfigPath)
	if err != nil {
		log.Errorf("Could not create Kubernetes client: %s", err)
		return
	}

	// Create a watcher and an informer for CertificateSigningRequests.
	// The Add function
	watchList := cache.NewListWatchFromClient(
		client.CertificatesV1alpha1Client.RESTClient(),
		"certificatesigningrequests",
		v1.NamespaceAll,
		fields.Everything(),
	)

	f := func(obj interface{}) {
		if req, ok := obj.(*certificates.CertificateSigningRequest); ok {
			if err := tryApprove(filters, deniers, warners, client, req); err != nil {
				log.Errorf("Failed to approve %q: %s", req.ObjectMeta.Name, err)
				return
			}
		}
	}

	_, controller := cache.NewInformer(
		watchList,
		&certificates.CertificateSigningRequest{},
		time.Second*30,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				f(obj)
			},
			UpdateFunc: func(_, obj interface{}) {
				f(obj)
			},
		},
	)

	controller.Run(make(chan struct{}))
}

func newClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if kubeconfigPath != "" {
		// Initialize a configuration from the provided kubeconfig.
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// Initialize a configuration based on the default service account.
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(config)
}

func tryApprove(filters inspectors.Inspectors, deniers inspectors.Inspectors, warners inspectors.Inspectors, client *kubernetes.Clientset, request *certificates.CertificateSigningRequest) error {
	for {
		// Verify that the CSR hasn't been approved or denied already.
		//
		// There are only two possible conditions (CertificateApproved and
		// CertificateDenied). Therefore if the CSR already has a condition,
		// it means that the request has already been approved or denied, and that
		// we should ignore the request.
		if len(request.Status.Conditions) > 0 {
			return nil
		}

		for _, filter := range filters {
			message, err := filter.Inspector.Inspect(client, request)
			if err != nil {
				return err
			}
			if message != "" {
				return nil
			}
		}

		condition := certificates.CertificateSigningRequestCondition{
			Type:    certificates.CertificateApproved,
			Reason:  "AutoApproved",
			Message: "Approved by kapprover",
		}

		for _, denier := range deniers {
			message, err := denier.Inspector.Inspect(client, request)
			if err != nil {
				return err
			}
			if message != "" {
				condition.Type = certificates.CertificateDenied
				condition.Reason = denier.Name
				condition.Message = message
				break
			}
		}

		if condition.Type == certificates.CertificateApproved {
			for _, warner := range warners {
				message, _ := warner.Inspector.Inspect(client, request)
				if message != "" {
					log.Warnf("Approving CSR from %s despite %s: %s", request.Spec.Username, warner.Name, message)
				}
			}
		}

		request.Status.Conditions = append(request.Status.Conditions, condition)

		// Submit the updated CSR.
		signingRequestInterface := client.CertificatesV1alpha1Client.CertificateSigningRequests()
		if _, err := signingRequestInterface.UpdateApproval(request); err != nil {
			if strings.Contains(err.Error(), "the object has been modified") {
				// The CSR might have been updated by a third-party, retry until we
				// succeed.
				request, err = signingRequestInterface.Get(request.ObjectMeta.Name)
				if err != nil {
					return err
				}
				continue
			}

			return err
		}

		log.Infof("Successfully approved %q", request.ObjectMeta.Name)

		return nil
	}
}
