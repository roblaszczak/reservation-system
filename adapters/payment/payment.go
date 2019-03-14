package payment

import (
	"github.com/pkg/errors"
	"github.com/roblaszczak/room-reservation-system/domain/payment"
	"log"
	"math/rand"
	"time"
)

type MockInitializer struct {
}

func (MockInitializer) InitializePayment(p payment.Payment) error {
	start := time.Now()

	if rand.Intn(5) == 0 {
		// simulating not working payments provider
		return errors.New("payment provider is not working")
	}
	if rand.Intn(5) == 0 {
		// simulating slow payments provider
		time.Sleep(time.Second)
	}

	log.Printf("Payment %s for booking %s initialized in %s", p.ID(), p.BookingID(), time.Now().Sub(start))

	return nil
}
