package main

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/infrastructure/kafka"
	watermillMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
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

var watermillLogger = watermill.NewStdLogger(false, false)

type BookRoom struct {
	BookingUUID string
	RoomID      string

	StartTime time.Time
	EndTime   time.Time

	GuestsCount int

	PaymentChannel string
}

type RoomBooked struct {
	BookingUUID    string
	RoomID         string
	Price          int
	PaymentChannel string
}

func main() {
	rand.Seed(time.Now().Unix())

	offerRepo := reservationAdapters.MockRoomOfferRepository{}
	bookingRepo := reservationAdapters.MemoryBookingRepository{}

	publisher, err := kafka.NewPublisher(
		[]string{"localhost:9092"},
		kafka.DefaultMarshaler{},
		nil, // no custom sarama config
		watermillLogger,
	)
	if err != nil {
		panic(err)
	}

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

		if err := bookingRepo.AddBooking(booking); err != nil {
			panic(err)
		}

		event, err := json.Marshal(RoomBooked{
			BookingUUID:    booking.ID(),
			RoomID:         req.RoomID,
			Price:          booking.Price(),
			PaymentChannel: req.PaymentChannel,
		})
		if err != nil {
			panic(err)
		}
		if err := publisher.Publish("room_bookings", message.NewMessage(watermill.NewUUID(), event)); err != nil {
			panic(err)
		}

		log.Printf("%s room booked, booking %s", req.RoomID, req.BookingUUID)
	})

	go processMessages()

	if err := http.ListenAndServe(":6060", r); err != http.ErrServerClosed {
		panic(err)
	}
}

func processMessages() {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       []string{"localhost:9092"},
			ConsumerGroup: "test_consumer_group",
		},
		saramaSubscriberConfig,
		kafka.DefaultMarshaler{},
		watermillLogger,
	)
	if err != nil {
		panic(err)
	}

	watermillRouter, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}
	watermillRouter.AddMiddleware(watermillMiddleware.Recoverer)

	paymentsInitializer := paymentAdapters.MockInitializer{}
	watermillRouter.AddNoPublisherHandler(
		"BookRoom",
		"room_bookings",
		subscriber,
		func(msg *message.Message) ([]*message.Message, error) {
			event := RoomBooked{}
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				panic(err)
			}

			p, err := payment.NewPayment(watermill.NewUUID(), event.BookingUUID, event.Price, event.PaymentChannel)
			if err != nil {
				panic(err)
			}

			if err := paymentsInitializer.InitializePayment(p); err != nil {
				panic(err)
			}

			return nil, nil
		},
	)

	// todo - suppport close
	if err := watermillRouter.Run(); err != nil {
		panic(err)
	}
}
