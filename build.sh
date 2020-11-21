#!/usr/bin/env bash
# Stops the process if something fails
set -xe

# create the application binary that eb uses
GOOS=linux GOARCH=amd64 go build  -ldflags="-s -w" -o ./application ./src/server.go 

zip aws-eb-swan-demo.zip application appsettings.json Procfile images/190811762.jpeg images/221406343.jpeg images/234657570.jpeg