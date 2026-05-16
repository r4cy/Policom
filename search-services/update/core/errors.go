package core

import "errors"

var ErrNotFound = errors.New("resource is not found")

// Domain error
var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrAlreadyRunning = errors.New("method update already running")

// Infrastructure error
var ErrResourceExhausted = errors.New("argument is too big")
var ErrDeadlineExceeded = errors.New("deadline server exceed")
var ErrUnknow = errors.New("server don't know this error")
