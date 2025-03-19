package main

import (
	"time"

	"gorm.io/gorm"
)

// GORM Models
type Event struct {
	gorm.Model
	Name        string
	Description string
	Date        time.Time
	Venue       string
	Duration    int
	Sections    []SeatSection `gorm:"foreignKey:EventID"`
}

type SeatSection struct {
	gorm.Model
	EventID uint
	Name    string
	IsVIP   bool
	Seats   []Seat `gorm:"foreignKey:SectionID"`
}

type Seat struct {
	gorm.Model
	EventID    uint
	SectionID  uint
	Row        string
	Number     int
	SeatCode   string
	Status     string `gorm:"default:available"`
	Price      float64
	Section    SeatSection `gorm:"foreignKey:SectionID"`
}

type Booking struct {
	gorm.Model
	BookingReference string `gorm:"unique"`
	CustomerName     string
	CustomerEmail    string
	CustomerPhone    string
	TicketType       string
	TotalAmount      float64
	BookingTime      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Seats            []Seat    `gorm:"many2many:booking_seats"`
}

// Request/Response structs
type BookingRequest struct {
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	CustomerPhone string `json:"customerPhone"`
	TicketType    string `json:"ticketType"`
	SeatIDs       []uint `json:"seatIds"`
}