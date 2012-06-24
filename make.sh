#!/bin/sh

echo "Building..."

for i in $(find . -type f -name "_make.sh"); do
	${i}
done

echo -e "\n-> DONE"
