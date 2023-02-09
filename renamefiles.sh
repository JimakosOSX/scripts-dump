#!/usr/bin/env bash
#
#
# exit on error
set -e 

usage() {
    cat << EOM
    Usage:
    $(basename $0) TARGET_DIR
EOM
}

if [[ $# == 0 ]];then
    usage;
    exit 1
fi

# preparing
OLDIFS=$IFS
IFS=$'\n'
TARGET_DIR=$1
cd $TARGET_DIR

# Main logic is here
for file in $(ls);do
    hash=$(md5sum $file | cut -d ' ' -f1)
    filename=$(basename -- "$file")
    extension="${filename##*.}"
    filename="${filename%.*}"
    mv -v ${file} ${hash}.${extension}
done

# Cleanup
IFS=$OLDIFS
cd - 
echo "Done."

exit 0
