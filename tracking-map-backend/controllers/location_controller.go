package controllers

import (
	"net/http"
	"time"

	"tracking-map-backend/models" 
	"tracking-map-backend/repositories"
	"tracking-map-backend/services"

	"github.com/gin-gonic/gin"
)

type LocationController struct {
	Repo    *repositories.LocationRepo
	Service *services.LocationService
}

// Vị trí mới nhất của thiết bị sẽ được lưu vào redis và postgreSQL, 
func (ctrl *LocationController) PostLocation(c *gin.Context) {
	var loc models.Location
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}
	if err := ctrl.Repo.SaveLocation(c.Request.Context(), loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu dữ liệu"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Lấy vị trí mới nhất của thiết bị
func (ctrl *LocationController) GetLatest(c *gin.Context) {
	loc, err := ctrl.Repo.GetLatestLocation(c.Request.Context(), c.Param("device_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy dữ liệu"})
		return
	}
	c.JSON(http.StatusOK, loc)
}

// lấy ra lịch sử di chuyển của thiết bị
func (ctrl *LocationController) GetHistory(c *gin.Context) {
	history, err := ctrl.Repo.GetHistory(c.Request.Context(), c.Param("device_id"), c.Query("from_time"), c.Query("to_time"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy xuất lịch sử"})
		return
	}
	if history == nil {
		history = []models.Location{}
	}
	c.JSON(http.StatusOK, history)
}


// Lấy danh sách thiết bị đã từng gửi vị trí 
func (ctrl *LocationController) GetDeviceList(c *gin.Context) {
	devices, err := ctrl.Repo.GetDeviceList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn danh sách"})
		return
	}
	if devices == nil {
		devices = []string{}
	}
	c.JSON(http.StatusOK, devices)
}

// dẫn đường giữa 2 điểm
func (ctrl *LocationController) GetRoute(c *gin.Context) {
	route, err := ctrl.Service.GetRoute(c.Request.Context(), c.Query("start_lon"), c.Query("start_lat"), c.Query("end_lon"), c.Query("end_lat"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dịch vụ dẫn đường đang lỗi"})
		return
	}
	c.JSON(http.StatusOK, route)
}