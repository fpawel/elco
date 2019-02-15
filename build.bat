SET APP_DIR=%HOMEDRIVE%%HOMEPATH%\.elco
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\elco.exe github.com/fpawel/elco/cmd/elco
start %APP_DIR%
