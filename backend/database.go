package main

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global DB variable
var db *gorm.DB

// Initialize database connection and migrations
func initDatabase(dsn string) error {
	var err error
	
	// Connect to the database
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	
	// Auto migrate the schema
	err = db.AutoMigrate(&Event{}, &SeatSection{}, &Seat{}, &Booking{})
	if err != nil {
		return err
	}
	
	// Check if we need to seed the database
	var count int64
	db.Model(&Event{}).Count(&count)
	if count == 0 {
		// Create sample data
		if err := seedDatabase(); err != nil {
			return err
		}
	}
	
	return nil
}

// Seed the database with sample data
func seedDatabase() error {
	// Create an event
	event := Event{
		Name:        "Ultimate Music Experience",
		Description: "Experience an unforgettable night of music with the world's top performers.",
		Date:        time.Date(2025, 3, 25, 19, 0, 0, 0, time.Local),
		Venue:       "Grand Arena, Downtown",
		Duration:    180,
	}
	if err := db.Create(&event).Error; err != nil {
		return err
	}
	
	// Create seat sections
	sections := []SeatSection{
		{EventID: event.ID, Name: "VIP", IsVIP: true},
		{EventID: event.ID, Name: "Premium", IsVIP: false},
		{EventID: event.ID, Name: "Standard", IsVIP: false},
	}
	
	for i := range sections {
		if err := db.Create(&sections[i]).Error; err != nil {
			return err
		}
	}
	
	// Create VIP seats (2 rows, 10 seats per row)
	for row := 1; row <= 2; row++ {
		rowLetter := string(rune('A' - 1 + row))
		for seat := 1; seat <= 10; seat++ {
			seatObj := Seat{
				EventID:   event.ID,
				SectionID: sections[0].ID,
				Row:       rowLetter,
				Number:    seat,
				SeatCode:  fmt.Sprintf("%s%d", rowLetter, seat),
				Status:    "available",
				Price:     200.00,
			}
			if err := db.Create(&seatObj).Error; err != nil {
				return err
			}
		}
	}
	
	// Create Premium seats (5 rows, 15 seats per row)
	for row := 1; row <= 5; row++ {
		rowLetter := string(rune('B' + row))
		for seat := 1; seat <= 15; seat++ {
			seatObj := Seat{
				EventID:   event.ID,
				SectionID: sections[1].ID,
				Row:       rowLetter,
				Number:    seat,
				SeatCode:  fmt.Sprintf("%s%d", rowLetter, seat),
				Status:    "available",
				Price:     100.00,
			}
			if err := db.Create(&seatObj).Error; err != nil {
				return err
			}
		}
	}
	
	// Create Standard seats (10 rows, 15 seats per row)
	for row := 1; row <= 10; row++ {
		rowLetter := string(rune('G' + row))
		for seat := 1; seat <= 15; seat++ {
			seatObj := Seat{
				EventID:   event.ID,
				SectionID: sections[2].ID,
				Row:       rowLetter,
				Number:    seat,
				SeatCode:  fmt.Sprintf("%s%d", rowLetter, seat),
				Status:    "available",
				Price:     50.00,
			}
			if err := db.Create(&seatObj).Error; err != nil {
				return err
			}
		}
	}
	
	fmt.Println("Database seeded successfully!")
	return nil
}