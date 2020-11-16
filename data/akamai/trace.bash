#!/usr/bin/env bash
#
# Akamai's backbone (2018)
#
# Source: https://www.linx.net/wp-content/uploads/LINX101-Akamai-ICN-ChristianKaufmann.pdf

set -e

# Automatically trace nodes
# ../../tracegeog trace-nodes \
#     -i orig.png \
#     -o xygraph.json \
#     -icon icon.png \
#     -node-color-accuracy 0.7 \
#     -max-node-count 28

# Automatically trace links (unused)
# ../../tracegeog trace-links \
#     -i orig.png \
#     -g xygraph.json \
#     -o xygraph-tracedlinks.json \
#     -line-color '#28328f' \
#     -line-color-accuracy 0.55 \
#     -line-dir-deg 45 \
#     -line-gap 5 \
#     -line-node-dist 10 \
#     -line-width 1

# Manually specify links

# Visualize links
# ../../tracegeog vis \
#     -i orig.png \
#     -g xygraph-manual-links.json \
#     -png xygraph.png \
#     -overlaypng overlayed.png

# Invert projected data to lat, lon pairs
../../tracegeog unproj \
    -g xygraph-manual-links.json \
    -extra-margin-left 760 \
    -prime-meridian-x 1290 \
    -equator-y 830 \
    -scale-y 0.93 \
    -o geograph-manual-links.json

# Replot on map to compare against source data
../../scripts/plotgeo.py \
    geograph-manual-links.json \
    redrawn.pdf
