#!/bin/sh
test=$1
if [ ! -d "$test" ]; then
    echo "No test directory: "$test
    exit 2
fi
TSTD=`mktemp -d -p $test test.XXXX`
mkdir -p $TSTD/j

../bcplus --log d --log-to $TSTD/bcplus.log -d $TSTD -j $TSTD/j &
BCPPID=$!

sleep 1
for journal in `ls -1 $test/Journal.*.log`; do
    jreplay -j $TSTD/j -p 300ms $journal
done
kill -INT $BCPPID
wait $BCPPID

texst -r $test/bcplus.json.texst $TSTD/bcplus.json | tee $TSTD/texst-bcplus.json
