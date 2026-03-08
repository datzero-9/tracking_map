package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"math"
)

// DownloadTile tải 1 mảnh ảnh từ Google và lưu vào thư mục local
func DownloadTile(z, x, y int) error {
	// 1. Tạo URL dựa trên tọa độ
	url := fmt.Sprintf("http://mt1.google.com/vt/lyrs=m&x=%d&y=%d&z=%d", x, y, z)

	// 2. Tạo đường dẫn lưu file (Ví dụ: tiles/13/x/y.png)
	dir := fmt.Sprintf("tiles/%d/%d", z, x)
	os.MkdirAll(dir, os.ModePerm) // Tạo thư mục nếu chưa có
	filePath := filepath.Join(dir, fmt.Sprintf("%d.png", y))

	// 3. Kiểm tra nếu file đã tồn tại thì không tải lại (Caching đơn giản)
	if _, err := os.Stat(filePath); err == nil {
		return nil 
	}

	// 4. Tiến hành tải ảnh
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lỗi khi tải tile: %v", err)
	}
	defer resp.Body.Close()

	// 5. Ghi dữ liệu vào file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func LonLatToTile(lon, lat float64, zoom int) (int, int) {
	n := math.Exp2(float64(zoom))
	x := int(math.Floor((lon + 180.0) / 360.0 * n))
	latRad := lat * math.Pi / 180.0
	y := int(math.Floor((1.0 - math.Log(math.Tan(latRad)+(1/math.Cos(latRad)))/math.Pi) / 2.0 * n))
	return x, y
}

// 2. Hàm gom cụm: Tải toàn bộ khu vực xung quanh 1 điểm
func DownloadMapArea(centerLon, centerLat float64) {
	zooms := []int{13, 14, 15} // Đạt chuẩn yêu cầu: Load ở 3 mức zoom
	radius := 2 // Tải rộng ra xung quanh 2 mảnh (để khi người dùng kéo chuột không bị màn hình xám)

	fmt.Println("⏳ Đang tải bản đồ Google Offline...")
	for _, z := range zooms {
		centerX, centerY := LonLatToTile(centerLon, centerLat, z)

		for x := centerX - radius; x <= centerX + radius; x++ {
			for y := centerY - radius; y <= centerY + radius; y++ {
				err := DownloadTile(z, x, y)
				if err != nil {
					fmt.Printf("❌ Lỗi tải tile z:%d x:%d y:%d\n", z, x, y)
				}
			}
		}
	}
	fmt.Println("✅ Hoàn tất tải bản đồ Offline (Lưu tại thư mục /tiles)!")
}