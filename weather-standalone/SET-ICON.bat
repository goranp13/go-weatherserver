@echo off
title Setting Icon...
echo.
echo ========================================
echo    🌤️  VREMENSKA PROGNOZA
echo ========================================
echo.
echo Setting custom icon for weather app...
echo.

REM Create a shortcut with custom icon
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $shortcut = $WshShell.CreateShortcut('%USERPROFILE%\Desktop\Vremenska Prognoza.lnk'); $shortcut.TargetPath = '%~dp0weather-silent-simple.exe'; $shortcut.IconLocation = '%~dp0weather.ico'; $shortcut.Save()"

echo.
echo ✅ Icon set! 
echo.
echo A shortcut has been created on your Desktop
echo with the weather icon.
echo.
echo You can now use the shortcut instead of the exe.
echo.
pause
