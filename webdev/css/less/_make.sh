#!/bin/sh

echo -e "\nBuilding LESS files"

TARGET="./web/assets/css"
mkdir -p ${TARGET}

FILES=(
	$(dirname ${0})"/bootstrap.less"
	$(dirname ${0})"/responsive.less"
)

function log {
	echo "LESS->CSS: ${1} -> ${TARGET}/$(basename ${1} .less).css"
}

function compile {
	lessc ${1} "${TARGET}/$(basename ${1} .less).css";
}

for file in ${FILES[@]}; do
	log ${file}
	compile ${file}
done
