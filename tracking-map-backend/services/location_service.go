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

	//Kiểm tra xem tuyến đường này đã lưu trong Redis chưa?
	cachedRoute, err := s.Repo.Redis.Get(ctx, routeKey).Result()
	if err == nil {
		var result map[string]interface{}
		json.Unmarshal([]byte(cachedRoute), &result)
		fmt.Println("Trả về tuyến đường siêu tốc từ Redis Cache!")
		return result, nil
	}

	//chưa có trong Redis, gọi API của OSRM
	osrmURL := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s,%s;%s,%s?overview=full&geometries=geojson", startLon, startLat, endLon, endLat)
	// client := &http.Client{Timeout: 15 * time.Second}
	// resp, err := client.Get(osrmURL)
	// if err != nil {
	// 	return nil, err
	// }
	req, err := http.NewRequest("GET", osrmURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "TrackingMapApp/1.0 (StudentProject)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM API từ chối kết nối, mã lỗi: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Lưu kết quả OSRM vào Redis Cache (7 ngày)
	routeJSON, _ := json.Marshal(result)
	s.Repo.Redis.Set(ctx, routeKey, routeJSON, 7*24*time.Hour)
	fmt.Println("Gọi API OSRM và đã lưu vào Redis Cache!")

	return result, nil
}
