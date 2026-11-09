// Package failure gives errors a stable name that programs check and people can search for.
package failure

// Code is an error with a stable name. A code is one value made once, so returning it allocates
// nothing and comparing it compares two pointers.
type Code struct {
	name    string
	meaning string
}

// Define makes a code from its name and what it means.
func Define(name, meaning string) *Code {
	return &Code{name: name, meaning: meaning}
}

// Name is the code's stable name.
func (code *Code) Name() string { return code.name }

// Error gives the code's name and what it means.
func (code *Code) Error() string { return code.name + ": " + code.meaning }

// detailed is a code with what happened underneath.
type detailed struct {
	code    *Code
	details string
	cause   error
}

// Wrap adds details and the cause underneath to a code. The result answers to the code and to the
// cause through errors.Is. For cold paths: unlike a bare code, it allocates.
func Wrap(code *Code, details string, cause error) error {
	return &detailed{code: code, details: details, cause: cause}
}

// Error gives the code, then the details and the cause when there are any.
func (failure *detailed) Error() string {
	text := failure.code.Error()
	if failure.details != "" {
		text += " (" + failure.details + ")"
	}
	if failure.cause != nil {
		text += ": " + failure.cause.Error()
	}
	return text
}

// Is says whether this failure is the given code.
func (failure *detailed) Is(target error) bool { return target == error(failure.code) }

// Unwrap gives the cause underneath.
func (failure *detailed) Unwrap() error { return failure.cause }
