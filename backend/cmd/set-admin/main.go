package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	// Parse command line flags
	email := flag.String("email", "", "Email của user cần set thành admin")
	userID := flag.Int("id", 0, "ID của user cần set thành admin")
	dbPath := flag.String("db", "./data/dashboard.db", "Đường dẫn đến database")
	flag.Parse()

	if *email == "" && *userID == 0 {
		fmt.Println("Cách sử dụng:")
		fmt.Println("  go run cmd/set-admin/main.go -email=admin@example.com")
		fmt.Println("  go run cmd/set-admin/main.go -id=1")
		fmt.Println("  go run cmd/set-admin/main.go -email=admin@example.com -db=./custom/path/dashboard.db")
		os.Exit(1)
	}

	// Mở database
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Không thể mở database: %v", err)
	}
	defer db.Close()

	// Update user role
	var result sql.Result
	if *email != "" {
		result, err = db.Exec("UPDATE users SET role = 'admin' WHERE email = ?", *email)
		if err != nil {
			log.Fatalf("Lỗi khi update user: %v", err)
		}
	} else {
		result, err = db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", *userID)
		if err != nil {
			log.Fatalf("Lỗi khi update user: %v", err)
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Lỗi khi lấy số dòng bị ảnh hưởng: %v", err)
	}

	if rowsAffected == 0 {
		fmt.Println("⚠️  Không tìm thấy user phù hợp")
		fmt.Println("\nDanh sách users hiện có:")
		printUsers(db)
	} else {
		fmt.Printf("✅ Đã cập nhật %d user thành admin\n", rowsAffected)
		fmt.Println("\nDanh sách users sau khi cập nhật:")
		printUsers(db)
	}
}

func printUsers(db *sql.DB) {
	rows, err := db.Query("SELECT id, email, name, role FROM users")
	if err != nil {
		log.Printf("Lỗi khi query users: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n┌────────┬──────────────────────────┬──────────────────────────┬────────┐")
	fmt.Println("│ ID     │ Email                    │ Name                     │ Role   │")
	fmt.Println("├────────┼──────────────────────────┼──────────────────────────┼────────┤")

	for rows.Next() {
		var id int
		var email, name, role string
		if err := rows.Scan(&id, &email, &name, &role); err != nil {
			log.Printf("Lỗi khi scan row: %v", err)
			continue
		}
		fmt.Printf("│ %-6d │ %-24s │ %-24s │ %-6s │\n", id, email, name, role)
	}

	fmt.Println("└────────┴──────────────────────────┴──────────────────────────┴────────┘")
}
