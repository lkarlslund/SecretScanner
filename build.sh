#!/bin/bash

BUILDDATE=`date +%Y%m%d`
COMMIT=`git rev-parse --short HEAD`

GOOS=linux go build -o SecretScanner-linux-x64-$BUILDDATE-$COMMIT
