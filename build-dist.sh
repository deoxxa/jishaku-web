#!/bin/sh

REV=$(git show HEAD | head -n 1 | awk '{print $2}' | head -c 10)
TIME=$(date -u '+%Y%m%dT%H%I%SZ')

GOOS=windows GOARCH=386 go build -o jishaku.exe fknsrs.biz/p/jishaku-web/server
zip -r dist_windows_${TIME}_${REV}.zip jishaku.exe public templates
rm jishaku.exe

GOOS=darwin GOARCH=amd64 go build -o jishaku fknsrs.biz/p/jishaku-web/server
zip -r dist_mac_${TIME}_${REV}.zip jishaku public templates
rm jishaku

GOOS=linux GOARCH=amd64 go build -o jishaku fknsrs.biz/p/jishaku-web/server
zip -r dist_linux_${TIME}_${REV}.zip jishaku public templates
rm jishaku
