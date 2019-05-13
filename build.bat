SET APP_DIR=build
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\gorunex.exe github.com/fpawel/gorunex/cmd
buildmingw32 go build -o %APP_DIR%\elco.exe github.com/fpawel/elco/cmd/elco
