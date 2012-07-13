#!/bin/bash

echo -e "\nBuilding Go..."
go install -ldflags "-X main.version `git rev-parse HEAD` -s"
