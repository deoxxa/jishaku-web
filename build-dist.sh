#!/bin/sh

GOOS=windows GOARCH=386 go build -o jishaku.exe fknsrs.biz/p/jishaku-web/server
zip -r dist-windows.zip jishaku.exe public templates
rm jishaku.exe

GOOS=darwin GOARCH=amd64 go build -o jishaku fknsrs.biz/p/jishaku-web/server
zip -r dist-mac.zip jishaku public templates
rm jishaku

GOOS=linux GOARCH=amd64 go build -o jishaku fknsrs.biz/p/jishaku-web/server
zip -r dist-linux.zip jishaku public templates
rm jishaku
