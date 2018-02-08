#!/bin/bash

export GOOS=linux
go build -o Handler
zip Handler.zip Handler
rm -f Handler
