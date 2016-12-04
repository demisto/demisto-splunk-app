#!/usr/bin/env bash

VERSION="1.0.6"
SPLFILE="SA-DemistoAlertApp_$VERSION.spl"

cd SA-DemistoAlertApp
echo "[-] Building..."
go build
if [ $? -ne '0' ];then
	echo "[x] Build failed!"
	exit 1
fi
rm -f ./SA-DemistoAlertApp

# Build darwin
echo "[-] Building darwin 32 bit"
mkdir -p darwin_x86/bin/
GOOS=darwin GOARCH=386 go build -o darwin_x86/bin/demisto_app.exe

echo "[-] Building darwin 64 bit"
mkdir -p darwin_x86_64/bin/
GOOS=darwin GOARCH=amd64 go build -o darwin_x86_64/bin/demisto_app.exe

# Build windows
echo "[-] Building windows 32 bit"
mkdir -p windows_x86/bin/
GOOS=windows GOARCH=386 go build -o windows_x86/bin/demisto_app.exe

echo "[-] Building windows 64 bit"
mkdir -p windows_x86_64/bin/
GOOS=windows GOARCH=amd64 go build -o windows_x86_64/bin/demisto_app.exe

# Build linux
echo "[-] Building linux 32 bit"
mkdir -p linux_x86/bin/
GOOS=linux GOARCH=386 go build -o linux_x86/bin/demisto_app.exe

echo "[-] Building linux 64 bit"
mkdir -p linux_x86_64/bin/
GOOS=linux GOARCH=amd64 go build -o linux_x86_64/bin/demisto_app.exe

echo "[+] Done"

cd ..
echo "[-] Creating splunk app package..."
tar -zcvf demisto_app.tar.gz --exclude='*.go' SA-DemistoAlertApp/
if [ $? -ne '0' ];then
	echo "[x] Build failed!"
	exit 2
fi
mv demisto_app.tar.gz $SPLFILE
echo "[+] Done"
