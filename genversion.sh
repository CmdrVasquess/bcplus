#!/bin/sh
. ./VERSION
echo "package main" > version.go
echo "" >> version.go
echo "const (" >> version.go
echo "	BCpMajor   uint16 = "$major >> version.go
echo "	BCpMinor   uint16 = "$minor >> version.go
echo "	BCpBugfix  uint16 = "$bugfix >> version.go
echo "	BCpDate    string = \""`date`"\"" >> version.go
echo "	BCpQuality string = \""$quality"\"" >> version.go
echo ")" >> version.go
