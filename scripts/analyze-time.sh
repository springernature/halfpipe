#!/usr/bin/env bash

######### example usage ########################
# $ source analyze-time.sh
# $ downloadForTeam oscar
# $ analyzeGitGet oscar | sort -n | uniq -c
################################################

function downloadForTeam {
	echo "Gonna get the build events from the last 500 builds, and filter out the succeeded ones and download the event logs for those"
	TEAM=$1
	mkdir -p /tmp/$TEAM
	for id in `fly -t $TEAM builds -c 500 --json | jq -c '.[] | select( .status == "succeeded" ) | .id'`; do
		echo $id
		# Since the loadbalancer doesnt seem to terminate the connection, just grab whatever we can in 2 seconds.
		fly -t $TEAM curl /api/v1/builds/$id/events -- -m 2 | egrep -o "{(.*)}" | > /tmp/$TEAM/$id 
		echo "==="
	done
}

# JQ is a bit slow to initialise, this function takes some time..
function analyzeGitGet {
	TEAM=$1
	DATA_FILE=`mktemp`

	for FILE in /tmp/$TEAM/*; do
		ID=`cat $FILE | grep author_date | head -n1 | jq -r '.data.origin.id'`
		START_TIME=`cat $FILE | grep $ID | grep "initialize-get" | jq '.data.time'`
		END_TIME=`cat $FILE | grep $ID | grep "finish-get" | jq '.data.time'`
		echo | awk -v start="$START_TIME" -v end="$END_TIME" '{print end-start}' >> $DATA_FILE
	done
	cat $DATA_FILE | sort -n | uniq -c
	cat $DATA_FILE | awk -F : '{sum+=$1} END {print "We are spending on avarage",sum/NR, "seconds to get the git repository. Based on",NR,"datapoints"}'
	cat $DATA_FILE | sort -n | awk ' { a[i++]=$1; } END { print "median time to get", a[int(i/2)]; }'
}
