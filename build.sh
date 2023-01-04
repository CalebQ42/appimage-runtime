#!/bin/sh

go build -trimpath -ldflags="-linkmode=external -s -w"