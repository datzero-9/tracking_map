package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

type Location struct {
	DeviceID  string    `json:"device_id" binding:"required"`
	Latitude  float64   `json:"latitude" binding:"required"`
	Longitude float64   `json:"longitude" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	dsn := "postgres://user_tracking:password123@localhost:5432/tracking_map"
	var err error
	dbPool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic("Kết nối Database thất bại! Vui lòng kiểm tra Docker hoặc PostgreSQL.")
	}
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
		panic("Lỗi khi tự động tạo bảng dữ liệu: " + err.Error())
	}
	fmt.Println("Đã kiểm tra và khởi tạo Database thành công!")
	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api/v1")
	{

		api.GET("/devices", getDeviceList)                 // Lấy danh sách xe
		api.POST("/locations", postLocation)               //  Nhận dữ liệu vị trí GPS
		api.GET("/devices/:device_id/latest", getLatest)   //  Lấy vị trí mới nhất
		api.GET("/devices/:device_id/history", getHistory) // Lấy lịch sử di chuyển
		api.GET("/route", getRoute)                        // Dẫn đường thực tế
	}

	r.Run(":8080")
}

// Lấy danh sách thiết bị
func getDeviceList(c *gin.Context) {
	query := `SELECT DISTINCT device_id FROM device_locations ORDER BY device_id`
	rows, err := dbPool.Query(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn danh sách"})
		return
	}
	defer rows.Close()

	var devices []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			devices = append(devices, id)
		}
	}

	if devices == nil {
		devices = []string{}
	}
	c.JSON(http.StatusOK, devices)
}

// Lưu vị trí mới
func postLocation(c *gin.Context) {
	var loc Location
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}

	query := `INSERT INTO device_locations (device_id, latitude, longitude, timestamp) VALUES ($1, $2, $3, $4)`
	_, err := dbPool.Exec(context.Background(), query, loc.DeviceID, loc.Latitude, loc.Longitude, loc.Timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu vào Database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Lấy vị trí mới nhất
func getLatest(c *gin.Context) {
	deviceID := c.Param("device_id")
	query := `SELECT device_id, latitude, longitude, timestamp FROM device_locations 
              WHERE device_id = $1 ORDER BY timestamp DESC LIMIT 1`
	var loc Location
	err := dbPool.QueryRow(context.Background(), query, deviceID).Scan(&loc.DeviceID, &loc.Latitude, &loc.Longitude, &loc.Timestamp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy dữ liệu"})
		return
	}
	c.JSON(http.StatusOK, loc)
}

// Lịch sử di chuyển
func getHistory(c *gin.Context) {
	deviceID := c.Param("device_id")
	from := c.Query("from_time")
	to := c.Query("to_time")

	query := `SELECT device_id, latitude, longitude, timestamp FROM device_locations 
              WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp ASC`

	rows, err := dbPool.Query(context.Background(), query, deviceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy xuất lịch sử"})
		return
	}
	defer rows.Close()

	history := []Location{}
	for rows.Next() {
		var l Location
		rows.Scan(&l.DeviceID, &l.Latitude, &l.Longitude, &l.Timestamp)
		history = append(history, l)
	}
	c.JSON(http.StatusOK, history)
}

// Lấy đường vẽ OSRM
func getRoute(c *gin.Context) {
	osrmURL := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s,%s;%s,%s?overview=full&geometries=geojson",
		c.Query("start_lon"), c.Query("start_lat"), c.Query("end_lon"), c.Query("end_lat"))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(osrmURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dịch vụ dẫn đường đang lỗi"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(http.StatusOK, result)
}
