#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/feature man | gzip -c -9 >manpages/.1.gz
