#!/bin/bash

echo -e "\nBuilding Go..."
go install -ldflags "-X main.version `date -u +%Y%m%d.%M%S` -s"
