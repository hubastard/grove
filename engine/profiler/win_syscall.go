//go:build profile && windows

package profiler

import "syscall"

type syscallSysProcAttr = syscall.SysProcAttr

func hideWindowAttr() any {
	return &syscall.SysProcAttr{HideWindow: true}
}
