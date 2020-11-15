#!/usr/bin/env python3

import json

import cartopy.crs as ccrs
import matplotlib.pyplot as plt

def PlotGraph(graph, output_path):
    ax = plt.axes(projection=ccrs.PlateCarree())
    ax.coastlines()

    for l in graph["Links"]:
        src = graph["Nodes"][l["Src"]]
        dst = graph["Nodes"][l["Dst"]]
        plt.plot(
                [src["Lon"], dst["Lon"]],
                [src["Lat"], dst["Lat"]],
                linewidth=0.5,
                color='orange',
                transform=ccrs.Geodetic())

    for i in range(len(graph["Nodes"])):
        n = graph["Nodes"][i]
        lat = n["Lat"]
        lon = n["Lon"]
        plt.plot(lon, lat, color="blue", markersize=3, marker='o')
        #plt.text(lon - 3, lat - 12, str(i),
        #        horizontalalignment='right',
        #        transform=ccrs.Geodetic())

    plt.savefig(output_path)

def main():
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("input_json")
    parser.add_argument("output_path")

    args = parser.parse_args()
    with open(args.input_json) as fin:
        graph = json.load(fin)

    PlotGraph(graph, args.output_path)

if __name__ == "__main__":
    main()
