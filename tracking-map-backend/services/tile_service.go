package services

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// Khai báo biến toàn cục lưu đường biên giới
var vnBoundary orb.MultiPolygon

// 1. Hàm nạp file vn.json vào RAM (Chỉ chạy 1 lần lúc bật server)
func LoadVnBoundary() error {
	data, err := os.ReadFile("vn.json")
	if err != nil {
		return fmt.Errorf("không thể đọc file vn.json: %v", err)
	}
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return fmt.Errorf("lỗi parse JSON: %v", err)
	}
	vnBoundary = fc.Features[0].Geometry.(orb.MultiPolygon)
	fmt.Println("🗺️ Đã nạp thành công dữ liệu biên giới Việt Nam!")
	return nil
}

// 2. Chuyển đổi ngược: Từ tọa độ viên gạch (Tile) sang GPS (Kinh/Vĩ độ)
func TileToLonLat(z, x, y int) (float64, float64) {
	n := math.Exp2(float64(z))
	lon := float64(x)/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	lat := latRad * 180.0 / math.Pi
	return lon, lat
}

// 3. Hàm cốt lõi: Kiểm tra viên gạch có nằm trong Việt Nam không?
func IsTileInVietnam(z, x, y int) bool {
	// Tính tọa độ GPS 4 góc của viên gạch
	lon1, lat1 := TileToLonLat(z, x, y)     
	lon2, lat2 := TileToLonLat(z, x+1, y+1) 

	centerLon := (lon1 + lon2) / 2
	centerLat := (lat1 + lat2) / 2

	// Kiểm tra 5 điểm: 4 góc và tâm. Nếu 1 điểm dính vào VN -> Giữ lại viên gạch này
	points := []orb.Point{
		{lon1, lat1}, {lon2, lat1},
		{lon1, lat2}, {lon2, lat2},
		{centerLon, centerLat},
	}

	for _, p := range points {
		if planar.MultiPolygonContains(vnBoundary, p) {
			return true
		}
	}
	return false
}

// 4. Hàm tải ảnh gốc của bạn (Giữ nguyên)
func DownloadTile(z, x, y int) error {
	url := fmt.Sprintf("http://mt1.google.com/vt/lyrs=m&x=%d&y=%d&z=%d", x, y, z)
	dir := fmt.Sprintf("tiles/%d/%d", z, x)
	os.MkdirAll(dir, os.ModePerm)
	filePath := filepath.Join(dir, fmt.Sprintf("%d.png", y))

	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lỗi khi tải tile: %v", err)
	}
	defer resp.Body.Close()

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



// 5. Cào bản đồ theo Bounding Box (Đóng khung toàn bộ Việt Nam)
func DownloadVietnamBBox() {
	// Các mức zoom bạn muốn tải (Test thử zoom 10, 11 trước nhé)
	zooms := []int{10, 11, 12} 

	// 4 Điểm cực của Việt Nam (Đã mở rộng ra một chút để bao trọn đất liền và bờ biển)
	minLon, maxLon := 102.0, 110.0 // Từ Tây sang Đông
	minLat, maxLat := 8.0, 24.0    // Từ Nam lên Bắc

	fmt.Println("🚀 Đang khởi động Crawler cào BBox Việt Nam...")
	downloadedCount := 0

	for _, z := range zooms {
		// Đổi tọa độ GPS sang số thứ tự gạch (Lưu ý: Lat max thì Y min, Lat min thì Y max)
		minX, maxY := LonLatToTile(minLon, minLat, z) 
		maxX, minY := LonLatToTile(maxLon, maxLat, z) 

		fmt.Printf("🗺️ [Zoom %d] Khung tải X: %d -> %d | Y: %d -> %d\n", z, minX, maxX, minY, maxY)

		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				// Không cần check biên giới nữa, quất hết!
				err := DownloadTile(z, x, y)
				if err == nil {
					downloadedCount++
				}

				if downloadedCount%1000 == 0 && downloadedCount > 0 {
					fmt.Printf("⏳ [Zoom %d] Đã tải/Quét qua: %d ảnh...\n", z, downloadedCount)
				}
			}
		}
	}
	fmt.Printf("✅ HOÀN TẤT CHIẾN DỊCH BBOX! Tổng số ảnh: %d.\n", downloadedCount)
}