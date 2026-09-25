# Themes and shrinks the SVG that gnuplot writes.
#
# The colours gnuplot writes are left in place as presentation attributes, so
# that a renderer which ignores stylesheets still draws the light theme; a
# stylesheet overriding them for a dark colour scheme is added on top. Only
# the canvas, the text, the axes, the grid and the grey bar of the standard
# library are themed, as the colours of the other bars read on either
# background. The colours are first written in a canonical form, since gnuplot
# pads them with spaces that an attribute selector would have to match.
#
# What is left of the file is then shrunk: the marker shapes that nothing
# refers to, the indentation and a digit of every coordinate come out, and the
# font and the long group attributes that gnuplot repeats on every element are
# hoisted into the stylesheet.
#
#   Usage: awk -f theme.awk bench.svg >themed.svg

BEGIN {
    dark["fill"  , "#ffffff"]          = "#0d1117" # canvas
    dark["fill"  , "rgb(32,32,32)"]    = "#e6edf3" # tic labels and legend
    dark["stroke", "rgb(32,32,32)"]    = "#e6edf3" # error bars
    dark["fill"  , "rgb(96,96,96)"]    = "#9198a1" # footnote
    dark["fill"  , "rgb(95,99,104)"]   = "#9198a1" # the standard library bar
    dark["stroke", "rgb(128,128,128)"] = "#6e7681" # axes
    dark["stroke", "rgb(204,204,204)"] = "#30363d" # major grid lines
    dark["stroke", "rgb(232,232,232)"] = "#21262d" # minor grid lines
    dark["color" , "black"]            = "#e6edf3" # inherited by currentColor

    # The attributes gnuplot repeats on the group wrapping every element,
    # which cost more than a third of the file until hoisted into a class.
    group = "<g fill=\"none\" color=\"black\" stroke=\"currentColor\"" \
            " stroke-width=\"1\" stroke-linecap=\"butt\"" \
            " stroke-linejoin=\"miter\">"
}

# Strips the padding gnuplot writes inside every rgb() colour.
function canonical(s,   out, token) {
    out = ""
    while (match(s, /rgb\([ 0-9]+,[ 0-9]+,[ 0-9]+\)/)) {
        token = substr(s, RSTART, RLENGTH)
        gsub(/ /, "", token)
        out = out substr(s, 1, RSTART - 1) token
        s = substr(s, RSTART + RLENGTH)
    }
    return out s
}

# Rounds every number to a single decimal, which is far below a pixel.
function round(s,   out, token) {
    out = ""
    while (match(s, /[0-9]+\.[0-9][0-9]+/)) {
        token = sprintf("%.1f", substr(s, RSTART, RLENGTH) + 0)
        sub(/\.0$/, "", token)
        out = out substr(s, 1, RSTART - 1) token
        s = substr(s, RSTART + RLENGTH)
    }
    return out s
}

# Only the first marker shape is ever referred to, by the invisible plot that
# places the labels of a group, so the rest of the definitions are dropped.
/<defs>/  { indefs = 1 }
/<\/defs>/ { indefs = 0; print; next }
indefs && !/<defs>/ && !/gpPt0/ { next }

/<title>Gnuplot<\/title>/ || /<desc>/ { next }

{
    $0 = canonical($0)
    # Never round inside a label: it would rewrite a version number such as
    # 1.23 in the footnote. Only markup lines carry coordinates worth losing
    # a digit from.
    if ($0 !~ /<text/) {
        $0 = round($0)
    }
    gsub(/#[fF][fF][fF][fF][fF][fF]/, "#ffffff")
    gsub(/fill="black"/, "fill=\"rgb(32,32,32)\"")
    gsub(/stroke="black"/, "stroke=\"rgb(32,32,32)\"")
    gsub(/ ?font-family="monospace"/, "")  # hoisted into the stylesheet
    gsub(group, "<g class=\"e\">")
}

/^<g id="gnuplot_canvas">/ {
    print "<style>"
    print "text,tspan { font-family: monospace; }"
    print ".e { fill: none; color: black; stroke: currentColor;"
    print "     stroke-width: 1; stroke-linecap: butt; stroke-linejoin: miter; }"
    print "@media (prefers-color-scheme: dark) {"
    print "  .e { color: #e6edf3 !important; }"
    for (key in dark) {
        split(key, part, SUBSEP)
        printf "  [%s=\"%s\"] { %s: %s !important; }\n",
               part[1], part[2], part[1], dark[key]
    }
    print "}"
    print "</style>"
}

NF { print }
