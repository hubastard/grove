package scratch

import (
	"strconv"
	"unsafe"
)

// Package-level reusable buffer (single-threaded usage).
// Initialize once with Init(capacity). Reset() every frame.
// If you ever exceed capacity, call GrowTo(...) at init/loading time (not per-frame).
var buf []byte

// Init sets up the global scratch buffer. Call once at startup.
// Example: scratch.Init(4 * 1024)
func Init(capacity int) {
	if capacity <= 0 {
		capacity = 1024
	}
	buf = make([]byte, 0, capacity)
}

// Reset clears the buffer length without freeing memory.
// Call this ONCE per frame (e.g., right before drawing UI).
func Reset() { buf = buf[:0] }

// Cap returns the current capacity. Useful for tuning.
func Cap() int { return cap(buf) }

// Len returns the current length.
func Len() int { return len(buf) }

// GrowTo increases capacity (and copies current contents) if needed.
// Prefer calling this during load or on rare resize events, not every frame.
func GrowTo(minCapacity int) {
	if minCapacity <= cap(buf) {
		return
	}
	nb := make([]byte, len(buf), minCapacity)
	copy(nb, buf)
	buf = nb
}

// Ensure ensures there is room for at least n more bytes (amortized, not every call).
// This may allocate ONCE if capacity is insufficient.
func Ensure(n int) {
	if len(buf)+n > cap(buf) {
		newCap := cap(buf) * 2
		if newCap < len(buf)+n {
			newCap = len(buf) + n
		}
		GrowTo(newCap)
	}
}

// ----- Chainable builder over the global buffer -----

type Builder struct{}

// F returns a builder bound to the global buffer.
func F() Builder { return Builder{} }

// Mark returns a bookmark to later slice the output.
// Example: m := scratch.Mark(); ...; s := scratch.BytesFrom(m)
func Mark() int { return len(buf) }

// BytesFrom returns the bytes produced since mark.
func BytesFrom(mark int) []byte { return buf[mark:] }

// StringFrom SAFE copy: converts the range since mark to string (allocates 1 string).
func StringFrom(mark int) string { return string(buf[mark:]) }

// StringViewFrom ZERO-COPY (unsafe): string view into the buffer since mark.
// DO NOT modify the buffer while using this string. Valid only until next Reset()/append.
func StringViewFrom(mark int) string {
	b := buf[mark:]
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// Bytes returns the whole buffer.
func Bytes() []byte { return buf }

// String SAFE copy of whole buffer (allocates 1 string).
func String() string { return string(buf) }

// StringView ZERO-COPY string referencing the whole buffer.
// Lifetime: only until next append/Reset. Use cautiously.
func StringView() string {
	if len(buf) == 0 {
		return ""
	}
	return unsafe.String(&buf[0], len(buf))
}

// ----- Append primitives (chainable) -----

func (Builder) B(b []byte) Builder {
	buf = append(buf, b...)
	return Builder{}
}

func (Builder) S(s string) Builder {
	buf = append(buf, s...)
	return Builder{}
}

func (Builder) C(c byte) Builder {
	buf = append(buf, c)
	return Builder{}
}

func (Builder) R(r rune) Builder {
	// ASCII fast-path; for full UTF-8, use utf8.AppendRune
	if r < 128 {
		buf = append(buf, byte(r))
		return Builder{}
	}
	var tmp [4]byte
	n := utf8EncodeRune(tmp[:], r)
	buf = append(buf, tmp[:n]...)
	return Builder{}
}

// I appends a base-10 integer.
func (Builder) I(v int) Builder {
	buf = strconv.AppendInt(buf, int64(v), 10)
	return Builder{}
}

// U appends an unsigned base-10 integer.
func (Builder) U(v uint) Builder {
	buf = strconv.AppendUint(buf, uint64(v), 10)
	return Builder{}
}

// F64 appends a float with given precision (digits after decimal).
// Example: F64(3.14159, 2) -> "3.14"
func (Builder) F64(v float64, prec int) Builder {
	buf = strconv.AppendFloat(buf, v, 'f', prec, 64)
	return Builder{}
}

// Bool appends "true"/"false".
func (Builder) Bool(v bool) Builder {
	buf = strconv.AppendBool(buf, v)
	return Builder{}
}

// Hex appends an integer in hexadecimal without "0x".
func (Builder) Hex(u uint64) Builder {
	buf = strconv.AppendUint(buf, u, 16)
	return Builder{}
}

// Pad appends n copies of byte c.
func (Builder) Pad(n int, c byte) Builder {
	if n <= 0 {
		return Builder{}
	}
	Ensure(n)
	for i := 0; i < n; i++ {
		buf = append(buf, c)
	}
	return Builder{}
}

// ----- Minimal % formatter (no allocations, no reflection) -----
// Supports a tiny subset: %s %d %u %f (with .prec) %%
// Format: "HP %d/%d  RTT %.2f ms"
//
// Usage:
//
//	scratch.Reset()
//	scratch.Printf("HP %d/%d RTT %.2f ms", hp, max, rtt)
//
// NOTE: This avoids fmt’s heap use, but still boxes arguments into interfaces
// at the callsite for the variadic args slice. In practice, Go often keeps
// small arg slices on stack; if you want *strict* zero-alloc, prefer the chainable API.
func Sprintf(format string, args ...any) string {
	var ai int
	mark := len(buf)
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch != '%' {
			buf = append(buf, ch)
			continue
		}
		// "%%" escape
		if i+1 < len(format) && format[i+1] == '%' {
			buf = append(buf, '%')
			i++
			continue
		}
		// parse verb (+ optional .precision for %f)
		i++
		prec := -1
		if i < len(format) && format[i] == '.' {
			// .<digits>
			i++
			start := i
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				i++
			}
			prec = parseUint(format[start:i])
		}
		if i >= len(format) || ai >= len(args) {
			break
		}
		switch format[i] {
		case 's':
			buf = append(buf, toString(args[ai])...)
		case 'd':
			buf = strconv.AppendInt(buf, toInt64(args[ai]), 10)
		case 'u':
			buf = strconv.AppendUint(buf, toUint64(args[ai]), 10)
		case 'f':
			p := 3
			if prec >= 0 {
				p = prec
			}
			buf = strconv.AppendFloat(buf, toFloat64(args[ai]), 'f', p, 64)
		default:
			// unknown verb, write literally
			buf = append(buf, '%', format[i])
		}
		ai++
	}
	return string(buf[mark:])
}

// ----- tiny helpers (no alloc) -----

func parseUint(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x) // allocs a string copy
	default:
		// fall back minimally
		return "<unsupported>"
	}
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	default:
		return 0
	}
}

func toUint64(v any) uint64 {
	switch x := v.(type) {
	case uint:
		return uint64(x)
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	case uint64:
		return x
	case int:
		return uint64(x)
	case int8:
		return uint64(x)
	case int16:
		return uint64(x)
	case int32:
		return uint64(x)
	case int64:
		return uint64(x)
	default:
		return 0
	}
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	default:
		return 0
	}
}

// Minimal UTF-8 encoding to avoid importing utf8 package.
func utf8EncodeRune(dst []byte, r rune) int {
	switch {
	case r < 0x80:
		dst[0] = byte(r)
		return 1
	case r < 0x800:
		dst[0] = 0xC0 | byte(r>>6)
		dst[1] = 0x80 | byte(r&0x3F)
		return 2
	case r < 0x10000:
		dst[0] = 0xE0 | byte(r>>12)
		dst[1] = 0x80 | byte((r>>6)&0x3F)
		dst[2] = 0x80 | byte(r&0x3F)
		return 3
	default:
		dst[0] = 0xF0 | byte(r>>18)
		dst[1] = 0x80 | byte((r>>12)&0x3F)
		dst[2] = 0x80 | byte((r>>6)&0x3F)
		dst[3] = 0x80 | byte(r&0x3F)
		return 4
	}
}
