#!/bin/bash
set -eu -o pipefail -o nounset

exclude=${EXCLUDE:-}

if [[ -z ${exclude} ]]; then
    echo "$PWD/Procfile"
    exit 0
fi

tempfile=$(mktemp)
cat ./Procfile > $tempfile

IFS=', ' read -r -a excluded <<< "$exclude"

for PROC in "${excluded[@]}"
do
    sed -i "" "/^$PROC/d" $tempfile
    #cat $tempfile
    #echo $filtered > $tempfile
done

echo $tempfile