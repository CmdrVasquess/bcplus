#!/bin/bash
sed 's/^.*"event":"\([^"]*\)".*$/\1/' $1/Journal*.log | sort -u >> events.txt

evts=$(cat events.txt)
for e in $evts; do
	echo "filtering for event: "$e
	fgrep -h '"event":"'$e'"' $1/Journal*.log > $e.events
	genson -i 3 $e.events > $e.jschema
done
