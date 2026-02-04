# Hướng dẫn chuyển đổi Markdown sang Word

## Phương pháp 1: Sử dụng Script tự động (Khuyên dùng)

### Cài đặt Pandoc

**Windows (Winget):**
```bash
winget install --id JohnMacFarlane.Pandoc
```

**Windows (Manual):**
1. Tải về: https://pandoc.org/installing.html
2. Chạy file cài đặt `.msi`
3. Khởi động lại terminal

### Chạy script

```bash
cd e:\Dashboard\docs
convert-to-word.bat
```

Kết quả: Tạo 6 file `.docx` trong thư mục `docs/`

---

## Phương pháp 2: Chuyển đổi thủ công với Pandoc

```bash
pandoc README.md -o README.docx
pandoc API.md -o API.docx
pandoc ARCHITECTURE.md -o ARCHITECTURE.docx
pandoc DATABASE.md -o DATABASE.docx
pandoc DEPLOYMENT.md -o DEPLOYMENT.docx
pandoc TECHNOLOGY.md -o TECHNOLOGY.docx
```

---

## Phương pháp 3: Online Converters (Không cần cài đặt)

### Dillinger.io
1. Mở: https://dillinger.io/
2. Paste nội dung Markdown
3. Export → Word

### CloudConvert
1. Mở: https://cloudconvert.com/md-to-docx
2. Upload file `.md`
3. Convert và download

### Aspose
1. Mở: https://products.aspose.app/words/conversion/md-to-docx
2. Upload file `.md`
3. Convert và download

---

## Phương pháp 4: Microsoft Word

1. Mở Microsoft Word
2. File → Open → chọn file `.md`
3. Word tự động nhận diện Markdown
4. Save As → `.docx`

---

## Phương pháp 5: VSCode Extension

1. Cài extension: **Markdown PDF**
2. Mở file `.md` trong VSCode
3. `Ctrl+Shift+P` → "Markdown PDF: Export (docx)"
4. File `.docx` được tạo cùng thư mục

---

## Danh sách file cần chuyển đổi

- README.md → README.docx
- API.md → API.docx
- ARCHITECTURE.md → ARCHITECTURE.docx
- DATABASE.md → DATABASE.docx
- DEPLOYMENT.md → DEPLOYMENT.docx
- TECHNOLOGY.md → TECHNOLOGY.docx

---

## Lưu ý

- Phương pháp 1 (Pandoc) cho kết quả tốt nhất
- Phương pháp 3 (Online) nhanh nhưng cần internet
- Phương pháp 4 (Word) đơn giản nhưng format có thể khác
- Phương pháp 5 (VSCode) tiện nếu đã dùng VSCode
