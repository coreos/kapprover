package always

import (
	"github.com/coreos/kapprover/pkg/approvers"
	"k8s.io/client-go/kubernetes/typed/certificates/v1alpha1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
)

const (
	kubeletBootstrapUsername = "kubelet-bootstrap"
	kubeletBootstrapGroup    = "system:kubelet-bootstrap"
)

func init() {
	approvers.Register("always", &Always{})
}

// Always is an Approver that automatically approves any pending CSR submitted
// by kubelets during their TLS bootstrapping process, without making any kind
// of validation besides checking that the requester's user/group are
// respectively kubeletBootstrapUsername / kubeletBootstrapGroup.
type Always struct{}

// Approve approves CSRs
func (*Always) Approve(client v1alpha1.CertificateSigningRequestInterface, request *certificates.CertificateSigningRequest) (certificates.CertificateSigningRequestCondition, error) {
	condition := certificates.CertificateSigningRequestCondition{
		Type:    certificates.CertificateApproved,
		Reason:  "AutoApproved",
		Message: "Auto approving of all kubelet CSRs is enabled on bootkube",
	}

	// Ensure the CSR has been submitted by a kubelet performing its TLS
	// bootstrapping by checking the username and the group.
	if request.Spec.Username != kubeletBootstrapUsername {
		return approvers.NoAction, nil
	}

	isKubeletBootstrapGroup := false
	for _, group := range request.Spec.Groups {
		if group == kubeletBootstrapGroup {
			isKubeletBootstrapGroup = true
			break
		}
	}
	if !isKubeletBootstrapGroup {
		return approvers.NoAction, nil
	}

	return condition, nil
}
