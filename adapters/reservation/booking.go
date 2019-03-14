package reservation

import (
	"github.com/roblaszczak/room-reservation-system/domain/reservation"
	"log"
)

type MemoryBookingRepository struct {
}

func (MemoryBookingRepository) AddBooking(booking *reservation.Booking) error {
	log.Printf("Saved booking %s", booking.ID())
	return nil
}
