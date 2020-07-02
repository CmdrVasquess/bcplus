#!/bin/sh
test=$1
if [ ! -d "$test" ]; then
    echo "No test directory: "$test
    exit 2
fi
TSTD=`mktemp -d -p $test test.XXXX`
mkdir -p $TSTD/j

../bcplus -web-tls=false \
	  -log d -log-to $TSTD/bcplus.log \
	  -d $TSTD -j $TSTD/j 2> /dev/null &
BCPPID=$!

sleep 1
for journal in `ls -1 $test/Journal.*.log`; do
    jreplay -v -j $TSTD/j -p 300ms $journal
done
kill -INT $BCPPID
wait $BCPPID

texst -r $test/bcplus.json.texst $TSTD/bcplus.json | tee $TSTD/texst-bcplus.json

#texsts=`find $test -name '*.texst'`
#for t in $texsts; do
#    dir=`dirname $t`
#    base=`basename $t .texst`
#    echo $t
#    echo $TSTD/$base
#done



