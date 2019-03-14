package payment

import "github.com/pkg/errors"

type Payment struct {
	id string

	bookingID    string
	bookingPrice int

	channel string
}

func (p Payment) ID() string {
	return p.id
}

func (p Payment) BookingID() string {
	return p.bookingID
}

func NewPayment(id string, bookingID string, bookingPrice int, channel string) (Payment, error) {
	if id == "" {
		return Payment{}, errors.New("payment ID is empty")
	}
	if bookingID == "" {
		return Payment{}, errors.New("booking ID is empty")
	}

	if channel == "" {
		return Payment{}, errors.New("empty payment channel")
	}
	if bookingPrice <= 0 {
		return Payment{}, errors.New("bookingPrice is less or equal to 0")
	}

	return Payment{id, bookingID, bookingPrice, channel}, nil
}

type Initializer interface {
	InitializePayment(Payment) error
}
