package main

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type eqReceiptMetadata struct {
	TransactionID   string `json:"tx_id"`
	QuestionnaireID string `json:"questionnaire_id"`
}

type eqReceipt struct {
	TimeCreated string            `json:"timeCreated"`
	Metadata    eqReceiptMetadata `json:"metadata"`
}

type rmResponse struct {
	CaseID          *string `json:"caseId"` // Why a reference type?
	QuestionnaireID string  `json:"questionnaireId"`
	Unreceipt       bool    `json:"unreceipt"`
}

type rmPayload struct {
	Response rmResponse `json:"response"`
}

type rmEvent struct {
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Channel       string    `json:"channel"`
	DateTime      string    `json:"dateTime"`
	TransactionID string    `json:"transactionId"`
	Payload       rmPayload `json:"payload"`
}

type rmMessage struct {
	Event   rmEvent   `json:"event"`
	Payload rmPayload `json:"payload"`
}

func main() {
	c := make(chan eqReceipt)

	go pullMsgs("project", "rm-receipt-subscription", &c)

	go func(c *chan eqReceipt) {
		for {
			eqReceiptReceived := <-*c
			rmMessageToSend, err := convertEqReceiptToRmMessage(&eqReceiptReceived)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to convert receipt to message"))
			}
			sendRabbitMessage(rmMessageToSend)
		}
	}(&c)

	select {}
}

func pullMsgs(projectID, subID string, c *chan eqReceipt) {
	log.Println("Launched PubSub message listener")

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("pubsub.NewClient: %v\n", err)
	}

	// Consume message.

	// // Received and mutex is used to auto-exit the go routine on number of iterations
	// var mu sync.Mutex
	// received := 0

	sub := client.Subscription(subID)

	// Not using the cancel function here (which makes WithCancel a bit redundant!)
	// Ideally the cancel would be deferred/delegated to a signal watcher to
	// enable graceful shutdown that can kill the active go routines.
	cctx, _ := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))

		eqReceiptReceived := eqReceipt{}
		json.Unmarshal(msg.Data, &eqReceiptReceived)

		log.Printf("Got QID: %q\n", eqReceiptReceived.Metadata.QuestionnaireID)

		*c <- eqReceiptReceived

		msg.Ack()

		// // Only need to use the mutex if we're wanting to increment the exit
		// // counter. The counter is used to set a maximum number of messages to
		// // process before stopping the consumer.
		// mu.Lock()
		// defer mu.Unlock()
		// received++
		// if received == -999 { // Never quit
		// 	cancel()
		// }
	})
	if err != nil {
		log.Printf("Receive: %v\n", err)
	}
}

func sendRabbitMessage(message *rmMessage) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:6672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	byteMessage, err := json.Marshal(message)
	failOnError(err, "Failed to marshall data")

	err = ch.Publish(
		"",            // default exchange
		"goTestQueue", // routing key (the queue)
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        byteMessage,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", string(byteMessage))
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func convertEqReceiptToRmMessage(eqReceipt *eqReceipt) (*rmMessage, error) {
	if eqReceipt == nil {
		return nil, errors.New("receipt has nil content")
	}

	return &rmMessage{
		Event: rmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      eqReceipt.TimeCreated,
			TransactionID: eqReceipt.Metadata.TransactionID,
		},
		Payload: rmPayload{
			Response: rmResponse{
				QuestionnaireID: eqReceipt.Metadata.QuestionnaireID,
			},
		},
	}, nil
}
