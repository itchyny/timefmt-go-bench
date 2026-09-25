#!/bin/bash

# Runs the formatting and parsing benchmarks against the other strftime and
# strptime libraries, and draws the graph shown in the README. Requires
# gnuplot to plot the results.
#
#   Usage: ./bench.sh [benchtime] [count]
#
# Every benchmark is run count times, and a bar is the mean of the runs. The
# standard deviation is written to the data files as well, so that setting
# errorbars in bench.gnuplot draws it without running the benchmarks again.

set -eu
set -o pipefail # so that a failing go test aborts the script

cd "$(dirname "$0")"

benchtime=${1-1s}
count=${2-1}

# Formats to benchmark, each line being a strftime format and the equivalent
# layout of the standard library, separated by a tab.
formats="%H:%M:%S	15:04:05
%Y-%m-%d	2006-01-02
%Y-%m-%dT%H:%M:%S%z	2006-01-02T15:04:05-0700
%a %b %d %H:%M:%S %Z %Y	Mon Jan 02 15:04:05 MST 2006"

# The libraries to draw, in the order the bars are laid out. Each line is the
# suffix of a benchmark function and the name to show in the legend, separated
# by a tab, so that a bar cannot be labelled with the name of another library.
# Benchmarks left out of a list are still recorded in bench.log, but not drawn
# in a graph, either because they take an order of magnitude longer and would
# flatten the other bars, or because they are a fork of a library already
# drawn. Move a line in or out of a list to change what is drawn. The standard
# library comes first in a list, and is drawn in grey, as the baseline that
# the other libraries are read against.
format_series="FormatStandard	standard library
FormatTimefmt	{/:Bold itchyny/timefmt-go}
FormatTwmb	twmb/go-strftime
FormatFastly	fastly/go-utils/strftime
FormatNcruces	ncruces/go-strftime
FormatKarpeles	KarpelesLab/strftime
FormatLestrrat	lestrrat-go/strftime
FormatJehiah	jehiah/go-strftime
FormatKlauspost	klauspost/lctime
FormatCactus	cactus/gostrftime
FormatGoing	going/strftime
FormatLeekchan	leekchan/timeutil
FormatCockroachdb	cockroachdb/strtime"

# Only a few libraries implement strptime at all, so all of them are drawn.
parse_series="ParseStandard	standard library
ParseTimefmt	{/:Bold itchyny/timefmt-go}
ParseNcruces	ncruces/go-strftime
ParseCsotherden	csotherden/strftime
ParseCockroachdb	cockroachdb/strtime
ParseGoing	going/strftime
ParsePbnjay	pbnjay/strptime"

# An older Go toolchain to benchmark imperfectgo/go-strftime with, which reaches
# into the unexported internals of the time package with go:linkname and does
# not link with Go 1.23 or later. Its results cannot be compared with the rest
# directly, so they are scaled by the ratio of the standard library benchmarks
# of the two toolchains, and labelled as the estimates they are. Install one by
#
#   go install golang.org/dl/go1.22.12@latest && go1.22.12 download
#
# or point BENCH_OLD_GO at a go binary; without one the bars are left out.
old_go=${BENCH_OLD_GO-go1.22.12}
command -v "$old_go" >/dev/null 2>&1 || old_go=
if [ -n "$old_go" ]; then
  format_series=$(sed '/twmb/a \
FormatImperfectGo	imperfectgo/go-strftime^{/*0.7 *}
' <<<"$format_series")
fi

results=$(mktemp -d)
trap 'rm -rf "$results"' EXIT

: >bench.log

# Run every benchmark once per format, appending the raw output to bench.log
# and keeping the "BenchmarkName ns/op" pairs of each format in its own file.
groups=0
while IFS=$'\t' read -r format standard; do
  echo "==> $format ($standard)" >&2
  groups=$((groups + 1))
  printf '%s\t%s' "$format" "$standard" >"$results/$groups.group"
  BENCH_FORMAT="$format" BENCH_STANDARD_FORMAT="$standard" \
    go test -bench . -benchmem -benchtime "$benchtime" -count "$count" |
    tee -a bench.log |
    awk '/ ns\/op/ { sub(/-[0-9]+$/, "", $1); print $1, $3 }' >"$results/$groups.ns"

  [ -n "$old_go" ] || continue

  echo "=== imperfectgo, $(GOTOOLCHAIN=local "$old_go" version)" >>bench.log
  (cd imperfectgo &&
    BENCH_FORMAT="$format" BENCH_STANDARD_FORMAT="$standard" GOTOOLCHAIN=local \
      "$old_go" test -bench . -benchmem -benchtime "$benchtime" -count "$count") |
    tee -a bench.log |
    awk '/ ns\/op/ { sub(/-[0-9]+$/, "", $1); print $1, $3 }' >"$results/$groups.old"

  # Scale the results of the older toolchain by the ratio of the standard
  # library benchmarks, the only code both toolchains have in common. Every
  # sample is scaled, so that the spread reaches the error bar as well.
  awk -v ns="$results/$groups.ns" '
    $1 == "BenchmarkFormatStandard" { oldsum += $2; oldn++ }
    { sample[NR] = $0 }
    END {
      while ((getline line < ns) > 0) {
        split(line, field, " ")
        if (field[1] == "BenchmarkFormatStandard") { newsum += field[2]; newn++ }
      }
      ratio = (newsum / newn) / (oldsum / oldn)
      for (i = 1; i <= NR; i++) {
        split(sample[i], field, " ")
        if (field[1] ~ /ImperfectGo/)
          printf "%s %.4g\n", field[1], field[2] * ratio
      }
    }' "$results/$groups.old" >>"$results/$groups.ns"
done <<<"$formats"

# Writes the results of the listed benchmarks to a tab separated file, which
# bench.gnuplot draws as one of the graphs.
write_tsv() { # write_tsv <file> <series>
  {
    printf 'format\tstandard'
    cut -f2 <<<"$2" | while IFS= read -r label; do
      printf '\t%s\t%s error' "$label" "$label"
    done
    printf '\n'
    for group in $(seq 1 "$groups"); do
      cat "$results/$group.group"
      awk -v keys="$(cut -f1 <<<"$2" | tr '\n' ' ')" '
        { runs[$1]++; sum[$1] += $2; square[$1] += $2 * $2 }
        END {
          for (i = 1; i <= split(keys, key, " "); i++) {
            name = "Benchmark" key[i]
            if (!(name in runs)) {
              printf "bench.sh: %s not found\n", name >"/dev/stderr"
              exit 1
            }
            mean = sum[name] / runs[name]
            variance = runs[name] > 1 \
              ? (square[name] - runs[name] * mean * mean) / (runs[name] - 1) : 0
            printf "\t%.4g\t%.3g", mean, (variance > 0 ? sqrt(variance) : 0)
          }
        }' "$results/$group.ns"
      printf '\n'
    done
  } >"$1"
  echo "==> wrote $1" >&2
}

write_tsv bench-format.tsv "$format_series"
write_tsv bench-parse.tsv "$parse_series"

gnuplot bench.gnuplot
awk -f theme.awk bench.svg >bench.svg.tmp && mv bench.svg.tmp bench.svg

echo "==> wrote bench.svg (raw benchmark output in bench.log)" >&2
