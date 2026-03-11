# Tracking Map Project

## 1. Mô tả kiến trúc hệ thống
Hệ thống được thiết kế theo mô hình **Microservices**, đóng gói hoàn toàn bằng Docker, bao gồm 4 thành phần chính hoạt động độc lập và giao tiếp qua mạng nội bộ (Docker Network):

* **Frontend (React.js):**
  * Sử dụng thư viện **React-Leaflet** kết hợp với nền **Bản đồ Offline (Tile Server nội bộ)**, cho phép tải trang và hiển thị mượt mà không phụ thuộc hoàn toàn vào nguồn map bên ngoài.
  * Giao diện UI/UX được tối ưu hóa: Tự động zoom vừa vặn tuyến đường (RouteFitter), hỗ trợ chọn Điểm B trực tiếp bằng cú click chuột trên bản đồ.
* **Backend (Golang / Gin Framework):**
  * Cung cấp các RESTful API phục vụ: Cập nhật GPS, lấy vị trí mới nhất, trích xuất lịch sử theo khoảng thời gian và tính toán đường đi.
  * Tích hợp cơ chế chạy ngầm để tự động tải và lưu trữ các mảnh bản đồ (Map Tiles) về ổ cứng phục vụ cho Frontend.
* **Database (PostgreSQL):**
  * Chạy qua Docker container (`tracking_db`). Lưu trữ cấu trúc dữ liệu thiết bị: `device_id`, `latitude`, `longitude`, `timestamp`.
* **Cache & Middleware (Redis):**
  * Chạy qua Docker container (`redis_cache`), đóng vai trò tối ưu hóa hiệu suất và bảo vệ hệ thống.

---

## 2. Routing engine / Service đã sử dụng
Hệ thống sử dụng **OSRM (Open Source Routing Machine)** thông qua Public API của dự án.
* **Cơ chế hoạt động:** Frontend gửi tọa độ Điểm A  và Điểm B xuống Backend. Backend đóng vai trò như một Proxy, giả lập Custom HTTP Client để gọi sang API của OSRM .
* **Dữ liệu trả về:** Hệ thống tính toán dựa trên mạng lưới đường xá thực tế và bóc tách các thông số:
  * Trả về mảng tọa độ `geometry (geojson)` để Frontend vẽ Polyline bám sát mặt đường thực tế.
  * Tính toán và hiển thị khoảng cách tổng `distance` (km).
  * Tính toán và hiển thị thời gian di chuyển dự kiến `duration` (phút).

---

## 3. Redis được dùng để làm gì?
Trong dự án này, Redis được ứng dụng để giải quyết 3 bài toán kiến trúc quan trọng:

1. **Rate Limiting (Chống Spam API):** Redis được dùng làm Middleware đếm số lượng Request của từng IP. Nếu có dấu hiệu spam API, hệ thống sẽ tự động chặn  để bảo vệ Database PostgreSQL không bị quá tải. Đặc biệt, luồng tải ảnh bản đồ được cấu hình bypass (bỏ qua) giới hạn này để đảm bảo bản đồ load mượt mà.
2. **Cache Route :** Các tuyến đường OSRM tính toán xong sẽ được lưu vào Redis với TTL  là 7 ngày. Khi yêu cầu lại tuyến đường trùng lặp, Backend trả kết quả ngay lập tức từ RAM thay vì gọi API OSRM, giúp tăng tốc phản hồi và chống bị OSRM chặn IP.
3. **Cache Latest Location :** Mỗi khi thiết bị gửi tọa độ mới, hệ thống lưu đồng thời vào PostgreSQL và RAM của Redis. Khi Frontend yêu cầu vị trí mới nhất, Backend sẽ lấy trực tiếp từ Redis  thay vì Query vào Database, giúp giảm tải triệt để cho hệ cơ sở dữ liệu.

---

## 4. Hướng dẫn cài đặt và chạy Project

Hệ thống đã được tự động hóa hoàn toàn. 
**Yêu cầu môi trường duy nhất:** Máy tính cài đặt sẵn **Docker** và **Docker Compose**.

### Bước 1: Khởi động toàn bộ hệ thống
Mở Terminal / PowerShell tại thư mục gốc của project (nơi chứa file `docker-compose.yml`) và chạy lệnh:


**Lần đầu tiên khởi chạy (sẽ build image và tải thư viện):**

* docker-compose up --build 

**Các lần chạy sau (khởi động nhanh):**

* docker-compose up -d

### Bước 2: Truy cập ứng dụng
Sau khi hệ thống khởi động thành công (các container báo `Started`), hãy mở trình duyệt:

* **Frontend Web:** `http://localhost:3000`
* **Backend API / Tile Server:** `http://localhost:8080`

### Bước 3: Thêm dữ liệu mẫu (Fake Data) để test Lịch sử
**Để xem được lịch sử di chuyển, bạn cần có dữ liệu trong Database. Mở một Terminal mới và chạy lệnh sau để truy cập thẳng vào Database bên trong Docker:**

* docker exec -it tracking_map-postgres_db-1 psql -U user_tracking -d tracking_map

*INSERT INTO device_locations (device_id, latitude, longitude, timestamp) VALUES 
('car_01', 16.04500, 108.20500, NOW() - INTERVAL '90 minutes'),
('car_01', 16.04700, 108.20600, NOW() - INTERVAL '60 minutes'),
('car_01', 16.04900, 108.20800, NOW() - INTERVAL '30 minutes'),
('car_01', 16.05200, 108.21100, NOW() - INTERVAL '5 minutes');*


### Bước 4: Tắt hệ thống an toàn
**Khi không sử dụng, hãy tắt hệ thống để giải phóng tài nguyên máy bằng lệnh:**

* docker-compose down

