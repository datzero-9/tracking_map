# Tracking Map 
## 1. Mô tả kiến trúc hệ thống
Hệ thống được thiết kế theo mô hình *Client-Server*, bao gồm 3 thành phần chính:

* **Frontend:** Xây dựng bằng **React.js**.
  * Sử dụng thư viện **React-Leaflet** kết hợp với nền bản đồ của **OpenStreetMap** .
  * Giao diện UI/UX được tối ưu hóa: Tự động zoom vừa vặn tuyến đường, hỗ trợ chọn điểm B trực tiếp bằng cú click chuột trên bản đồ.
* **Backend:** Xây dựng bằng **Golang** với framework **Gin**.
  * Cung cấp các RESTful API phục vụ: Cập nhật GPS, lấy vị trí mới nhất, trích xuất lịch sử theo khoảng thời gian và tính toán đường đi.
* **Database:** Sử dụng **PostgreSQL** chạy qua **Docker** (`tracking_db`).
  * Lưu trữ cấu trúc dữ liệu thiết bị: `device_id`, `latitude`, `longitude`, `timestamp`.

## 2. Routing engine / Service đã sử dụng
Hệ thống sử dụng **OSRM** thông qua Public API của dự án.
* **Cơ chế hoạt động:** Frontend gửi tọa độ Điểm A (Start) và Điểm B (End) xuống Backend. Backend đóng vai trò gọi sang API của OSRM (`router.project-osrm.org/route/v1/driving/...`).
* **Dữ liệu trả về:** Hệ thống tính toán dựa trên mạng lưới đường xá thực tế và bóc tách các thông số:
  * Trả về mảng tọa độ `geometry (geojson)` để Frontend vẽ Polyline bám sát mặt đường thực tế.
  * Tính toán và hiển thị khoảng cách tổng `distance_km`.
  * Tính toán và hiển thị thời gian di chuyển dự kiến `duration`.

## 3. Hướng dẫn cài đặt và chạy Project

Yêu cầu môi trường: Cài đặt sẵn **Docker**, **Go (Golang)**, và **Node.js**.

### Bước 1: Khởi động Database (PostgreSQL)
Mở Terminal / PowerShell tại thư mục gốc của project:
docker-compose up -d ( lần đầu khởi động )
docker start tracking_db ( các lần sau chạy leengj này )
### Bước 2: Khởi động Backend (Golang)
### Di chuyển vào thư mục backend (nơi chứa main.go)
### Tải các thư viện
go mod tidy
### Khởi chạy server API 
go run main.go

Bước 3: Khởi động Frontend (React.js)
### Di chuyển vào thư mục frontend
### Cài đặt các thư viện 
npm install
### Khởi chạy giao diện Web
npm start

### 1 số lưu ý
### để xem lịch di chuyển cần 1 số dữ liệu, nên phải fake dữ liệu:
### tại terminal dự án chạy lệnh: 
docker exec -it tracking_db psql -U user_tracking -d tracking_map
### sau đó nhập:
INSERT INTO device_locations (device_id, latitude, longitude, timestamp) VALUES 
('car_01', 16.04500, 108.20500, NOW() - INTERVAL '90 minutes'),
('car_01', 16.04700, 108.20600, NOW() - INTERVAL '60 minutes'),
('car_01', 16.04900, 108.20800, NOW() - INTERVAL '30 minutes'),
('car_01', 16.05200, 108.21100, NOW() - INTERVAL '5 minutes');
### chú ý chọn thiết bị và thời gian để xem lịch sử
