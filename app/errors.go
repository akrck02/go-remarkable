package app

import "github.com/akrck02/go-remarkable/failure"

// The codes of this package's errors, for programs to check with errors.Is.
var (
	NoFramebufferKey = failure.Define("app.no-framebuffer-key", "QTFB_KEY is not a number: was the program opened by AppLoad?")
	InterfaceFailed  = failure.Define("app.interface-failed", "Go UI or its fonts could not start")
)
