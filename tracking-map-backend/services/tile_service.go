package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/maptile/tilecover"
)

var vnBoundary orb.Geometry

// LoadVnBoundary nạp file biên giới Việt Nam
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

// DownloadTile thực hiện tải một ô gạch và kiểm tra tính toàn vẹn (Phiên bản Siêu Bền Bỉ)
func DownloadTile(z, x, y int) error {
	// 1. TỐI ƯU URL: Xoay vòng 4 server mt0, mt1, mt2, mt3 để không bị Google chặn IP
	subdomain := (x + y) % 4
	url := fmt.Sprintf("http://mt%d.google.com/vt/lyrs=m&x=%d&y=%d&z=%d", subdomain, x, y, z)

	dir := fmt.Sprintf("tiles/%d/%d", z, x)
	os.MkdirAll(dir, os.ModePerm)
	filePath := filepath.Join(dir, fmt.Sprintf("%d.png", y))

	// 2. KIỂM TRA FILE CŨ: Nếu file tồn tại và dung lượng > 500 byte thì tự động bỏ qua (Skip nhanh)
	info, err := os.Stat(filePath)
	if err == nil && info.Size() > 500 {
		return nil
	}

	// Xóa file rác (0 byte hoặc quá nhỏ) để tải lại bản chuẩn
	if err == nil {
		os.Remove(filePath)
	}

	// 3. CƠ CHẾ THỬ LẠI (RETRY): Thử tải tối đa 3 lần nếu mạng chập chờn
	var resp *http.Response
	var lastErr error

	for i := 0; i < 3; i++ {
		client := &http.Client{Timeout: 15 * time.Second}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, lastErr = client.Do(req)
		
		// Nếu tải thành công (Mã 200 OK), thoát khỏi vòng lặp thử lại
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	// Nếu sau 3 lần vẫn thất bại thì báo lỗi để hàm cha biết
	if lastErr != nil || resp == nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("thất bại sau 3 lần thử (Z:%d X:%d Y:%d)", z, x, y)
	}
	defer resp.Body.Close()

	// 4. LƯU FILE ẢNH VÀO Ổ CỨNG
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadAreaSmart tải đa luồng để tăng tốc và tránh bỏ sót
// DownloadAreaSmart tải đa luồng và TỰ ĐỘNG LẤY THÊM TILE XUNG QUANH (Padding)
func DownloadAreaSmart(geometry orb.Geometry, zooms []int, areaName string) {
	fmt.Printf("🚀 BẮT ĐẦU CHIẾN DỊCH TỔNG LỰC (Có mở rộng viền): %s...\n", areaName)

	const maxWorkers = 15
	sem := make(chan bool, maxWorkers)
	var wg sync.WaitGroup

	for _, z := range zooms {
		// 1. Lấy danh sách Tile gốc từ vn.json
		originalTiles, err := tilecover.Geometry(geometry, maptile.Zoom(z))
		if err != nil {
			fmt.Printf("⚠️ Lỗi tính toán Zoom %d: %v\n", z, err)
			continue
		}

		// 2. THỰC HIỆN Ý TƯỞNG CỦA BẠN: Lấy thêm 1 Tile xung quanh (8 hướng)
		expandedTiles := make(map[maptile.Tile]struct{})
		for t := range originalTiles {
			// Giữ lại tile gốc
			expandedTiles[t] = struct{}{}
			
			// Lấy thêm 8 tile xung quanh (Trên, Dưới, Trái, Phải, và 4 góc chéo)
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					// Bỏ qua chính nó (vì đã thêm ở trên)
					if dx == 0 && dy == 0 {
						continue
					}
					neighbor := maptile.Tile{
						X: uint32(int(t.X) + dx),
						Y: uint32(int(t.Y) + dy),
						Z: t.Z,
					}
					// Map trong Go tự động loại bỏ các tile trùng lặp
					expandedTiles[neighbor] = struct{}{}
				}
			}
		}

		total := len(expandedTiles)
		fmt.Printf("🗺️ [%s - Zoom %d] Đã mở rộng viền. Tổng cộng: %d ảnh\n", areaName, z, total)

		count := 0
		// 3. Tiến hành tải danh sách Tile ĐÃ MỞ RỘNG
		for t := range expandedTiles {
			wg.Add(1)
			sem <- true

			go func(tile maptile.Tile) {
				defer wg.Done()
				defer func() { <-sem }()

				DownloadTile(int(tile.Z), int(tile.X), int(tile.Y))
			}(t)

			count++
			if count%5000 == 0 {
				fmt.Printf("⏳ [%s-Z%d] Tiến độ: %d/%d (%.2f%%)\n", areaName, z, count, total, float64(count)/float64(total)*100)
			}
		}
		wg.Wait()
	}
	fmt.Printf("✅ CHIẾN DỊCH HOÀN TẤT: %s\n", areaName)
}

// StartVietnamCrawler khởi chạy robot
func StartVietnamCrawler() {
	// Robot sẽ tự động lướt qua cực nhanh những ảnh đã tải và chỉ dừng lại "vá" những ảnh thiếu/lỗi
	zooms := []int{5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	go DownloadAreaSmart(vnBoundary, zooms, "Vietnam_Full_Map")
}

// GetTileFromDisk trả về file ảnh cho Controller
func GetTileFromDisk(z, x, y string) ([]byte, error) {
	filePath := fmt.Sprintf("./tiles/%s/%s/%s", z, x, y)
	return os.ReadFile(filePath)
}