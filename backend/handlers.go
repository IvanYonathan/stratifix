package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections in development
	},
}

// Connected WebSocket clients
var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// Get all events
func getEvents(c echo.Context) error {
	var events []Event
	if err := db.Find(&events).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch events")
	}
	
	return c.JSON(http.StatusOK, events)
}

// Get seats for an event
func getSeats(c echo.Context) error {
	eventID := c.Param("eventId")
	
	// Get the event
	var event Event
	if err := db.First(&event, eventID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}
	
	// Get all seats for this event
	var seats []Seat
	if err := db.Preload("Section").Where("event_id = ?", eventID).Find(&seats).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch seats")
	}
	
	// Count available seats
	var availableSeats int64
	db.Model(&Seat{}).Where("event_id = ? AND status = ?", eventID, "available").Count(&availableSeats)
	
	// Format the response
	seatData := make([]map[string]interface{}, len(seats))
	bookedSeats := make(map[uint]bool)
	
	for i, seat := range seats {
		seatData[i] = map[string]interface{}{
			"id":        seat.ID,
			"sectionId": seat.SectionID,
			"section":   seat.Section.Name,
			"row":       seat.Row,
			"number":    seat.Number,
			"seatCode":  seat.SeatCode,
			"status":    seat.Status,
			"price":     seat.Price,
			"isVip":     seat.Section.IsVIP,
		}
		
		if seat.Status == "booked" {
			bookedSeats[seat.ID] = true
		}
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"totalSeats":     len(seats),
		"availableSeats": availableSeats,
		"seats":          seatData,
		"bookedSeats":    bookedSeats,
	})
}

// Create a booking
func createBooking(c echo.Context) error {
	var req BookingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid booking data")
	}
	
	// Start a transaction
	tx := db.Begin()
	
	// Check if seats are available
	var unavailableSeats []string
	for _, seatID := range req.SeatIDs {
		var seat Seat
		if err := tx.Where("id = ? AND status = ?", seatID, "available").First(&seat).Error; err != nil {
			var seatCode string
			tx.Model(&Seat{}).Where("id = ?", seatID).Pluck("seat_code", &seatCode)
			unavailableSeats = append(unavailableSeats, seatCode)
		}
	}
	
	if len(unavailableSeats) > 0 {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Seats %v are already booked", unavailableSeats))
	}
	
	// Calculate total price
	var totalAmount float64
	tx.Model(&Seat{}).Where("id IN ?", req.SeatIDs).Pluck("SUM(price)", &totalAmount)
	
	// Create booking reference
	bookingRef := fmt.Sprintf("TKT-%d", time.Now().UnixNano()%100000)
	
	// Create booking
	booking := Booking{
		BookingReference: bookingRef,
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerPhone:    req.CustomerPhone,
		TicketType:       req.TicketType,
		TotalAmount:      totalAmount,
		BookingTime:      time.Now(),
	}
	
	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create booking")
	}
	
	// Update seat status and add to booking
	for _, seatID := range req.SeatIDs {
		if err := tx.Model(&Seat{}).Where("id = ?", seatID).Update("status", "booked").Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update seat status")
		}
		
		// Add association to booking_seats join table
		if err := tx.Exec("INSERT INTO booking_seats (booking_id, seat_id) VALUES (?, ?)", booking.ID, seatID).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to link seat to booking")
		}
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}
	
	// Get seat codes for display
	var seatCodes []string
	db.Model(&Seat{}).Where("id IN ?", req.SeatIDs).Pluck("seat_code", &seatCodes)
	
	// Notify other clients via WebSocket
	go broadcastSeatUpdate(req.SeatIDs)
	
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":      "Booking created successfully",
		"bookingId":    bookingRef,
		"customerName": req.CustomerName,
		"ticketType":   req.TicketType,
		"seats":        seatCodes,
		"totalAmount":  totalAmount,
	})
}

// Handle WebSocket connections
func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	
	// Register client
	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()
	
	// Keep connection alive
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			break
		}
	}
	
	return nil
}

// Broadcast seat update to all connected clients
func broadcastSeatUpdate(seatIDs []uint) {
	if len(seatIDs) == 0 {
		return
	}
	
	update := map[string]interface{}{
		"type":    "seat_update",
		"seatIds": seatIDs,
	}
	
	clientsMu.Lock()
	for client := range clients {
		err := client.WriteJSON(update)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
	clientsMu.Unlock()
}