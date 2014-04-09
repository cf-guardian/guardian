package guardian

import (
	"github.com/cf-guardian/guardian/kernel"
	"testing"
)

type testResource struct{ a int }

func (testResource) Init() error {
	return nil
}

func run(res kernel.Resource) {

}

func TestRM(t *testing.T) {
	var x testResource
	run(x)
}
