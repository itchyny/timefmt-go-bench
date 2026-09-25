# timefmt-go-bench
Benchmarks of [timefmt-go](https://github.com/itchyny/timefmt-go) against the
other time formatting (`strftime`) and parsing (`strptime`) libraries in Go.

```sh
./bench.sh [benchtime] [count]
```

The script runs the benchmarks of every library for each format, writes the
results to `bench-format.tsv` and `bench-parse.tsv`, and draws both graphs into
a single `bench.svg` with gnuplot, which `theme.awk` then makes follow the
light or dark theme of whatever renders it, and shrinks. The raw output of
`go test`, which includes the allocation counts, is kept in `bench.log`.

Each benchmark is run `count` times, one by default. The mean and the standard
deviation of the runs are both written to the files, but only the mean is
drawn; set `errorbars` in `bench.gnuplot` to draw the deviation as an error
bar, which needs no new run.

Requirements are Go and [gnuplot](http://www.gnuplot.info/). The benchmarked
copy of timefmt-go is the working tree of a sibling directory, so clone both
repositories next to each other:

```sh
git clone https://github.com/itchyny/timefmt-go
git clone https://github.com/itchyny/timefmt-go-bench
cd timefmt-go-bench && ./bench.sh
```

## Libraries
Every library is checked to yield the same result as the standard library
before it is timed, so that the benchmarks compare the same amount of work.
The graphs leave out the libraries that take an order of magnitude longer and
would flatten the other bars, and the forks of a library already drawn. Each
line of `format_series` and `parse_series` in `bench.sh` is the suffix of a
benchmark function and the name to show in the legend, separated by a tab;
move a line in or out of a list to change what is drawn. The benchmarks of the
libraries no longer drawn are commented out in `bench_test.go`, while their
results stay checked by `TestFormat`.

`imperfectgo/go-strftime` reaches into the internals of the time package with
`go:linkname` and no longer builds with Go 1.23 or later. It is benchmarked
with an older toolchain in `imperfectgo`, and scaled by the ratio of the
standard library benchmarks of the two toolchains, so its bar is an estimate.
Install one with

```sh
go install golang.org/dl/go1.22.12@latest && go1.22.12 download
```

or point `BENCH_OLD_GO` at a go binary; without one its bar is left out.

## License
This software is released under the MIT License, see LICENSE.
