#!/usr/bin/env bash
#
# Cloudflare's backbone (2020)
#
# Source: https://blog.cloudflare.com/cloudflare-outage-on-july-17-2020/

set -e

# Automatically trace nodes
# ../../tracegeog trace-nodes \
#     -i orig2.png \
#     -o xygraph.json \
#     -icon icon2.png \
#     -node-color-accuracy 0.5 \
#     -max-node-count 21

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

# Manually adjust Node 8
# Manually specify links

# Visualize links
# ../../tracegeog vis \
#     -i orig.png \
#     -g xygraph-manualfix-and-links.json \
#     -png xygraph.png \
#     -overlaypng overlayed.png

# Invert projected data to lat, lon pairs
../../tracegeog unproj \
    -g xygraph-manualfix-and-links.json \
    -extra-margin-left 450 \
    -prime-meridian-x 770 \
    -equator-y 460 \
    -scale-y 0.98 \
    -o geograph-manual-links.json

# Replot on map to compare against source data
../../scripts/plotgeo.py \
    geograph-manual-links.json \
    redrawn.pdf
