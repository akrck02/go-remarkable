package qtfb

import (
	"strings"
	"testing"

	"github.com/akrck02/go-remarkable/failure"
)

func TestEveryCodeOfThisPackageHasItsOwnName(test *testing.T) {
	seen := map[string]bool{}
	for _, code := range []*failure.Code{ServerNotFound, InitializeFailed, BadAnswer, SharedMemoryFailed, UpdateFailed} {
		if !strings.HasPrefix(code.Name(), "qtfb.") || seen[code.Name()] {
			test.Errorf("bad or repeated name: %q", code.Name())
		}
		seen[code.Name()] = true
	}
}
