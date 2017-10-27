package main

import (
	"flag"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/kapprover/pkg/approvers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
	certfiicateType "k8s.io/client-go/kubernetes/typed/certificates/v1alpha1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "github.com/coreos/kapprover/pkg/approvers/always"
)

var (
	kubeconfigPath = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	approverName   = flag.String("approver", "always", "name of the kubelet approver to use")
)

func main() {
	flag.Parse()

	// Create a Kubernetes client.
	client, err := newClient(*kubeconfigPath)
	if err != nil {
		log.Errorf("Could not create Kubernetes client: %s", err)
		return
	}

	// Get the requested approver.
	approver, exists := approvers.Get(*approverName)
	if !exists {
		log.Errorf(
			"Could not find approver %q, registered approvers: %s",
			*approverName,
			strings.Join(approvers.List(), ","),
		)
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
			if err := tryApprove(approver, client.CertificatesV1alpha1Client.CertificateSigningRequests(), req); err != nil {
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

func tryApprove(approver approvers.Approver, client certfiicateType.CertificateSigningRequestInterface, request *certificates.CertificateSigningRequest) error {
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

		condition, err := approver.Approve(client, request)
		if err != nil || condition.Type == "" {
			return err
		}
		request.Status.Conditions = append(request.Status.Conditions, condition)

		// Submit the updated CSR.
		if _, err = client.UpdateApproval(request); err != nil {
			if strings.Contains(err.Error(), "the object has been modified") {
				// The CSR might have been updated by a third-party, retry until we
				// succeed.
				request, err = client.Get(request.ObjectMeta.Name)
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