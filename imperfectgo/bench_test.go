// Package main benchmarks imperfectgo/go-strftime, which reaches into the
// unexported internals of the time package with go:linkname and therefore
// does not link with Go 1.23 or later, where those internals were rewritten.
// It is benchmarked with an older toolchain, along with the baselines that
// bench.sh scales its results by; see the comment on old_go in that script.
package main

import (
	"os"
	"testing"
	"time"

	imperfectgo "github.com/imperfectgo/go-strftime"
)

var (
	format         = os.Getenv("BENCH_FORMAT")
	standardFormat = os.Getenv("BENCH_STANDARD_FORMAT")
)

var now = time.Now().UTC()

func TestFormat(t *testing.T) {
	tm := time.Date(2020, time.July, 27, 1, 2, 3, 4000, time.UTC)
	expected := tm.Format(standardFormat)
	for _, tc := range []struct {
		name string
		got  string
	}{
		{"imperfectgo.Format", imperfectgo.Format(tm, format)},
		{"imperfectgo.AppendFormat", string(imperfectgo.AppendFormat(nil, tm, format))},
	} {
		if tc.got != expected {
			t.Errorf("%s(%q): got %q, expected %q", tc.name, format, tc.got, expected)
		}
	}
}

func BenchmarkFormatImperfectGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		imperfectgo.Format(now, format)
	}
}

func BenchmarkFormatImperfectGoNoAlloc(b *testing.B) {
	buf := make([]byte, 0, 128)
	for i := 0; i < b.N; i++ {
		imperfectgo.AppendFormat(buf[:0], now, format)
	}
}

// The baseline, benchmarked here as well so that bench.sh can scale the
// results above to the toolchain the other libraries are benchmarked with.
// The standard library is the only code this old toolchain and the current
// one have in common; the library itself cannot be built by both, since its
// go directive is newer than this toolchain.

func BenchmarkFormatStandard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now.Format(standardFormat)
	}
}
