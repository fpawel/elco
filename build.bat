SET dir=%HOMEDRIVE%%HOMEPATH%\.elco
set GOARCH=386
buildmingw32 go build -o %dir%\elco.exe github.com/fpawel/elco/cmd/app
start %dir%
