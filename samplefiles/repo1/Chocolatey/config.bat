choco source disable -n=chocolatey || exit /b
choco source add -n=pcccache -s "http://pccs3.tama-st-h.local:8005/choco/" --bypassproxy || exit /b
