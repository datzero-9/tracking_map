package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"math"
)

// tải mảnh ảnh từ Google và lưu vào thư mục
func DownloadTile(z, x, y int) error {
	
	url := fmt.Sprintf("http://mt1.google.com/vt/lyrs=m&x=%d&y=%d&z=%d", x, y, z)
	dir := fmt.Sprintf("tiles/%d/%d", z, x)
	os.MkdirAll(dir, os.ModePerm)
	filePath := filepath.Join(dir, fmt.Sprintf("%d.png", y))

	//Kiểm tra nếu file đã tồn tại thì không
	if _, err := os.Stat(filePath); err == nil {
		return nil 
	}

	//tải ảnh
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lỗi khi tải tile: %v", err)
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

// 2. Tải toàn bộ khu vực xung quanh 1 điểm
func DownloadMapArea(centerLon, centerLat float64) {
	zooms := []int{13, 14, 15} //3 mức zoom
	radius := 10 
	fmt.Println("Đang tải bản đồ Google Offline...")
	for _, z := range zooms {
		centerX, centerY := LonLatToTile(centerLon, centerLat, z)

		for x := centerX - radius; x <= centerX + radius; x++ {
			for y := centerY - radius; y <= centerY + radius; y++ {
				err := DownloadTile(z, x, y)
				if err != nil {
					fmt.Printf("Lỗi tải tile z:%d x:%d y:%d\n", z, x, y)
				}
			}
		}
	}
	fmt.Println("Hoàn tất tải bản đồ Offline (Lưu tại thư mục /tiles)!")
}