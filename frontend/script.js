// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Generate the seating layout
    generateSeatingLayout();
    
    // Set up event listeners
    setupEventListeners();
    
    // Start the countdown timer
    startCountdown();
    
    // Connect to WebSocket for real-time updates
    connectWebSocket();
});

// Connect to the WebSocket for real-time updates
function connectWebSocket() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
    
    console.log(`Connecting to WebSocket at ${wsUrl}`);
    const socket = new WebSocket(wsUrl);
    
    socket.onopen = function() {
        console.log('WebSocket connection established');
    };
    
    socket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        console.log('WebSocket message received:', data);
        
        if (data.type === 'initial_data') {
            // Handle initial seat data
            updateSeatsFromServerData(data.data);
        } else if (data.type === 'seat_update') {
            // Handle seat updates
            updateSeatsStatus(data.seats);
        }
    };
    
    socket.onclose = function() {
        console.log('WebSocket connection closed');
        // Try to reconnect after a delay
        setTimeout(connectWebSocket, 3000);
    };
    
    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}

// Update seats based on server data
function updateSeatsFromServerData(serverData) {
    if (serverData && serverData.bookedSeats) {
        const bookedSeats = serverData.bookedSeats;
        
        // Update each seat based on server data
        document.querySelectorAll('.seat').forEach(seat => {
            const seatId = seat.dataset.id;
            if (bookedSeats[seatId]) {
                seat.classList.remove('available', 'selected', 'vip');
                seat.classList.add('booked');
                seat.dataset.status = 'booked';
            }
        });
        
        updateAvailableSeatsCount();
    }
}

// Update specific seats to booked status
function updateSeatsStatus(seatIds) {
    if (!seatIds || !Array.isArray(seatIds)) return;
    
    seatIds.forEach(seatId => {
        const seat = document.querySelector(`.seat[data-id="${seatId}"]`);
        if (seat) {
            // Add a visual effect for newly booked seats
            seat.style.backgroundColor = '#ff9800';
            
            setTimeout(() => {
                seat.classList.remove('available', 'selected', 'vip');
                seat.classList.add('booked');
                seat.dataset.status = 'booked';
                seat.style.backgroundColor = '';
                
                // If this was a selected seat, update the summary
                if (seat.classList.contains('selected')) {
                    updateSummary();
                }
            }, 1000);
        }
    });
    
    updateAvailableSeatsCount();
}

// Generate the seating layout
function generateSeatingLayout() {
    const seatingGrid = document.getElementById('seatingGrid');
    seatingGrid.innerHTML = '';
    
    // Define seat sections and their properties
    const sections = [
        { name: 'VIP', rows: 2, seatsPerRow: 10, isVIP: true },
        { name: 'AISLE', isAisle: true },
        { name: 'Premium', rows: 5, seatsPerRow: 15, isVIP: false },
        { name: 'AISLE', isAisle: true },
        { name: 'Standard', rows: 10, seatsPerRow: 15, isVIP: false }
    ];
    
    // Generate seats based on sections
    let seatId = 1;
    
    sections.forEach(section => {
        if (section.isAisle) {
            const aisle = document.createElement('div');
            aisle.className = 'aisle';
            seatingGrid.appendChild(aisle);
            return;
        }
        
        const sectionRow = document.createElement('div');
        sectionRow.className = 'section-row';
        sectionRow.textContent = section.name + ' Section';
        seatingGrid.appendChild(sectionRow);
        
        for (let row = 1; row <= section.rows; row++) {
            const rowLetter = String.fromCharCode(64 + row); // Convert number to letter (1=A, 2=B, etc)
            
            for (let seat = 1; seat <= section.seatsPerRow; seat++) {
                const seatElement = document.createElement('div');
                const seatCode = `${rowLetter}${seat}`;
                seatElement.className = `seat ${section.isVIP ? 'vip' : 'available'}`;
                seatElement.dataset.id = seatId;
                seatElement.dataset.section = section.name;
                seatElement.dataset.code = seatCode;
                seatElement.textContent = seatCode;
                seatElement.dataset.status = 'available';
                
                seatingGrid.appendChild(seatElement);
                seatId++;
            }
        }
    });
    
    // After generating the seating layout, fetch current seat status from the server
    fetchSeatData();
}

// Fetch seat data from the server
function fetchSeatData() {
    fetch('/api/seats')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            console.log('Fetched seat data:', data);
            updateSeatsFromServerData(data);
        })
        .catch(error => {
            console.error('Error fetching seat data:', error);
        });
}

// Set up event listeners
function setupEventListeners() {
    // Ticket type selection
    const ticketTypes = document.querySelectorAll('.ticket-type');
    ticketTypes.forEach(type => {
        type.addEventListener('click', function() {
            ticketTypes.forEach(t => t.classList.remove('selected'));
            this.classList.add('selected');
            updateSummary();
        });
    });
    
    // Seat selection
    document.getElementById('seatingGrid').addEventListener('click', function(e) {
        if (e.target.classList.contains('seat') && !e.target.classList.contains('booked')) {
            e.target.classList.toggle('selected');
            updateSummary();
        }
    });
    
    // Form submission
    document.getElementById('bookingForm').addEventListener('submit', function(e) {
        e.preventDefault();
        processBooking();
    });
    
    // Modal close buttons
    document.getElementById('closeSuccessModal').addEventListener('click', function() {
        document.getElementById('successModal').style.display = 'none';
    });
    
    // Download ticket button (just a placeholder in this demo)
    document.getElementById('downloadTicketBtn').addEventListener('click', function() {
        alert('Ticket download functionality would be implemented here.');
    });
}

// Update the booking summary
function updateSummary() {
    const selectedTicketType = document.querySelector('.ticket-type.selected');
    const selectedSeats = document.querySelectorAll('.seat.selected');
    
    // Update summary elements
    const summaryTicketType = document.getElementById('summaryTicketType');
    const summarySeats = document.getElementById('summarySeats');
    const summarySelectedSeats = document.getElementById('summarySelectedSeats');
    const summaryPrice = document.getElementById('summaryPrice');
    const summaryFee = document.getElementById('summaryFee');
    const summaryTotal = document.getElementById('summaryTotal');
    
    if (selectedTicketType) {
        const ticketPrice = parseInt(selectedTicketType.dataset.price);
        const numSeats = selectedSeats.length;
        const subtotal = ticketPrice * numSeats;
        const bookingFee = numSeats * 5; // $5 booking fee per seat
        const total = subtotal + bookingFee;
        
        // Collect selected seat codes
        const seatCodes = Array.from(selectedSeats).map(seat => seat.dataset.code);
        
        summaryTicketType.textContent = selectedTicketType.querySelector('h4').textContent;
        summarySeats.textContent = numSeats;
        summarySelectedSeats.textContent = seatCodes.join(', ') || '-';
        summaryPrice.textContent = `$${subtotal}`;
        summaryFee.textContent = `$${bookingFee}`;
        summaryTotal.textContent = `$${total}`;
    } else {
        summaryTicketType.textContent = '-';
        summarySeats.textContent = '0';
        summarySelectedSeats.textContent = '-';
        summaryPrice.textContent = '$0';
        summaryFee.textContent = '$0';
        summaryTotal.textContent = '$0';
    }
}

// Process the booking with the server
function processBooking() {
    const selectedTicketType = document.querySelector('.ticket-type.selected');
    const selectedSeats = document.querySelectorAll('.seat.selected');
    const fullName = document.getElementById('fullName').value;
    const email = document.getElementById('email').value;
    const phone = document.getElementById('phone').value;
    
    // Basic validation
    if (!selectedTicketType) {
        showError('Please select a ticket type');
        return;
    }
    
    if (selectedSeats.length === 0) {
        showError('Please select at least one seat');
        return;
    }
    
    if (!fullName || !email || !phone) {
        showError('Please fill in all personal information fields');
        return;
    }
    
    // Collect the booking data
    const ticketType = selectedTicketType.dataset.type;
    const seatIds = Array.from(selectedSeats).map(seat => seat.dataset.id);
    const seatCodes = Array.from(selectedSeats).map(seat => seat.dataset.code);
    
    // Create the booking data object
    const bookingData = {
        customerName: fullName,
        customerEmail: email,
        customerPhone: phone,
        ticketType: ticketType,
        seatIds: seatIds
    };
    
    // Send the booking data to the server
    fetch('/api/bookings', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(bookingData)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        console.log('Booking success:', data);
        
        // Mark the selected seats as booked locally (the server will broadcast this via WebSocket)
        selectedSeats.forEach(seat => {
            seat.classList.remove('selected');
            seat.classList.add('booked');
            seat.dataset.status = 'booked';
        });
        
        // Update ticket details in the success modal
        document.getElementById('ticketName').textContent = fullName;
        document.getElementById('ticketType').textContent = selectedTicketType.querySelector('h4').textContent;
        document.getElementById('ticketSeats').textContent = seatCodes.join(', ');
        document.getElementById('ticketOrderId').textContent = data.bookingId || 'TKT-12345';
        
        // Show success modal
        document.getElementById('successModal').style.display = 'block';
        
        // Reset form and update summary
        document.getElementById('bookingForm').reset();
        document.querySelectorAll('.ticket-type').forEach(t => t.classList.remove('selected'));
        updateSummary();
        updateAvailableSeatsCount();
    })
    .catch(error => {
        console.error('Error creating booking:', error);
        showError('There was an error processing your booking. Please try again.');
    });
}

// Show error message
function showError(message) {
    const errorMessage = document.getElementById('errorMessage');
    errorMessage.textContent = message;
    errorMessage.style.display = 'block';
    
    // Hide after 5 seconds
    setTimeout(() => {
        errorMessage.style.display = 'none';
    }, 5000);
}

// Start countdown timer to event
function startCountdown() {
    // Set the event date (demo: 7 days from now)
    const eventDate = new Date();
    eventDate.setDate(eventDate.getDate() + 7);
    
    function updateCountdown() {
        const now = new Date();
        const diff = eventDate - now;
        
        if (diff <= 0) {
            document.getElementById('days').textContent = '00';
            document.getElementById('hours').textContent = '00';
            document.getElementById('minutes').textContent = '00';
            document.getElementById('seconds').textContent = '00';
            return;
        }
        
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((diff % (1000 * 60)) / 1000);
        
        document.getElementById('days').textContent = days.toString().padStart(2, '0');
        document.getElementById('hours').textContent = hours.toString().padStart(2, '0');
        document.getElementById('minutes').textContent = minutes.toString().padStart(2, '0');
        document.getElementById('seconds').textContent = seconds.toString().padStart(2, '0');
    }
    
    // Update every second
    updateCountdown();
    setInterval(updateCountdown, 1000);
}

// Update the count of available seats
function updateAvailableSeatsCount() {
    const totalSeats = document.querySelectorAll('.seat').length;
    const bookedSeats = document.querySelectorAll('.seat.booked').length;
    const availableSeats = totalSeats - bookedSeats;
    
    document.getElementById('availableSeatsCount').textContent = availableSeats;
}