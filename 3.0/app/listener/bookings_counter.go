package listener

import "github.com/roblaszczak/room-reservation-system/3.0/app/command"

type BookingsCounterIncrementer interface {
	IncrementBookingCounter(bookingUUID string) error
}

type BookingsCounterGenerator struct {
	incrementer BookingsCounterIncrementer
}

func NewBookingsCounterGenerator(incrementer BookingsCounterIncrementer) *BookingsCounterGenerator {
	return &BookingsCounterGenerator{incrementer}
}

func (BookingsCounterGenerator) NewEvent() interface{} {
	return &command.RoomBooked{}
}

func (b BookingsCounterGenerator) Handle(e interface{}) error {
	event := e.(*command.RoomBooked)

	return b.incrementer.IncrementBookingCounter(event.BookingUUID)
}
