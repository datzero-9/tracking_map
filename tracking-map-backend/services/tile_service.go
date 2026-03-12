package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
)

var vnBoundary orb.Geometry

func LoadVnBoundary() error {
	data, err := os.ReadFile("vn.json")
	if err != nil {
		return fmt.Errorf("không thể đọc file vn.json: %v", err)
	}
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return fmt.Errorf("lỗi parse JSON: %v", err)
	}
	vnBoundary = fc.Features[0].Geometry
	fmt.Println("🗺️ Đã nạp thành công dữ liệu biên giới Việt Nam!")
	return nil
}

func DownloadTile(z, x, y int) error {
	url := fmt.Sprintf("http://mt1.google.com/vt/lyrs=m&x=%d&y=%d&z=%d", x, y, z)
	dir := fmt.Sprintf("tiles/%d/%d", z, x)
	os.MkdirAll(dir, os.ModePerm)
	filePath := filepath.Join(dir, fmt.Sprintf("%d.png", y))

	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lỗi tải (Status: %d)", resp.StatusCode)
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

func DownloadAreaSmart(geometry orb.Geometry, zooms []int, areaName string) {
	fmt.Printf(" Bắt đầu chiến dịch: %s...\n", areaName)
	downloadedCount := 0

	for _, z := range zooms {
		// 👇 ĐÃ FIX: Hứng thêm biến 'err' trả về từ thư viện
		tiles, err := tilecover.Geometry(geometry, maptile.Zoom(z))

		// Bắt lỗi nếu file JSON hình học bị sai
		if err != nil {
			fmt.Printf("⚠️ Lỗi tính toán toán học ở Zoom %d: %v\n", z, err)
			continue // Bỏ qua zoom này, chạy tiếp zoom khác
		}

		fmt.Printf("🗺️ [%s - Zoom %d] Cần tải %d ảnh...\n", areaName, z, len(tiles))

		for tile := range tiles {
			errTile := DownloadTile(int(tile.Z), int(tile.X), int(tile.Y))
			if errTile == nil {
				downloadedCount++
			}
			if downloadedCount%1000 == 0 && downloadedCount > 0 {
				fmt.Printf("⏳ [%s] Đã tải: %d ảnh...\n", areaName, downloadedCount)
			}
		}
	}
	fmt.Printf("✅ HOÀN TẤT [%s]! Tổng tải mới: %d.\n", areaName, downloadedCount)
}

func StartVietnamCrawler() {
	// Danh sách các mức zoom bạn muốn tải cho TOÀN BỘ Việt Nam.
	// Bao gồm zoom nhỏ (5-10) để nhìn toàn cảnh, và zoom lớn (11-14) để nhìn rõ đường.
	zooms := []int{5, 6, 7, 8, 9, 10, 11, 12, 13,14}

	go DownloadAreaSmart(vnBoundary, zooms, "Toàn Quốc (Full 5-13)")
}

// Dành cho Controller gọi để lấy ảnh ra khỏi ổ cứng
func GetTileFromDisk(z, x, y string) ([]byte, error) {
	filePath := fmt.Sprintf("./tiles/%s/%s/%s", z, x, y)
	return os.ReadFile(filePath)
}
