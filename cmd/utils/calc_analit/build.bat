set GOARCH=386
rsrc -manifest calc.exe.manifest -o rsrc.syso
go build -ldflags="-H windowsgui"