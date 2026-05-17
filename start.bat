@echo off
if "%1"=="" (
  start "" /B "D:\ccm-desktop-v2\node_modules\electron\dist\electron.exe" "D:\ccm-desktop-v2\desktop\main.js"
)
