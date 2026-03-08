package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// khởi tạo kết nối và tạo bảng 
func InitPostgres() *pgxpool.Pool {
	dsn := "postgres://user_tracking:password123@localhost:5432/tracking_map"
	dbPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic("Kết nối Database thất bại: " + err.Error())
	}

	// tạo bảng
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS device_locations (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(50) NOT NULL,
		latitude DOUBLE PRECISION NOT NULL,
		longitude DOUBLE PRECISION NOT NULL,
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_device_time ON device_locations (device_id, timestamp DESC);
	`
	_, err = dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		panic("Lỗi khi tự động tạo bảng: " + err.Error())
	}

	fmt.Println("Đã khởi tạo PostgreSQL thành công!")
	return dbPool
}