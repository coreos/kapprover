package username_test

import (
	"github.com/coreos/kapprover/pkg/inspectors"
	"k8s.io/client-go/kubernetes"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1alpha1"
	"testing"

	_ "github.com/coreos/kapprover/pkg/inspectors/group"
)

var (
	client *kubernetes.Clientset
)

func TestInspect(t *testing.T) {
	inspector, exists := inspectors.Get("username")
	if !exists {
		t.Fatal("Expected inspectors.Get(\"username\") to exist, did not")
	}

	for username, expectedMessage := range map[string]string{
		"kubelet-bootstrap": "",
		"someone-else":      "Requesting user someone-else is not kubelet-bootstrap",
	} {
		request := certificates.CertificateSigningRequest{
			Spec: certificates.CertificateSigningRequestSpec{
				Username: username,
				Groups: []string{
					"someRandomGroup",
					"someOtherGroup",
				},
			},
		}

		message, err := inspector.Inspect(client, &request)

		if message != expectedMessage {
			t.Errorf("Group %s: expected %q, got %q", username, expectedMessage, message)
		}
		if err != nil {
			t.Errorf("Group %s: expected nil error, got %s", username, err)
		}
	}
}
