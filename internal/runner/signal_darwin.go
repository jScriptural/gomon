//go:build darwin

package signals

import "syscall"

const (
	SIGTERM = syscall.SIGTERM
	SIGKILL = syscall.SIGKILL
	SIGINT  = syscall.SIGINT
	SIGQUIT = syscall.SIGQUIT
)

var platformSignals = []syscall.Signal{
	SIGTERM,
	SIGKILL,
	SIGQUIT,
	SIGINT,
	syscall.SIGHUP,
	syscall.SIGILL,
	syscall.SIGTRAP,
	syscall.SIGABRT,
	syscall.SIGFPE,
	syscall.SIGBUS,
	syscall.SIGSEGV,
	syscall.SIGSYS,
	syscall.SIGPIPE,
	syscall.SIGALRM,
	syscall.SIGURG,
	syscall.SIGSTOP,
	syscall.SIGTSTP,
	syscall.SIGCONT,
	syscall.SIGCHLD,
	syscall.SIGTTIN,
	syscall.SIGTTOU,
	syscall.SIGIO,
	syscall.SIGXCPU,
	syscall.SIGXFSZ,
	syscall.SIGVTALRM,
	syscall.SIGPROF,
	syscall.SIGWINCH,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
	syscall.SIGINFO,
	syscall.SIGEMT,
}
