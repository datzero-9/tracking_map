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

	repo := &repositories.LocationRepo{DB: dbPool, Redis: redisClient}
	service := &services.LocationService{Repo: repo}
	controller := &controllers.LocationController{Repo: repo, Service: service}

	r := gin.Default()
	r.Use(cors.Default())

	// 1. Nạp file biên giới vào RAM để Crawler sử dụng
	err := services.LoadVnBoundary()
	if err != nil {
		panic("Không thể khởi động vì thiếu vn.json: " + err.Error())
	}
	
	// 2. Chạy ngầm con Robot cào bản đồ
	go services.DownloadVietnamBBox()

	// 3. CÁC ĐƯỜNG DẪN TĨNH (Để ngoài API Group để không bị Rate Limit chặn)
	r.Static("/tiles", "./tiles")
	
	// 👉 MỞ CỔNG CHO FRONTEND LẤY FILE BIÊN GIỚI (Ý tưởng tuyệt vời của bạn)
	r.StaticFile("/api/boundary", "./vn.json") 

	// 4. CÁC API CỐT LÕI (Bảo vệ bằng Rate Limiter Redis)
	api := r.Group("/api/v1")
	api.Use(middlewares.RateLimiter(redisClient))
	{
		api.GET("/devices", controller.GetDeviceList)
		
		// 🚨 Tôi đã sửa lỗi đánh máy chữ "laocations" thành "locations" cho bạn nhé!
		api.POST("/locations", controller.PostLocation) 
		api.GET("/devices/:device_id/latest", controller.GetLatest)
		api.GET("/devices/:device_id/history", controller.GetHistory)
		api.GET("/route", controller.GetRoute)
	}

	// Khởi động server
	r.Run(":8080")
}