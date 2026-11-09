package pen

import "github.com/akrck02/go-remarkable/failure"

// DeviceUnreadable says the pen's input device could not be opened.
var DeviceUnreadable = failure.Define("pen.device-unreadable", "the pen's input device could not be opened")
