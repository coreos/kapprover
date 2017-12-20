package approvers

import (
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1beta1"
)

var (
	approvers  = make(map[string]Approver)
	approversM sync.RWMutex
)

// Approver reprensents anything capable to validating and approving a CSR.
type Approver interface {
	Approve(v1beta1.CertificateSigningRequestInterface, *certificates.CertificateSigningRequest) error
}

// Register makes an Approver available by the provided name.
//
// If called twice with the same name, the name is blank, or if the provided
// Extractor is nil, this function panics.
func Register(name string, a Approver) {
	approversM.Lock()
	defer approversM.Unlock()

	if name == "" {
		panic("approvers: could not register an Approver with an empty name")
	}

	if a == nil {
		panic("approvers: could not register a nil Approver")
	}

	// Enforce lowercase names, so that they can be reliably be found in a map.
	name = strings.ToLower(name)

	if _, dup := approvers[name]; dup {
		panic("approvers: RegisterApprover called twice for " + name)
	}

	approvers[name] = a
}

// List returns the list of the registered approvers's name.
func List() []string {
	approversM.RLock()
	defer approversM.RUnlock()

	ret := make([]string, 0, len(approvers))
	for k := range approvers {
		ret = append(ret, k)
	}

	return ret
}

// Unregister removes a Approverwith a particular name from the list.
func Unregister(name string) {
	approversM.Lock()
	defer approversM.Unlock()
	delete(approvers, name)
}

// Get returns the registered Approver with a provided name.
func Get(name string) (a Approver, exists bool) {
	approversM.Lock()
	defer approversM.Unlock()

	a, exists = approvers[name]
	return
}
