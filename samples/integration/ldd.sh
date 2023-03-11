#!/bin/sh
set -o errexit
set -o nounset
set -o pipefail

# If there are multiple files found, only use the first result
file_path=$(find ./ -name "$FUNC_NAME" | head -n 1)
if [ "$file_path" ]; then
    mv "$file_path" "/app/$FUNC_NAME"
    echo "loading app $FUNC_NAME"
else
# Exit and report an err when no $FUNC_NAME file is found
    echo "cannot find file $FUNC_NAME!" >&2
    exit 1
fi

if [ ! -d "/shared_lib" ]; then
    mkdir "/shared_lib"
fi
cd "/shared_lib"
echo "The shared_lib dir created"

# Identify which libs need to loaded
sharedlibs=$(ldd "/app/$FUNC_NAME" | grep -vE "ld-musl-x86_64|mpi" | awk '{print $3}' || true)
if [ "$sharedlibs" != "" ]; then
    echo "shared libs found"
    echo "$sharedlibs" > path.txt
else
    echo "no shared lib need to be loaded"
    exit 0
fi

# Seek for the share libs
while read -r line
do
    if [ "$line" = "" ]; then continue; fi
    islink=$(ls -l "$line" | grep ' -> ' || true)
    if [ "$islink" != "" ]
    then
        dir=$(dirname "$line")
        base=$(basename "$line")
        file=$(ls -l "$line" | awk '{print $NF}')
        cp -p "${dir}/${file}" "/shared_lib/$base"
        echo "cp ${dir}/${file} /shared_lib/$base"
    else
        cp -p "$line" "/shared_lib/"
        echo "cp $line /shared_lib/"
    fi
done < path.txt
rm path.txt
echo "ldd analysis success!"