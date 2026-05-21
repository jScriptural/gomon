//go:build !darwin && !linux

package runner

import "syscall"

const (
	SIGQUIT = syscall.SIGINT
	SIGINT  = syscall.SIGINT
	SIGTERM = syscall.SIGINT
	SIGKILL = syscall.SIGKILL
)

// signals guaranteed by golang
var platformSignals = []syscall.Signal{
	SIGKILL,
	SIGQUIT,
}
