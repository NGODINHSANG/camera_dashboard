# Scripts - Hướng dẫn sử dụng

## Set Admin Role

Có 3 cách để set một user thành admin:

### Cách 1: Sử dụng Go script (Khuyên dùng)

```bash
# Di chuyển vào thư mục backend
cd backend

# Set admin bằng email
go run cmd/set-admin/main.go -email=admin@example.com

# Set admin bằng ID
go run cmd/set-admin/main.go -id=1

# Sử dụng đường dẫn database tùy chỉnh
go run cmd/set-admin/main.go -email=admin@example.com -db=./custom/path/dashboard.db
```

### Cách 2: Sử dụng SQL script

```bash
# Di chuyển vào thư mục backend
cd backend

# Chỉnh sửa file scripts/set_admin.sql để thay đổi email hoặc ID
# Sau đó chạy:
sqlite3 ./data/dashboard.db < ./scripts/set_admin.sql
```

### Cách 3: Sử dụng sqlite3 command line

```bash
# Mở database
sqlite3 ./backend/data/dashboard.db

# Chạy lệnh SQL
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';

# Kiểm tra kết quả
SELECT id, email, name, role FROM users;

# Thoát
.exit
```

## Kiểm tra danh sách users

```bash
cd backend
sqlite3 ./data/dashboard.db "SELECT id, email, name, role FROM users;"
```

## Lưu ý

- Role có 2 giá trị: `admin` và `user`
- Chỉ admin mới có quyền:
  - Thêm/Sửa/Xóa dự án
  - Thêm/Sửa/Xóa camera
- User thường chỉ có quyền xem camera
