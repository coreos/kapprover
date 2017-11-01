package group

import (
	"fmt"
	"github.com/coreos/kapprover/pkg/inspectors"
	"k8s.io/client-go/kubernetes"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
)

const (
	kubeletBootstrapGroup = "system:kubelet-bootstrap"
)

func init() {
	inspectors.Register("group", &group{})
}

// Group is an Inspector that verifies the CSR was submitted
// by a user in the "system:kubelet-bootstrap".
type group struct{}

func (*group) Inspect(client *kubernetes.Clientset, request *certificates.CertificateSigningRequest) (string, error) {
	isKubeletBootstrapGroup := false
	for _, group := range request.Spec.Groups {
		if group == kubeletBootstrapGroup {
			isKubeletBootstrapGroup = true
			break
		}
	}
	if !isKubeletBootstrapGroup {
		return fmt.Sprintf("Requesting user %s is not in the %s group", request.Spec.Username, kubeletBootstrapGroup), nil
	}

	return "", nil
}
