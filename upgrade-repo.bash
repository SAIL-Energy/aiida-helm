#!/usr/bin/bash

make package TRACK=beta CHART_PATH="./chart"
make package TRACK=beta CHART_PATH="./installer/chart"

make package TRACK=stable CHART_PATH="./chart"
make package TRACK=stable CHART_PATH="./installer/chart"

make index TRACK=beta
make index TRACK=stable