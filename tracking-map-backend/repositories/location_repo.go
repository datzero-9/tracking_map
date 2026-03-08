package repositories

import (
	"context"
	"encoding/json"
	"time"

	"tracking-map-backend/models" // Chú ý: Thay "tracking_map" bằng tên module trong file go.mod của bạn

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type LocationRepo struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

// Lưu vào PostgreSQL và Update luôn vào Redis Cache (Latest Location)
func (r *LocationRepo) SaveLocation(ctx context.Context, loc models.Location) error {
	query := `INSERT INTO device_locations (device_id, latitude, longitude, timestamp) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(ctx, query, loc.DeviceID, loc.Latitude, loc.Longitude, loc.Timestamp)
	if err == nil {
		// Cache lại vị trí mới nhất vào Redis, hết hạn sau 24h
		locJSON, _ := json.Marshal(loc)
		r.Redis.Set(ctx, "latest_loc:"+loc.DeviceID, locJSON, 24*time.Hour)
	}
	return err
}

// Lấy vị trí mới nhất (Ưu tiên đọc từ Redis trước, nếu không có mới chọc xuống DB)
func (r *LocationRepo) GetLatestLocation(ctx context.Context, deviceID string) (models.Location, error) {
	var loc models.Location

	// 1. Tìm trong Redis trước (Cực nhanh O(1))
	val, err := r.Redis.Get(ctx, "latest_loc:"+deviceID).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &loc)
		return loc, nil
	}

	// 2. Cache miss -> Lấy từ PostgreSQL
	query := `SELECT device_id, latitude, longitude, timestamp FROM device_locations WHERE device_id = $1 ORDER BY timestamp DESC LIMIT 1`
	err = r.DB.QueryRow(ctx, query, deviceID).Scan(&loc.DeviceID, &loc.Latitude, &loc.Longitude, &loc.Timestamp)
	
	// Cập nhật lại Cache để lần sau lấy cho nhanh
	if err == nil {
		locJSON, _ := json.Marshal(loc)
		r.Redis.Set(ctx, "latest_loc:"+deviceID, locJSON, 24*time.Hour)
	}
	return loc, err
}

func (r *LocationRepo) GetHistory(ctx context.Context, deviceID, from, to string) ([]models.Location, error) {
	query := `SELECT device_id, latitude, longitude, timestamp FROM device_locations WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp ASC`
	rows, err := r.DB.Query(ctx, query, deviceID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.DeviceID, &l.Latitude, &l.Longitude, &l.Timestamp)
		history = append(history, l)
	}
	return history, nil
}

func (r *LocationRepo) GetDeviceList(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT device_id FROM device_locations ORDER BY device_id`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			devices = append(devices, id)
		}
	}
	return devices, nil
}