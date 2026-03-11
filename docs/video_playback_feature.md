# Tính năng Xem lại Video (Video Playback)

## 1. Tổng quan

Tính năng cho phép người dùng **ghi hình** từ camera giám sát và **xem lại** trực tiếp trên giao diện web Dashboard. Người dùng có thể tua nhanh, tua chậm, chọn thời điểm phát tùy ý mà không cần tải file về máy.

Hệ thống sử dụng giao thức **HLS (HTTP Live Streaming)** – chia video thành các đoạn nhỏ (4-6 giây/đoạn), trình duyệt chỉ tải đoạn đang xem giúp tiết kiệm băng thông.

## 2. Công nghệ sử dụng

- **React.js + HLS.js** – Giao diện web và trình phát video trên trình duyệt
- **Go (Golang)** – Backend API xử lý ghi hình, quản lý file video
- **FFmpeg** – Ghi hình từ luồng camera và chuyển đổi định dạng video
- **MediaMTX** – Server chuyển đổi luồng RTMP từ camera sang HLS
- **Nginx** – Reverse proxy phân phối video, giới hạn băng thông
- **SQLite** – Cơ sở dữ liệu lưu thông tin bản ghi hình
- **Docker Compose** – Đóng gói và triển khai toàn bộ hệ thống

## 3. Quy trình hoạt động

### Ghi hình

1. Camera đẩy luồng video RTMP vào server MediaMTX.
2. MediaMTX chuyển đổi RTMP → HLS và phát trên Dashboard.
3. Khi người dùng bấm **"Ghi hình"**, Backend gọi FFmpeg tải luồng HLS và lưu thành file `.mp4`.
4. Khi bấm **"Dừng ghi hình"**, FFmpeg dừng an toàn, file MP4 được lưu vào thư mục `recordings/`.

### Chuyển đổi HLS (tự động)

1. Backend chạy nền, quét thư mục `recordings/` mỗi 2 phút.
2. File MP4 chưa có bản HLS sẽ được FFmpeg tự động chia thành các segment `.ts` + playlist `.m3u8`.
3. Khi xem lại, trình duyệt chỉ tải từng segment thay vì toàn bộ file → tiết kiệm băng thông.

### Xem lại

1. Người dùng chọn camera → chọn file video từ danh sách.
2. Trình duyệt tải playlist HLS (`.m3u8`) từ Backend.
3. HLS.js tải từng segment 6 giây để phát video.
4. Người dùng có thể tua, tạm dừng, phát lại bình thường.

## 4. Hướng dẫn sử dụng

### Ghi hình camera

1. Mở Dashboard, chọn camera cần ghi hình.
2. Bấm nút **🔴 Ghi hình (Record)**.
3. Hệ thống bắt đầu ghi, biểu tượng ghi hình hiện trên camera.
4. Khi muốn dừng, bấm nút **⏹ Dừng ghi hình (Stop)**.
5. File video được lưu tự động vào thư mục `recordings/<Dự án>/<Camera>/`.

### Xem lại video đã ghi

1. Bấm vào camera muốn xem lại.
2. Chọn tab **"Xem lại" (Playback)**.
3. Danh sách các file video đã ghi sẽ hiện ra (sắp xếp mới nhất trước).
4. Bấm vào file muốn xem → video phát ngay trên Dashboard.
5. Sử dụng thanh tiến trình để tua đến vị trí bất kỳ.

### Tải video về máy

1. Trong danh sách video, bấm nút **⬇ Tải xuống (Download)**.
2. File `.mp4` sẽ được tải về máy tính.

### Xóa video

1. Trong danh sách video, bấm nút **🗑 Xóa (Delete)**.
2. Xác nhận xóa → file bị xóa khỏi server.
