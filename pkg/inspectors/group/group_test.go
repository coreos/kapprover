package group_test

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
	inspector, exists := inspectors.Get("group")
	if !exists {
		t.Fatal("Expected inspectors.Get(\"group\") to exist, did not")
	}

	for group, expectedMessage := range map[string]string{
		"system:kubelet-bootstrap": "",
		"someOtherGroup":           "Requesting user someRandomUser is not in the system:kubelet-bootstrap group",
	} {
		assertInspectionResult(t, inspector, group, expectedMessage)
	}
}

func TestInspectConfigured(t *testing.T) {
	inspector, exists := inspectors.Get("group")
	if !exists {
		t.Fatal("Expected inspectors.Get(\"group\") to exist, did not")
	}

	err := inspector.Configure("system:serviceaccount")
	if err != nil {
		t.Errorf("Expected Configure to not fail, got %s", err)
	}

	for group, expectedMessage := range map[string]string{
		"system:serviceaccount":    "",
		"system:kubelet-bootstrap": "Requesting user someRandomUser is not in the system:serviceaccount group",
	} {
		assertInspectionResult(t, inspector, group, expectedMessage)
	}
}

func assertInspectionResult(t *testing.T, inspector inspectors.Inspector, group string, expectedMessage string) {
	request := certificates.CertificateSigningRequest{
		Spec: certificates.CertificateSigningRequestSpec{
			Username: "someRandomUser",
			Groups: []string{
				"someRandomGroup",
				group,
			},
		},
	}
	message, err := inspector.Inspect(client, &request)
	if message != expectedMessage {
		t.Errorf("Username %s: expected %q, got %q", group, expectedMessage, message)
	}
	if err != nil {
		t.Errorf("Username %s: expected nil error, got %s", group, err)
	}
}
