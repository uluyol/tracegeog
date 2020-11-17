#!/usr/bin/env bash
#
# Google's B4 backbone (2018)
#
# Source: https://www2.cs.duke.edu/courses/fall18/compsci514/readings/b4after-sigcomm18.pdf

set -e

# Automatically trace nodes
# ../../tracegeog trace-nodes \
#     -i orig.png \
#     -o xygraph.json \
#     -icon icon.png \
#     -node-color-accuracy 0.7 \
#     -max-node-count 18

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
# ../../tracegeog unproj \
#     -g xygraph-manual-links.json \
#     -extra-margin-left 330 \
#     -prime-meridian-x 1415 \
#     -equator-y 467 \
#     -o geograph-manual-links.json

# Replot on map to compare against source data
# ../../scripts/plotgeo.py \
#     geograph-manual-links.json \
#     redrawn.pdf

# Export to Repetita
../../tracegeog export-repetita \
    -g geograph-manual-links.json \
    -o Traced_GoogleB4.graph
