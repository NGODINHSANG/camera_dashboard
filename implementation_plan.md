# Auto-Recording Implementation Plan

Mục tiêu của kế hoạch này là thêm chức năng tự động ghi hình (auto-record) khi có tín hiệu stream đẩy lên MediaMTX, và tự động dừng ghi hình khi mất tín hiệu stream. Kế hoạch này được thiết kế dựa trên kiến trúc hiện tại kết hợp giữa MediaMTX và Backend Camera Dashboard.

## Tóm tắt giải pháp

1. **Sử dụng tính năng `runOnReady` / `runOnNotReady` của MediaMTX**: Thay vì Backend phải liên tục kiểm tra (polling) xem stream có live hay không, MediaMTX sẽ chủ động gọi một Webhook (HTTP POST) xuống Backend mỗi khi có một stream mới bắt đầu (Ready) hoặc kết thúc (NotReady).
2. **Xử lý Webhook trên Backend**: Backend cung cấp các API nội bộ để nhận Webhook từ MediaMTX. Dựa vào `path` của stream, Backend sẽ tra cứu ID của Camera tương ứng.
3. **Tự động bắt đầu/dừng FFmpeg**: 
   - Khi nhận Webhook "Ready": Backend gọi hàm [StartRecording](file:///e:/Dashboard/backend/internal/handlers/recordings.go#67-134) để khởi tạo FFmpeg.
   - Khi nhận Webhook "NotReady": Backend gọi hàm [StopRecording](file:///e:/Dashboard/backend/internal/handlers/recordings.go#135-154) để kết thúc quá trình ghi file an toàn.
4. **Cấu hình tùy chọn Auto-Record trên Camera**: Mặc định, mỗi Camera có lưu thuộc tính `is_recording` (có thể dùng làm cờ định báo Auto-Record ON/OFF). Webhook chỉ bắt đầu ghi nếu Camera được cấu hình cho phép auto-record.

## Chi tiết các thay đổi dự kiến

### 1. Cấu hình MediaMTX ([mediamtx.yml](file:///e:/Dashboard/mediamtx.yml))

Thêm các lệnh trigger HTTP POST vào cấu hình [mediamtx.yml](file:///e:/Dashboard/mediamtx.yml) để báo hiệu cho Backend (có thể cấu hình trong mục `paths` / `all`).

```yaml
paths:
  all:
    source: publisher
    sourceOnDemand: no
    record: no
    overridePublisher: yes
    
    # Kích hoạt khi có stream đẩy lên (Ready)
    runOnReady: curl -X POST http://backend:8080/api/internal/webhooks/stream-ready -H "Content-Type: application/json" -d "{\"path\":\"$MTX_PATH\"}"
    
    # Kích hoạt khi stream ngắt kết nối (NotReady)
    runOnNotReady: curl -X POST http://backend:8080/api/internal/webhooks/stream-not-ready -H "Content-Type: application/json" -d "{\"path\":\"$MTX_PATH\"}"
```
*(Lưu ý: URL trỏ tới `http://backend:8080` vì 2 container cùng chung network `app-network`)*

### 2. Backend - Model & Routing

#### [NEW] Webhook Handlers (`internal/handlers/webhooks.go`)
Tạo một file handler mới để xử lý webhook từ MediaMTX.
- `HandleStreamReady`: Nhận HTTP POST với payload `{"path": "..."}`.
  - Tìm kiếm trong cơ sở dữ liệu `cameras` để lấy Camera phù hợp (ví dụ: `stream_url` khớp với `path` hoặc chứa `path`).
  - Kiểm tra xem Camera này đã đang được record chưa (tránh trùng lặp do reconnect).
  - Khởi tạo hàm `h.recorder.StartRecording`.
- `HandleStreamNotReady`: Nhận HTTP POST.
  - Tìm CameraID dựa trên thẻ `path`.
  - Tìm active recording của Camera này và gọi `h.recorder.StopRecording`.

#### [MODIFY] Router ([internal/router/router.go](file:///e:/Dashboard/backend/internal/router/router.go))
Đăng ký các webhook endpoints này:
```go
// Internal webhook routes (can be restricted to local docker network IPs if needed)
r.Route("/internal/webhooks", func(r chi.Router) {
    r.Post("/stream-ready", webhookHandler.HandleStreamReady)
    r.Post("/stream-not-ready", webhookHandler.HandleStreamNotReady)
})
```

#### [MODIFY] API Cameras / Models (nếu cần)
- Cần đảm bảo có cơ chế ánh xạ rõ ràng giữa **MediaMTX path** (ví dụ: `camera_01`) và thuộc tính của **Camera** trong Database (có thể dùng trường `stream_url` chứa path này).
- Tận dụng trường [IsRecording](file:///e:/Dashboard/backend/internal/services/recorder.go#361-373) hiện tại trên Camera request (`req.IsRecording`) để quyết định xem Camera này CÓ cho phép Tự động Record khi có luồng hay không. Nếu `IsRecording == false`, webhook sẽ bỏ qua việc gọi StartRecording.

### 3. Workflow Luồng Dữ Liệu

1. **Thiết bị (Camera / Drone)** đẩy rtmp stream lên: `rtmp://<ip>:1935/live/camera1`
2. **MediaMTX** nhận được stream ở path `live/camera1`. Nó thực thi lệnh cURL báo cho Backend.
3. **Backend** nhận được trigger, lookup DB thấy `live/camera1` tương ứng với Camera "Drone 1" (ID: 5). Camera 5 có thuộc tính bật Auto-Record.
4. **Backend** kích hoạt luồng [ffmpeg](file:///e:/Dashboard/backend/internal/services/recorder.go#21-25) nội bộ pull luồng từ MediaMTX lưu thành file MP4 vào ổ đĩa. Nó cập nhật Database status thành "recording".
5. Giao diện người dùng Dashboard sẽ tự động hiển thị Camera "Drone 1" đang trong trạng thái "Recording" khi gọi polling API `/api/recordings/active`.
6. Khi **Thiết bị ngưng đẩy stream**, MediaMTX gọi webhook NotReady. Backend send lệnh `q` cho ffmpeg và chốt file MP4, update status db thành "completed".

## Kế hoạch kiểm thử hệ thống

### Kiểm tra tĩnh/Unit:
- Đảm bảo logic lookup đúng camera dựa trên `$MTX_PATH`.

### Manual Verification (Khi anh setup lên server):
- Đẩy một luồng test bằng OBS hoặc FFmpeg rtmp giả lập lên MediaMTX.
- Quan sát logs của Backend để confirm webhook `/stream-ready` đã được gọi thành công và [ffmpeg](file:///e:/Dashboard/backend/internal/services/recorder.go#21-25) được trigger.
- Stop OBS/FFmpeg rtmp đẩy lên phương tiện. Backend nhận webhook `/stream-not-ready`, kết thúc ffmpeg graceful và MP4 file đã sẵn sàng phục vụ playback.

> [!TIP]
> **Định danh Path**: Để hoạt động ổn định nhất, khi anh gán Stream URL cho Camera trên Frontend Dashboard, thay vì nhập IP thì nên có format thống nhất. Ví dụ: nhập `/live/cam1` hoặc tên định danh để Backend dễ match với chuỗi từ `$MTX_PATH` của MediaMTX.
