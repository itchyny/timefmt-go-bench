// Package main compares the performance of github.com/itchyny/timefmt-go
// against the other strftime and strptime libraries. Run ./bench.sh to collect
// the numbers and draw the graph. Note that bench.sh leaves some of the
// libraries out of the graph; see the comment on the keys in that script.
package main

import (
	"os"
	"testing"
	"time"

	karpeles "github.com/KarpelesLab/strftime"
	cactus "github.com/cactus/gostrftime"
	cockroachdb "github.com/cockroachdb/strtime"
	csotherden "github.com/csotherden/strftime"
	fastly "github.com/fastly/go-utils/strftime"
	going "github.com/going/strftime"
	timefmt "github.com/itchyny/timefmt-go"
	jehiah "github.com/jehiah/go-strftime"
	klauspost "github.com/klauspost/lctime"
	leekchan "github.com/leekchan/timeutil"
	lestrrat "github.com/lestrrat-go/strftime"
	ncruces "github.com/ncruces/go-strftime"
	pbnjay "github.com/pbnjay/strptime"
	twmb "github.com/twmb/go-strftime"
)

var (
	format         = os.Getenv("BENCH_FORMAT")
	standardFormat = os.Getenv("BENCH_STANDARD_FORMAT")
)

var (
	now   = time.Now().UTC()
	value = now.Format(standardFormat)
)

// TestFormat makes sure that the libraries yield the same result, so that the
// formatting benchmarks compare the same amount of work.
func TestFormat(t *testing.T) {
	tm := time.Date(2020, time.July, 27, 1, 2, 3, 4000, time.UTC)
	expected := tm.Format(standardFormat)
	for _, tc := range []struct {
		name string
		got  string
	}{
		{"timefmt.Format", timefmt.Format(tm, format)},
		{"timefmt.AppendFormat", string(timefmt.AppendFormat(nil, tm, format))},
		{"ncruces.Format", ncruces.Format(format, tm)},
		{"ncruces.AppendFormat", string(ncruces.AppendFormat(nil, format, tm))},
		{"twmb.AppendFormat", string(twmb.AppendFormat(nil, format, tm))},
		{"fastly.Strftime", fastly.Strftime(format, tm)},
		{"karpeles.EnFormat", karpeles.EnFormat(format, tm)},
		{"cactus.Format", cactus.Format(format, tm)},
		{"klauspost.Strftime", klauspost.Strftime(format, tm)},
		{"lestrrat.Format", must(lestrrat.Format(format, tm))},
		{"jehiah.Format", jehiah.Format(format, tm)},
		// {"tebeka.Format", must(tebeka.Format(format, tm))},
		// {"caiguanhao.Format", caiguanhao.Format(format, tm)},
		{"leekchan.Strftime", leekchan.Strftime(&tm, format)},
		// {"arnodel.Format", must(arnodel.Format(format, tm))},
		// {"awoodbeck.Format", awoodbeck.Format(&tm, format)},
		// {"csotherden.Format", csotherden.Format(format, tm)},
		// {"hhkbp2.Format", hhkbp2.Format(format, tm)},
		// {"belfinor.Format", belfinor.Format(format, tm)},
		{"going.Format", going.Format(format, tm)},
		// {"osteele.Strftime", must(osteele.Strftime(format, tm))},
		// {"billhathaway.New", tm.Format(must(billhathaway.New(format)))},
		{"cockroachdb.Strftime", must(cockroachdb.Strftime(tm, format))},
	} {
		if tc.got != expected {
			t.Errorf("%s(%q): got %q, expected %q", tc.name, format, tc.got, expected)
		}
	}
}

// TestParse makes sure that the libraries parse a formatted time back to the
// same time, so that the parsing benchmarks compare the same amount of work.
// The parsed times are compared after formatting them again, because the
// libraries disagree on the fields a format does not mention; this library
// defaults the year to 1900 as the C strptime does, while the libraries
// converting a format to a layout of the standard library default it to 0.
func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func() (time.Time, error)
	}{
		{"timefmt.Parse", func() (time.Time, error) { return timefmt.Parse(value, format) }},
		{"ncruces.Parse", func() (time.Time, error) { return ncruces.Parse(format, value) }},
		{"pbnjay.Parse", func() (time.Time, error) { return pbnjay.Parse(value, format) }},
		{"csotherden.Parse", func() (time.Time, error) { return csotherden.Parse(format, value) }},
		{"going.Parse", func() (time.Time, error) { return going.Parse(format, value) }},
		// {"cockroachdb.Strptime", func() (time.Time, error) { return cockroachdb.Strptime(value, format) }},
	} {
		switch tm, err := tc.fn(); {
		case err != nil:
			t.Errorf("%s(%q, %q): %s", tc.name, value, format, err)
		case tm.Format(standardFormat) != value:
			t.Errorf("%s(%q, %q): got %q, expected %q",
				tc.name, value, format, tm.Format(standardFormat), value)
		}
	}
}

func must(str string, err error) string {
	if err != nil {
		panic(err)
	}
	return str
}

// The formatting benchmarks of the libraries shown in the graph.

func BenchmarkFormatTimefmt(b *testing.B) {
	for b.Loop() {
		timefmt.Format(now, format)
	}
}

func BenchmarkFormatNcruces(b *testing.B) {
	for b.Loop() {
		ncruces.Format(format, now)
	}
}

func BenchmarkFormatTwmb(b *testing.B) {
	for b.Loop() {
		twmb.AppendFormat(make([]byte, 0, 64), format, now)
	}
}

func BenchmarkFormatFastly(b *testing.B) {
	for b.Loop() {
		fastly.Strftime(format, now)
	}
}

func BenchmarkFormatKarpeles(b *testing.B) {
	for b.Loop() {
		karpeles.EnFormat(format, now)
	}
}

func BenchmarkFormatCactus(b *testing.B) {
	for b.Loop() {
		cactus.Format(format, now)
	}
}

func BenchmarkFormatKlauspost(b *testing.B) {
	for b.Loop() {
		klauspost.Strftime(format, now)
	}
}

func BenchmarkFormatLestrrat(b *testing.B) {
	for b.Loop() {
		lestrrat.Format(format, now)
	}
}

// func BenchmarkFormatLestrratCached(b *testing.B) {
// 	f, err := lestrrat.New(format)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for b.Loop() {
// 		f.FormatString(now)
// 	}
// }

func BenchmarkFormatJehiah(b *testing.B) {
	for b.Loop() {
		jehiah.Format(format, now)
	}
}

func BenchmarkFormatStandard(b *testing.B) {
	for b.Loop() {
		now.Format(standardFormat)
	}
}

// The formatting benchmarks of the libraries left out of the graph.

// func BenchmarkFormatTebeka(b *testing.B) {
// 	for b.Loop() {
// 		tebeka.Format(format, now)
// 	}
// }

// func BenchmarkFormatCaiguanhao(b *testing.B) {
// 	for b.Loop() {
// 		caiguanhao.Format(format, now)
// 	}
// }

func BenchmarkFormatLeekchan(b *testing.B) {
	for b.Loop() {
		leekchan.Strftime(&now, format)
	}
}

// func BenchmarkFormatArnodel(b *testing.B) {
// 	for b.Loop() {
// 		arnodel.Format(format, now)
// 	}
// }

// func BenchmarkFormatAwoodbeck(b *testing.B) {
// 	for b.Loop() {
// 		awoodbeck.Format(&now, format)
// 	}
// }

// func BenchmarkFormatCsotherden(b *testing.B) {
// 	for b.Loop() {
// 		csotherden.Format(format, now)
// 	}
// }

// func BenchmarkFormatHhkbp2(b *testing.B) {
// 	for b.Loop() {
// 		hhkbp2.Format(format, now)
// 	}
// }

// func BenchmarkFormatBelfinor(b *testing.B) {
// 	for b.Loop() {
// 		belfinor.Format(format, now)
// 	}
// }

func BenchmarkFormatGoing(b *testing.B) {
	for b.Loop() {
		going.Format(format, now)
	}
}

// func BenchmarkFormatOsteele(b *testing.B) {
// 	for b.Loop() {
// 		osteele.Strftime(format, now)
// 	}
// }

// func BenchmarkFormatOsteeleCompiled(b *testing.B) {
// 	f, err := osteele.Compile(format)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for b.Loop() {
// 		f.Format(now)
// 	}
// }

// // Converts the format to a layout of the standard library once, so its cost
// // is that of the standard library, but it shows the conversion is possible.
// func BenchmarkFormatBillhathaway(b *testing.B) {
// 	layout, err := billhathaway.New(format)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for b.Loop() {
// 		now.Format(layout)
// 	}
// }

func BenchmarkFormatCockroachdb(b *testing.B) {
	for b.Loop() {
		cockroachdb.Strftime(now, format)
	}
}

// The parsing benchmarks. Unlike formatting, only a few libraries implement
// strptime at all, so all of them are shown in the graph.

func BenchmarkParseTimefmt(b *testing.B) {
	for b.Loop() {
		timefmt.Parse(value, format)
	}
}

func BenchmarkParseNcruces(b *testing.B) {
	for b.Loop() {
		ncruces.Parse(format, value)
	}
}

func BenchmarkParsePbnjay(b *testing.B) {
	for b.Loop() {
		pbnjay.Parse(value, format)
	}
}

func BenchmarkParseCsotherden(b *testing.B) {
	for b.Loop() {
		csotherden.Parse(format, value)
	}
}

func BenchmarkParseGoing(b *testing.B) {
	for b.Loop() {
		going.Parse(format, value)
	}
}

func BenchmarkParseCockroachdb(b *testing.B) {
	for b.Loop() {
		cockroachdb.Strptime(value, format)
	}
}

func BenchmarkParseStandard(b *testing.B) {
	for b.Loop() {
		time.Parse(standardFormat, value)
	}
}
