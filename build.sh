#!/bin/bash

BUILDDATE=`date +%Y%m%d`
COMMIT=`git rev-parse --short HEAD`

GOOS=windows go build -o SecretScanner-windows-x64-$BUILDDATE-$COMMIT.exe
GOOS=darwin go build -o SecretScanner-osx-x64-$BUILDDATE-$COMMIT
GOOS=linux go build -o SecretScanner-linux-x64-$BUILDDATE-$COMMIT
