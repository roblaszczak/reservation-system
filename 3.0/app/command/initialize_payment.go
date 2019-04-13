package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/roblaszczak/room-reservation-system/domain/payment"
)

type InitializePayment struct {
	BookingUUID    string
	RoomID         string
	Price          int
	PaymentChannel string
}

type InitializePaymentHandler struct {
	eventBus *cqrs.EventBus

	paymentInitializer payment.Initializer
}

func NewInitializePaymentHandler(
	eventBus *cqrs.EventBus,
	paymentInitializer payment.Initializer,
) *InitializePaymentHandler {
	return &InitializePaymentHandler{eventBus, paymentInitializer}
}

// NewCommand returns type of command which this handle should handle. It must be a pointer.
func (b InitializePaymentHandler) NewCommand() interface{} {
	return &InitializePayment{}
}

func (b InitializePaymentHandler) Handle(c interface{}) error {
	cmd := c.(*InitializePayment)

	p, err := payment.NewPayment(
		watermill.NewUUID(),
		cmd.BookingUUID,
		cmd.Price,
		cmd.PaymentChannel,
	)
	if err != nil {
		return err
	}

	if err := b.paymentInitializer.InitializePayment(p); err != nil {
		return err
	}

	if err := b.eventBus.Publish(&PaymentInitialized{
		BookingUUID: cmd.BookingUUID,
		Price:       cmd.Price,
	}); err != nil {
		return err
	}

	return nil
}

type PaymentInitialized struct {
	BookingUUID string
	Price       int
}
