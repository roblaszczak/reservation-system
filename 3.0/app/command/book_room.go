package command

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/roblaszczak/room-reservation-system/domain/reservation"
	"log"
	"time"
)

type BookRoom struct {
	BookingUUID string
	RoomID      string

	StartTime time.Time
	EndTime   time.Time

	GuestsCount int

	PaymentChannel string
}

type BookRoomHandler struct {
	eventBus *cqrs.EventBus

	offerRepo   reservation.RoomOfferRepository
	bookingRepo reservation.BookingRepository
}

func NewBookRoomHandler(
	eventBus *cqrs.EventBus,
	offerRepo reservation.RoomOfferRepository,
	bookingRepo reservation.BookingRepository,
) *BookRoomHandler {
	return &BookRoomHandler{eventBus, offerRepo, bookingRepo}
}

// NewCommand returns type of command which this handle should handle. It must be a pointer.
func (b BookRoomHandler) NewCommand() interface{} {
	return &BookRoom{}
}

func (b BookRoomHandler) Handle(c interface{}) error {
	cmd := c.(*BookRoom)

	offer, err := b.offerRepo.GetRoomOffer(cmd.RoomID, cmd.StartTime, cmd.EndTime, cmd.GuestsCount)
	if err != nil {
		return err
	}

	booking, err := reservation.NewBooking(cmd.BookingUUID, offer)
	if err == reservation.ErrRoomNotAvailable {
		log.Printf("Room is %s not available", cmd.RoomID)
		return nil
	} else if err != nil {
		return err
	}

	if err := b.bookingRepo.AddBooking(booking); err != nil {
		return err
	}

	if err := b.eventBus.Publish(&RoomBooked{
		BookingUUID:    booking.ID(),
		RoomID:         cmd.RoomID,
		Price:          booking.Price(),
		PaymentChannel: cmd.PaymentChannel,
	}); err != nil {
		return err
	}

	log.Printf("Booking %s done", cmd.BookingUUID)

	return nil
}

type RoomBooked struct {
	BookingUUID    string
	RoomID         string
	Price          int
	PaymentChannel string
}
