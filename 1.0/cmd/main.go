package main

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	paymentAdapters "github.com/roblaszczak/room-reservation-system/adapters/payment"
	reservationAdapters "github.com/roblaszczak/room-reservation-system/adapters/reservation"
	"github.com/roblaszczak/room-reservation-system/domain/payment"
	"github.com/roblaszczak/room-reservation-system/domain/reservation"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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

func main() {
	rand.Seed(time.Now().Unix())

	offerRepo := reservationAdapters.MockRoomOfferRepository{}
	bookingRepo := reservationAdapters.MemoryBookingRepository{}
	paymentsInitializer := paymentAdapters.MockInitializer{}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Post("/book-room", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		req := BookRoom{}
		if err := json.Unmarshal(b, &req); err != nil {
			panic(err)
		}
		req.BookingUUID = watermill.NewUUID()

		offer, err := offerRepo.GetRoomOffer(req.RoomID, req.StartTime, req.EndTime, req.GuestsCount)
		if err != nil {
			panic(err)
		}

		booking, err := reservation.NewBooking(watermill.NewUUID(), offer)
		if err == reservation.ErrRoomNotAvailable {
			log.Printf("Room is %s not available", req.RoomID)
			return
		} else if err != nil {
			panic(err)
		}

		p, err := payment.NewPayment(watermill.NewUUID(), booking.ID(), booking.Price(), req.PaymentChannel)
		if err != nil {
			panic(err)
		}

		if err := paymentsInitializer.InitializePayment(p); err != nil {
			panic(err)
		}

		if err := bookingRepo.AddBooking(booking); err != nil {
			panic(err)
		}

		log.Printf("%s room booked, booking %s", req.RoomID, req.BookingUUID)
	})

	if err := http.ListenAndServe(":6060", r); err != http.ErrServerClosed {
		panic(err)
	}
}
