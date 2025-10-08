//go:build !profile

package profiler

import "runtime"

// Stubbed no-op versions when the "profile" build tag is not set.

type Scope struct{}

var m runtime.MemStats

func Init(capacity int)                  {}
func Start(name string) Scope            { return Scope{} }
func (Scope) End()                       {}
func OpenProfilerGraph() (string, error) { return "", nil }

func MemoryUsage() uint64 {
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func MemoryAllocs() uint64 {
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func NumGoroutine() int {
	return runtime.NumGoroutine()
}

func NumCPU() int {
	return runtime.NumCPU()
}
