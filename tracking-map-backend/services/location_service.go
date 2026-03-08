package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tracking-map-backend/repositories"

)

type LocationService struct {
	Repo *repositories.LocationRepo
}

func (s *LocationService) GetRoute(ctx context.Context, startLon, startLat, endLon, endLat string) (map[string]interface{}, error) {
	routeKey := fmt.Sprintf("route:%s,%s_%s,%s", startLon, startLat, endLon, endLat)
	
	// 1. Kiểm tra xem tuyến đường này đã từng được hỏi và lưu trong Redis chưa?
	cachedRoute, err := s.Repo.Redis.Get(ctx, routeKey).Result()
	if err == nil {
		var result map[string]interface{}
		json.Unmarshal([]byte(cachedRoute), &result)
		fmt.Println("⚡ Trả về tuyến đường siêu tốc từ Redis Cache!")
		return result, nil
	}

	// 2. Nếu chưa có trong Redis, gọi API của OSRM
	osrmURL := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s,%s;%s,%s?overview=full&geometries=geojson", startLon, startLat, endLon, endLat)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(osrmURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// 3. Lưu kết quả OSRM vào Redis Cache (Hạn sử dụng 7 ngày vì đường xá ít khi thay đổi)
	routeJSON, _ := json.Marshal(result)
	s.Repo.Redis.Set(ctx, routeKey, routeJSON, 7*24*time.Hour)
	fmt.Println("🌐 Gọi API OSRM và đã lưu vào Redis Cache!")

	return result, nil
}