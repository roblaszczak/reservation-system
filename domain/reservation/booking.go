package reservation

import (
	"github.com/pkg/errors"
)

type Booking struct {
	id string

	offer BookingOffer
}

var (
	ErrRoomNotAvailable = errors.New("Room not available")
)

func NewBooking(bookingUUID string, offer BookingOffer) (*Booking, error) {
	if bookingUUID == "" {
		return nil, errors.New("Empty bookingUUID")
	}
	if !offer.IsRoomAvailable() {
		return nil, ErrRoomNotAvailable
	}

	b := Booking{
		id:    bookingUUID,
		offer: offer,
	}

	return &b, nil
}

func (b Booking) ID() string {
	return b.id
}

func (b Booking) Price() int {
	return b.offer.price
}

type BookingRepository interface {
	AddBooking(booking *Booking) error
}
