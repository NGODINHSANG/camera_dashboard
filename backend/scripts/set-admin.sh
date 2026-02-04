#!/bin/bash

echo "========================================"
echo "  SET ADMIN ROLE - Camera Dashboard"
echo "========================================"
echo ""

if [ -z "$1" ]; then
    echo "Cách sử dụng:"
    echo "  ./scripts/set-admin.sh admin@example.com"
    echo "  ./scripts/set-admin.sh user@dashboard.com"
    echo ""
    read -p "Hoặc nhập email của user: " email
else
    email=$1
fi

if [ -z "$email" ]; then
    echo "Lỗi: Email không được để trống!"
    exit 1
fi

echo ""
echo "Đang cập nhật user '$email' thành admin..."
cd "$(dirname "$0")/.."
go run cmd/set-admin/main.go -email="$email"

echo ""
read -p "Nhấn Enter để thoát..."
