#!/bin/sh
. ./VERSION
echo "package main" > version.go
echo "" >> version.go
echo "const (" >> version.go
echo "	BCpMajor  uint16 = 0" >> version.go
echo "	BCpMinor  uint16 = 4" >> version.go
echo "	BCpBugfix uint16 = 4" >> version.go
echo "	BCpDate   string = \""`date`"\"" >> version.go
echo ")" >> version.go
