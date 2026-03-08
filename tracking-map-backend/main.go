package main

import (
	"fmt"

	"tracking-map-backend/controllers" 
	"tracking-map-backend/database"    
	"tracking-map-backend/middlewares"
	"tracking-map-backend/repositories"
	"tracking-map-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Kết nối Database & Redis  
	dbPool := database.InitPostgres()
	defer dbPool.Close() 

	redisClient := database.InitRedis()
	defer redisClient.Close() 

	fmt.Println("Hệ thống tracking map sẵn sàng!")

	// khởi tạo
	repo := &repositories.LocationRepo{DB: dbPool, Redis: redisClient}
	service := &services.LocationService{Repo: repo}
	controller := &controllers.LocationController{Repo: repo, Service: service}

	// Cấu hình Router 
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(middlewares.RateLimiter(redisClient))
	// 1. Chạy ngầm việc tải bản đồ khu vực Đà Nẵng (Tọa độ 16.0544, 108.2022)
	// Dùng từ khóa "go" ở đầu để nó chạy ngầm, không làm treo server
	go services.DownloadMapArea(108.2022, 16.0544)

	// 2. Mở cửa thư mục "tiles" ra Internet để Frontend có thể lấy ảnh về vẽ
	// Đây chính là mấu chốt để bạn tự làm Tile Server của riêng mình!
	r.Static("/tiles", "./tiles")
	api := r.Group("/api/v1")
	{
		api.GET("/devices", controller.GetDeviceList)
		api.POST("/locations", controller.PostLocation)
		api.GET("/devices/:device_id/latest", controller.GetLatest)
		api.GET("/devices/:device_id/history", controller.GetHistory)
		api.GET("/route", controller.GetRoute)
	}

	r.Run(":8080")
}
