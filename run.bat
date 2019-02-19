SET EXE_DIR=%HOMEDRIVE%%HOMEPATH%\.elco
%EXE_DIR%\hidecon
SET LOGS_DIR=%EXE_DIR%\logs
IF NOT EXIST %LOGS_DIR% mkdir %LOGS_DIR%
SET CUR_YYYY=%DATE:~6,4%
SET CUR_MM=%date:~3,2%
SET CUR_DD=%date:~7,2%
SET CUR_HH=%time:~0,2%
IF %CUR_HH% lss 10 (set CUR_HH=0%time:~1,1%)
SET CUR_NN=%time:~3,2%
SET CUR_SS=%time:~6,2%
SET CUR_MS=%time:~9,2%
SET PANIC_HTML_FILE=%LOGS_DIR%\panic.html
SET LOG_FILE=%LOGS_DIR%\%CUR_YYYY%-%CUR_MM%-%CUR_DD%-%CUR_HH%_%CUR_NN%_%CUR_SS%.log
IF EXIST %PANIC_HTML_FILE% del %PANIC_HTML_FILE%
%EXE_DIR%\elco 2>>%LOG_FILE% 1>&2
pp -html %PANIC_HTML_FILE% %LOG_FILE%
IF EXIST %PANIC_HTML_FILE%  (
    taskkill /F /IM elcoui.exe /T    
    start %LOG_FILE%
    start %PANIC_HTML_FILE%
    msg * "elco.exe: a error occured!"
)