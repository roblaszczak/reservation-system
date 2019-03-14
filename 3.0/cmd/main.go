package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/infrastructure/kafka"
	watermillMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"github.com/roblaszczak/room-reservation-system/3.0/app"
	"github.com/roblaszczak/room-reservation-system/3.0/app/command"
	"github.com/roblaszczak/room-reservation-system/3.0/app/listener"
	"github.com/roblaszczak/room-reservation-system/adapters/counter"
	"github.com/roblaszczak/room-reservation-system/adapters/payment"
	reservationAdapters "github.com/roblaszczak/room-reservation-system/adapters/reservation"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var watermillLogger = watermill.NewStdLogger(false, false)

func createSaramaSubscriber(handlerName string) (message.Subscriber, error) {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       []string{"localhost:9092"},
			ConsumerGroup: handlerName,
		},
		saramaSubscriberConfig,
		kafka.DefaultMarshaler{},
		watermillLogger,
	)
}

func main() {
	rand.Seed(time.Now().Unix())

	offerRepo := reservationAdapters.MockRoomOfferRepository{}
	bookingRepo := reservationAdapters.MemoryBookingRepository{}
	paymentsInitializer := payment.MockInitializer{}
	bookingsCounter := &counter.MemoryBookings{}

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

	watermillRouter, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	poisonQueue, err := watermillMiddleware.NewPoisonQueueWithFilter(
		publisher,
		"poison_queue",
		func(err error) bool {
			switch errors.Cause(err).(type) {
			case *json.InvalidUnmarshalError:
				return true
			default:
			}

			if errors.Cause(err) == app.ErrInvalidCommand {
				return true
			}
			if errors.Cause(err) == app.ErrInvalidEvent {
				return true
			}

			return false
		},
	)

	if err != nil {
		panic(err)
	}
	watermillRouter.AddMiddleware(
		watermillMiddleware.CorrelationIDWithAutogenerate(func() string {
			return watermill.NewShortUUID()
		}),
		poisonQueue.Middleware,
		watermillMiddleware.Retry{MaxRetries: 1, Logger: watermillLogger}.Middleware,
	)

	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		CommandsTopic: "commands",
		CommandHandlers: func(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.NewBookRoomHandler(eventBus, offerRepo, bookingRepo),
				command.NewInitializePaymentHandler(eventBus, paymentsInitializer),
			}
		},
		CommandsPublisher:             publisher,
		CommandsSubscriberConstructor: createSaramaSubscriber,
		EventsTopic:                   "events",
		EventHandlers: func(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				listener.NewInitializePaymentOnRoomBooked(commandBus),
				listener.NewBookingsCounterGenerator(bookingsCounter),
			}
		},
		EventsPublisher:             publisher,
		EventsSubscriberConstructor: createSaramaSubscriber,
		Router:                      watermillRouter,
		Logger:                      watermillLogger,
		CommandEventMarshaler:       cqrs.JSONMarshaler{},
	})
	if err != nil {
		panic(err)
	}

	r.Post("/book-room", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		cmd := command.BookRoom{}
		if err := json.Unmarshal(b, &cmd); err != nil {
			panic(err)
		}
		cmd.BookingUUID = watermill.NewUUID()
		log.Println("Starting booking process", cmd.BookingUUID)

		if err := cqrsFacade.CommandBus().Send(cmd); err != nil {
			panic(err)
		}
	})

	// todo -
	go watermillRouter.Run()

	go func() {
		for range time.Tick(time.Second * 5) {
			fmt.Printf("bookings done: %d\n", bookingsCounter.Count())
		}

	}()

	http.ListenAndServe(":6060", r)

	watermillRouter.Close()
}
