#!/bin/sh
# sed 's/^[ \t]*<!--/<!--/' travel.html | less
for f in $*; do
	tmp=$f~
	sed 's/^[ \t]*<!--/<!--/' $f > $tmp
	rc=$?
	if [ "$rc" -eq 0 ]; then
		mv $tmp $f
	else
		echo "failed to unindent "$f >&2
		exit 1
	fi
done
