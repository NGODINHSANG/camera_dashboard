-- Script để set user thành admin
-- Chạy script này bằng cách:
-- sqlite3 ./data/dashboard.db < ./scripts/set_admin.sql

-- Update user với email 'admin' hoặc 'admin@example.com' thành role admin
UPDATE users SET role = 'admin' WHERE email IN ('admin', 'admin@example.com', 'admin@dashboard.com');

-- Hoặc update user cụ thể bằng ID (thay 1 bằng ID của user)
-- UPDATE users SET role = 'admin' WHERE id = 1;

-- Hiển thị danh sách users để kiểm tra
SELECT id, email, name, role FROM users;
