@echo off
echo ========================================
echo   SET ADMIN ROLE - Camera Dashboard
echo ========================================
echo.

if "%1"=="" (
    echo Cach su dung:
    echo   set-admin.bat admin@example.com
    echo   set-admin.bat user@dashboard.com
    echo.
    echo Hoac nhap email truc tiep:
    set /p email="Nhap email cua user: "
) else (
    set email=%1
)

if "%email%"=="" (
    echo Loi: Email khong duoc de trong!
    pause
    exit /b 1
)

echo.
echo Dang cap nhat user "%email%" thanh admin...
cd ..
go run cmd/set-admin/main.go -email=%email%

echo.
pause
