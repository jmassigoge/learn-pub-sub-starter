package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jmassigoge/learn-pub-sub-starter/internal/gamelogic"
	"github.com/jmassigoge/learn-pub-sub-starter/internal/pubsub"
	"github.com/jmassigoge/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")
	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("could not create and bind queue: %v", err)
	}
	publishChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		if userInput[0] == "pause" {
			log.Println("Sending pause message")
			err = pubsub.PublishJSON(publishChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
			continue
		}
		if userInput[0] == "resume" {
			log.Println("Sending resume message")
			err = pubsub.PublishJSON(publishChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
			continue
		}
		if userInput[0] == "quit" {
			break
		}
		log.Printf("Did not understand command")
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

}
