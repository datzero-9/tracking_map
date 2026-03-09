package models

import "time"
//Cấu trúc DB
type Location struct {
	DeviceID  string    `json:"device_id" binding:"required"`
	Latitude  float64   `json:"latitude" binding:"required"`
	Longitude float64   `json:"longitude" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}