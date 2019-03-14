package reservation

import (
	"time"
)

// todo - use better types for prices/uuids/dates

type BookingOffer struct {
	roomID      string
	isAvailable bool
	startTime   time.Time
	endTime     time.Time
	price       int
}

func NewAvailableBookingOffer(
	roomID string,
	startTime time.Time,
	endTime time.Time,
	price int,
) BookingOffer {
	return BookingOffer{
		roomID:      roomID,
		isAvailable: true,
		startTime:   startTime,
		endTime:     endTime,
		price:       price,
	}
}

func NewNotAvailableBookingOffer(
	roomID string,
	startTime time.Time,
	endTime time.Time,
) BookingOffer {
	return BookingOffer{
		roomID:      roomID,
		isAvailable: false,
		startTime:   startTime,
		endTime:     endTime,
	}
}

func (r BookingOffer) IsRoomAvailable() bool {
	return r.isAvailable
}

type RoomOfferRepository interface {
	GetRoomOffer(roomID string, startTime time.Time, endTime time.Time, guestsCount int) (BookingOffer, error)
}
