package app

import (
	"strings"
	"testing"

	"github.com/akrck02/go-remarkable/failure"
)

func TestEveryCodeOfThisPackageHasItsOwnName(test *testing.T) {
	seen := map[string]bool{}
	for _, code := range []*failure.Code{NoFramebufferKey, InterfaceFailed} {
		if !strings.HasPrefix(code.Name(), "app.") || seen[code.Name()] {
			test.Errorf("bad or repeated name: %q", code.Name())
		}
		seen[code.Name()] = true
	}
}
