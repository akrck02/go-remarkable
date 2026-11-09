package qtfb

import "github.com/akrck02/go-remarkable/failure"

// The codes of this package's errors, for programs to check with errors.Is.
var (
	ServerNotFound     = failure.Define("qtfb.server-not-found", "nothing answers on the qtfb socket")
	InitializeFailed   = failure.Define("qtfb.initialize-failed", "the server did not give a framebuffer")
	BadAnswer          = failure.Define("qtfb.bad-answer", "the server's first answer is not an initialize message")
	SharedMemoryFailed = failure.Define("qtfb.shared-memory-failed", "the framebuffer's memory could not be mapped")
	UpdateFailed       = failure.Define("qtfb.update-failed", "the server did not take a refresh")
)
