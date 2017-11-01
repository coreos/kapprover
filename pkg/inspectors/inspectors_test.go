package inspectors_test

import (
	"github.com/coreos/kapprover/pkg/inspectors"
	"testing"

	_ "github.com/coreos/kapprover/pkg/inspectors/group"
	_ "github.com/coreos/kapprover/pkg/inspectors/username"
)

func TestInspectors(t *testing.T) {
	var i inspectors.Inspectors

	actual := i.String()
	if actual != "" {
		t.Errorf("Expected default Inspectors.String() to be \"\", got %q", actual)
	}
	if len(i) != 0 {
		t.Errorf("Expected default Inspectors to have len 0, got %s", len(i))
	}

	i.Set("group")
	actual = i.String()
	if actual != "group" {
		t.Errorf("Expected Inspectors.String() to be \"group\", got %q", actual)
	}
	if len(i) != 1 {
		t.Errorf("Expected Inspectors to have len 1, got %s", len(i))
	}
	if i[0].Name != "group" {
		t.Errorf("Expected Inspectors[0].Name to be \"group\", got %s", i[0].Name)
	}

	i.Set("username")
	actual = i.String()
	if actual != "group,username" {
		t.Errorf("Expected Inspectors.String() to be \"group,username\", got %q", actual)
	}
	if len(i) != 2 {
		t.Errorf("Expected Inspectors to have len 2, got %s", len(i))
	}
	if i[1].Name != "username" {
		t.Errorf("Expected Inspectors[1].Name to be \"username\", got %s", i[1].Name)
	}

	i = inspectors.Inspectors{}
	i.Set("username,group")
	actual = i.String()
	if actual != "username,group" {
		t.Errorf("Expected Inspectors.String() to be \"username,group\", got %q", actual)
	}
	if len(i) != 2 {
		t.Errorf("Expected Inspectors to have len 2, got %s", len(i))
	}
	if i[0].Name != "username" {
		t.Errorf("Expected Inspectors[0].Name to be \"username\", got %s", i[1].Name)
	}
	if i[1].Name != "group" {
		t.Errorf("Expected Inspectors[1].Name to be \"grou0\", got %s", i[1].Name)
	}
}
