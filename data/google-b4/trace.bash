#!/usr/bin/env bash

set -e

# ../../tracegeog trace-nodes \
#     -i orig.png \
#     -o xygraph.json \
#     -icon icon.png \
#     -node-color-accuracy 0.7 \
#     -max-node-count 18

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

# ../../tracegeog vis \
#     -i orig.png \
#     -g xygraph-manual-links.json \
#     -png xygraph.png \
#     -overlaypng overlayed.png

../../tracegeog unproj \
    -g xygraph-manual-links.json \
    -extra-margin-left 330 \
    -prime-meridian-x 1415 \
    -equator-y 467 \
    -o geograph-manual-links.json

../../scripts/plotgeo.py \
    geograph-manual-links.json \
    redrawn.pdf
