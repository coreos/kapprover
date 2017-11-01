package username

import (
	"fmt"
	"github.com/coreos/kapprover/pkg/inspectors"
	"k8s.io/client-go/kubernetes"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
)

const (
	kubeletBootstrapUsername = "kubelet-bootstrap"
)

func init() {
	inspectors.Register("username", &username{})
}

// Username is an Inspector that verifies the CSR was submitted
// by the "kubelet-bootstrap" user.
type username struct{}

func (*username) Inspect(client *kubernetes.Clientset, request *certificates.CertificateSigningRequest) (string, error) {
	if request.Spec.Username != kubeletBootstrapUsername {
		return fmt.Sprintf("Requesting user %s is not %s", request.Spec.Username, kubeletBootstrapUsername), nil
	}

	return "", nil
}
