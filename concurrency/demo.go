package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/streadway/amqp"
)

type eqReceiptMetadata struct {
	TransactionId   string `json:"tx_id"`
	QuestionnaireId string `json:"questionnaire_id"`
}

type eqReceipt struct {
	TimeCreated string            `json:"timeCreated"`
	Metadata    eqReceiptMetadata `json:"metadata"`
}

type rmResponse struct {
	CaseId          string `json:"caseId"`
	QuestionnaireId string `json:"questionnaireId"`
	Unreceipt       bool   `json:"unreceipt"`
}

type rmPayload struct {
	Response rmResponse `json:"response"`
}

type rmEvent struct {
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Channel       string    `json:"channel"`
	DateTime      string    `json:"dateTime"`
	TransactionId string    `json:"transactionId"`
	Payload       rmPayload `json:"payload"`
}

type rmMessage struct {
	Event   rmEvent   `json:"event"`
	Payload rmPayload `json:"payload"`
}

func main() {
	c := make(chan eqReceipt)

	go pullMsgs("project", "rm-receipt-subscription", &c)
	go convertAndSend(&c)

	// Infinite loop
	select {}
}

func pullMsgs(projectID, subID string, c *chan eqReceipt) {
	fmt.Println("Launched PubSub message listener")

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fmt.Printf("pubsub.NewClient: %v\n", err)
	}

	// Consume message.
	var mu sync.Mutex
	received := 0
	sub := client.Subscription(subID)
	cctx, cancel := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Printf("Got message: %q\n", string(msg.Data))

		eqReceiptReceived := eqReceipt{}
		json.Unmarshal(msg.Data, &eqReceiptReceived)

		fmt.Printf("Got QID: %q\n", eqReceiptReceived.Metadata.QuestionnaireId)

		*c <- eqReceiptReceived

		msg.Ack()
		mu.Lock()
		defer mu.Unlock()
		received++
		if received == -999 { // Never quit
			cancel()
		}
	})
	if err != nil {
		fmt.Printf("Receive: %v\n", err)
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

	fmt.Printf(" [x] Sent %s", string(byteMessage))
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

func convertEqReceiptToRmMessage(eqReceipt *eqReceipt) *rmMessage {
	if eqReceipt == nil {
		return nil
	}

	messageToSendToRm := &rmMessage{
		rmEvent{Type: "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      eqReceipt.TimeCreated,
			TransactionId: eqReceipt.Metadata.TransactionId},
		rmPayload{
			rmResponse{
				CaseId:          "",
				QuestionnaireId: "",
				Unreceipt:       false}},
	}

	messageToSendToRm.Payload = rmPayload{
		Response: rmResponse{
			QuestionnaireId: eqReceipt.Metadata.QuestionnaireId,
		},
	}

	return messageToSendToRm
}

func convertAndSend(c *chan eqReceipt) {
	fmt.Println("Launched message converter/sender")

	for {
		eqReceiptReceived := <-*c
		rmMessageToSend := convertEqReceiptToRmMessage(&eqReceiptReceived)
		sendRabbitMessage(rmMessageToSend)
	}
}
