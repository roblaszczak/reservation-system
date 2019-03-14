package listener

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/roblaszczak/room-reservation-system/3.0/app/command"
)

type InitializePaymentOnRoomBooked struct {
	commandBus *cqrs.CommandBus
}

func NewInitializePaymentOnRoomBooked(commandBus *cqrs.CommandBus) *InitializePaymentOnRoomBooked {
	return &InitializePaymentOnRoomBooked{commandBus}
}

func (InitializePaymentOnRoomBooked) NewEvent() interface{} {
	return &command.RoomBooked{}
}

func (o InitializePaymentOnRoomBooked) Handle(e interface{}) error {
	event := e.(*command.RoomBooked)

	orderBeerCmd := &command.InitializePayment{
		BookingUUID:    event.BookingUUID,
		RoomID:         event.RoomID,
		Price:          event.Price,
		PaymentChannel: event.PaymentChannel,
	}

	return o.commandBus.Send(orderBeerCmd)
}
