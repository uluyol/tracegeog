#!/usr/bin/env bash
#
# Cloudflare's backbone (2020)
#
# Source: https://aws.amazon.com/about-aws/global-infrastructure/global_network/
# (https://d1.awsstatic.com/diagrams/product-page-diagrams/5003_Global%20Infrastructure%20Map_update.eb8a0e26869e6f7761a723d9d93808ce756c36ff.png)

set -e

# Automatically trace nodes
# ../../tracegeog trace-nodes \
#     -i orig.png \
#     -o xygraph.json \
#     -icon icon.png \
#     -node-color-accuracy 0.5 \
#     -max-node-count 25 \
#     -transit-icon icon2.png \
#     -transit-color-accuracy 0.8 \
#     -max-transit-count 14

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
#     -i orig-no-transit.png \
#     -g xygraph-manualfix-and-links.json \
#     -png xygraph.png \
#     -overlaypng overlayed.png

# Invert projected data to lat, lon pairs
# ../../tracegeog unproj \
#     -g xygraph-manualfix-and-links.json \
#     -extra-margin-left -170 \
#     -extra-margin-right 100 \
#     -prime-meridian-x 745 \
#     -equator-y 537 \
#     -scale-y 0.85 \
#     -o geograph-manual-links.json

# Replot on map to compare against source data
# ../../scripts/plotgeo.py \
#     geograph-manual-links.json \
#     redrawn.pdf

# Export to Repetita
../../tracegeog export-repetita \
    -g geograph-manual-links.json \
    -o Traced_AWS.graph
