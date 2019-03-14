package reservation

import (
	"github.com/roblaszczak/room-reservation-system/domain/reservation"
	"time"
)

type MockRoomOfferRepository struct{}

func (MockRoomOfferRepository) GetRoomOffer(
	roomID string,
	startTime time.Time,
	endTime time.Time,
	guestsCount int,
) (reservation.BookingOffer, error) {
	return reservation.NewAvailableBookingOffer(roomID, startTime, endTime, 100*guestsCount), nil
}
