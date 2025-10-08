//go:build !profile

package profiler

// Stubbed no-op versions when the "profile" build tag is not set.

func Init(capacity int)                  {}
func Start(name string) func()           { return func() {} }
func OpenProfilerGraph() (string, error) { return "", nil }
func MemoryUsage() uint64                { return 0 }
func MemoryAllocs() uint64               { return 0 }
func NumGoroutine() int                  { return 0 }
func NumCPU() int                        { return 0 }
