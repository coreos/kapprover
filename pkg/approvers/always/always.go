package always

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/kapprover/pkg/approvers"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1beta1"
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

// Approve approves CSRs in a loop.
func (*Always) Approve(client v1beta1.CertificateSigningRequestInterface, request *certificates.CertificateSigningRequest) error {
	condition := certificates.CertificateSigningRequestCondition{
		Type:    certificates.CertificateApproved,
		Reason:  "AutoApproved",
		Message: "Auto approving of all kubelet CSRs is enabled on bootkube",
	}

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

		// Ensure the CSR has been submitted by a kubelet performing its TLS
		// bootstrapping by checking the username and the group.
		if request.Spec.Username != kubeletBootstrapUsername {
			return nil
		}

		isKubeletBootstrapGroup := false
		for _, group := range request.Spec.Groups {
			if group == kubeletBootstrapGroup {
				isKubeletBootstrapGroup = true
				break
			}
		}
		if !isKubeletBootstrapGroup {
			return nil
		}

		// Approve the CSR.
		request.Status.Conditions = append(request.Status.Conditions, condition)

		// Submit the updated CSR.
		if _, err := client.UpdateApproval(request); err != nil {
			if strings.Contains(err.Error(), "the object has been modified") {
				// The CSR might have been updated by a third-party, retry until we
				// succeed.
				request, err = client.Get(request.ObjectMeta.Name, meta.GetOptions{})
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
