package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
)

func Publish(msg string, publisher *pubsub.Publisher, ctx context.Context) error {
	// projectID := "my-project-id"
	// topicID := "my-topic"
	// msg := "Hello World"

	// client.Publisher can be passed a topic ID (e.g. "my-topic") or
	// a fully qualified name (e.g. "projects/my-project/topics/my-topic").
	// If a topic ID is provided, the project ID from the client is used.
	// Reuse this publisher for all publish calls to send messages in batches.
	result := publisher.Publish(ctx, &pubsub.Message{
		Data:        []byte(msg),
		OrderingKey: "order",
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	fmt.Printf("Published a message; msg ID: %v\n", id)
	return nil
}
