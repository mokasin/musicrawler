#!/bin/sh
echo -e "\nBuilding Coffeescript"

coffee -c -o website/assets/js/ -j $(dirname ${0})/*.coffee $(dirname ${0})/tracks-json.coffee
