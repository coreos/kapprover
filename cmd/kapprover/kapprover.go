package main

import (
	"flag"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/kapprover/pkg/approvers"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1beta1"
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
		client.CertificatesV1beta1Client.RESTClient(),
		"certificatesigningrequests",
		v1.NamespaceAll,
		fields.Everything(),
	)

	f := func(obj interface{}) {
		if req, ok := obj.(*certificates.CertificateSigningRequest); ok {
			if err := approver.Approve(client.CertificatesV1beta1Client.CertificateSigningRequests(), req); err != nil {
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
