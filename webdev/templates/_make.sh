#!/bin/sh

echo -e "\nBuilding HAML files..."

TARGET="./website/templates"
if [ ! -d ${TARGET} ]; then
  mkdir -p ${TARGET}
fi

for file in $(find $(dirname ${0}) -type f -name "*.haml"); do
	echo "HAML->HTML: ${file} -> ${TARGET}/$(basename ${file} .haml).tpl"
	haml ${file} "${TARGET}/$(basename ${file} .haml).tpl";
done
