#!/bin/sh

echo -e "\nBuilding HAML files..."

TARGET="./web/templates"

for file in $(find $(dirname ${0}) -type f -name "*.haml"); do
	echo "HAML->HTML: ${file} -> ${TARGET}/$(basename ${file} .haml).html"
	haml ${file} "${TARGET}/$(basename ${file} .haml).html";
done
