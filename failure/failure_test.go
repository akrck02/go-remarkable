package failure

import (
	"errors"
	"testing"
)

func TestAWrappedFailureAnswersToItsCodeAndItsCause(test *testing.T) {
	code := Define("example.broken", "something broke")
	cause := errors.New("the cause underneath")
	wrapped := Wrap(code, "a file", cause)
	if !errors.Is(wrapped, code) || !errors.Is(wrapped, cause) {
		test.Fatalf("got %v", wrapped)
	}
}

func TestAWrappedFailureDoesNotAnswerToAnotherCode(test *testing.T) {
	wrapped := Wrap(Define("example.one", "one"), "", nil)
	if errors.Is(wrapped, Define("example.two", "two")) {
		test.Fatal("it answered to a code it is not")
	}
}

func TestReturningACodeAllocatesNothing(test *testing.T) {
	code := Define("example.broken", "something broke")
	give := func() error { return code }
	if allocations := testing.AllocsPerRun(100, func() { _ = give() }); allocations != 0 {
		test.Fatalf("%v allocations", allocations)
	}
}
