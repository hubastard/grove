//go:build profile

package profiler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// -------- public API --------

// Init must be called once (e.g., on app start) with a capacity (#spans).
// Example: profiler.Init(1 << 20) // ~1M scope samples
func Init(capacity int) {
	if capacity <= 0 {
		capacity = 1 << 20
	}
	evrb.init(capacity)
}

// Start begins a scope and returns an end func to be deferred.
func Start(name string) func() {
	if !evrb.ready.Load() {
		return func() {}
	}
	fid := intern(name)
	// Emit OPEN now
	now := time.Now().UnixNano()
	evrb.push(evEntry{AtNS: now, FrameID: fid, Open: true})
	return func() {
		// Emit CLOSE now
		end := time.Now().UnixNano()
		// Guarantee end >= start order even if clock equal at µs granularity:
		if end < now {
			end = now
		}
		evrb.push(evEntry{AtNS: end, FrameID: fid, Open: false})
	}
}

// OpenProfilerGraph writes the stats into a temporary speedscope file and open it.
func OpenProfilerGraph() (string, error) {
	evs := evrb.snapshot()
	if len(evs) == 0 {
		return "", fmt.Errorf("profiler: no events to dump")
	}

	profilePath := filepath.Join(os.TempDir(), "grave.profile.speedscope.json")
	if err := dumpSpeedscopeEvents(evs, profilePath); err != nil {
		return "", err
	}

	cmd := exec.Command("speedscope", profilePath)
	// On Windows, hide console window:
	if runtime.GOOS == "windows" {
		if spa, ok := hideWindowAttr().(*syscall.SysProcAttr); ok {
			cmd.SysProcAttr = spa
		}
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error launching speedscope: %v\n", err)
	}

	return profilePath, nil
}

func MemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func MemoryAllocs() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Mallocs
}

func NumGoroutine() int {
	return runtime.NumGoroutine()
}

func NumCPU() int {
	return runtime.NumCPU()
}

// ---------- event ring ----------

type evEntry struct {
	AtNS    int64
	FrameID int
	Open    bool
}

type evRing struct {
	ready atomic.Bool
	cap   uint64
	write atomic.Uint64
	evs   []evEntry
}

func (r *evRing) init(capacity int) {
	r.cap = uint64(capacity)
	r.evs = make([]evEntry, r.cap)
	r.write.Store(0)
	r.ready.Store(true)
}

func (r *evRing) push(e evEntry) {
	i := r.write.Add(1) - 1
	r.evs[i%r.cap] = e
}

// snapshot preserves **write order** — no sorting later.
func (r *evRing) snapshot() []evEntry {
	n := r.write.Load()
	if n == 0 {
		return nil
	}
	start := uint64(0)
	if n > r.cap {
		start = n - r.cap
	}
	size := n - start
	out := make([]evEntry, 0, size)
	for k := start; k < n; k++ {
		out = append(out, r.evs[k%r.cap])
	}
	return out
}

var evrb evRing

// ---------- string interner ----------

var (
	muFrames sync.Mutex
	frames   []string
	index    = map[string]int{}
)

func intern(name string) int {
	muFrames.Lock()
	defer muFrames.Unlock()
	if id, ok := index[name]; ok {
		return id
	}
	id := len(frames)
	index[name] = id
	frames = append(frames, name)
	return id
}

// ---------- platform helper (windows optional) ----------

// Use a private alias to avoid importing syscall on non-windows files.
type syscallSysProcAttr = struct{ _ uintptr } // placeholder; real type set in windows file

func hideWindowAttr() any { return nil }

// -------- speedscope writer (evented) --------

// ---------- speedscope dump from EVENTS ----------
type ssFile struct {
	Schema             string      `json:"$schema"`
	Shared             ssShared    `json:"shared"`
	Profiles           []ssProfile `json:"profiles"`
	ActiveProfileIndex int         `json:"activeProfileIndex,omitempty"`
	Exporter           string      `json:"exporter,omitempty"`
	Name               string      `json:"name,omitempty"`
}
type ssShared struct {
	Frames []ssFrame `json:"frames"`
}
type ssFrame struct {
	Name string `json:"name"`
}

type ssProfile struct {
	Type       string    `json:"type"` // "evented"
	Name       string    `json:"name"`
	Unit       string    `json:"unit"` // "microseconds"
	StartValue int64     `json:"startValue"`
	EndValue   int64     `json:"endValue"`
	Events     []ssEvent `json:"events"`
}

// Your speedscope wants frame on both O and C
type ssEvent struct {
	Type  string `json:"type"`  // "O" or "C"
	At    int64  `json:"at"`    // µs since first event
	Frame int    `json:"frame"` // frame index
}

func dumpSpeedscopeEvents(evs []evEntry, path string) error {
	// snapshot frames
	muFrames.Lock()
	fs := make([]ssFrame, len(frames))
	for i, name := range frames {
		fs[i] = ssFrame{Name: name}
	}
	muFrames.Unlock()

	if len(evs) == 0 {
		return fmt.Errorf("no events")
	}

	base := evs[0].AtNS
	endUS := int64(0)

	// stream in write order with small stack filter
	out := make([]ssEvent, 0, len(evs)+16)
	stack := make([]int, 0, 64) // frame IDs
	lastUS := int64(-1)

	for _, e := range evs {
		atUS := (e.AtNS - base) / 1000
		if atUS < lastUS {
			atUS = lastUS // keep µs monotonic
		}

		if e.Open {
			out = append(out, ssEvent{Type: "O", At: atUS, Frame: e.FrameID})
			stack = append(stack, e.FrameID)
		} else {
			// unmatched/mismatched close? skip
			if len(stack) == 0 || stack[len(stack)-1] != e.FrameID {
				continue
			}
			stack = stack[:len(stack)-1]
			out = append(out, ssEvent{Type: "C", At: atUS, Frame: e.FrameID})
		}

		lastUS = atUS
		if atUS > endUS {
			endUS = atUS
		}
	}

	// --- AUTO-CLOSE ANY REMAINING OPENS (LIFO) ---
	// If we ended mid-frame, speedscope expects balanced events.
	// Close everything still on the stack at the final timestamp.
	if len(stack) > 0 {
		atUS := lastUS // same timestamp is OK; we already maintain push/pop order
		for i := len(stack) - 1; i >= 0; i-- {
			out = append(out, ssEvent{Type: "C", At: atUS, Frame: stack[i]})
		}
		// stack cleared implicitly; endUS unchanged (atUS == lastUS)
	}

	if len(out) == 0 {
		return fmt.Errorf("no usable events after filtering")
	}

	doc := ssFile{
		Schema: "https://www.speedscope.app/file-format-schema.json",
		Shared: ssShared{Frames: fs},
		Profiles: []ssProfile{{
			Type:       "evented",
			Name:       "Go Engine (evented)",
			Unit:       "microseconds",
			StartValue: 0,
			EndValue:   endUS,
			Events:     out,
		}},
		ActiveProfileIndex: 0,
		Exporter:           "goengine-profiler",
		Name:               "Go Engine capture",
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&doc); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
