package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

var projectID = "developer-certification-376713"

// var topicID = "testing"
var subID = "testing-order"

func process(buffer []string) error {
	fmt.Println(buffer)
	// return errors.New("testing")
	return nil
}

func batcher(ch <-chan *pubsub.Message) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	incomingMessage := []*pubsub.Message{}

	buffer := []string{}

	for {
		select {
		case msg := <-ch:
			incomingMessage = append(incomingMessage, msg)
			m := string(msg.Data)
			buffer = append(buffer, m)
		case <-ticker.C:
			fmt.Println("flushing")
			if len(buffer) == 0 {
				continue
			}

			err := process(buffer)
			if err != nil {
				continue
			} else {
				for _, m := range incomingMessage {
					m.Ack()
				}
			}

			buffer = []string{}
			incomingMessage = []*pubsub.Message{}
		}
	}

}

func PullMsgs(name string) error {
	// name = n
	// projectID := "my-project-id"
	// subID := "my-sub"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %w", err)
	}
	defer client.Close()

	// client.Subscriber can be passed a subscription ID (e.g. "my-sub") or
	// a fully qualified name (e.g. "projects/my-project/subscriptions/my-sub").
	// If a subscription ID is provided, the project ID from the client is used.
	sub := client.Subscriber(subID)

	// sub.ReceiveSettings.MaxOutstandingMessages = 1
	sub.ReceiveSettings.NumGoroutines = 1

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.

	ch := make(chan *pubsub.Message)
	go batcher(ch)

	start := time.Now()
	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()

	// messages := []string{}
	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		// fmt.Printf("Got message: %q, %s\n", string(msg.Data), name)
		// // messages = append(messages, string(msg.Data))
		// atomic.AddInt32(&received, 1)
		// msg.Ack()
		ch <- msg
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %w", err)
	}
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("Received %d messages\n", received)
	// fmt.Println(messages)
	fmt.Println("duration:", duration.Seconds())

	return nil
}
