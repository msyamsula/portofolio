package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	pb "github.com/msyamsula/portofolio/learning/pub-sub"
)

var projectID = "developer-certification-376713"
var topicID = "testing"

func main() {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return
	}
	defer client.Close()

	publisher := client.Publisher(topicID)
	publisher.EnableMessageOrdering = true

	i := 0
	for {
		// time.Sleep(20 * time.Millisecond)
		msg := fmt.Sprintf("mantap %d", i)
		fmt.Println(msg)
		pb.Publish(fmt.Sprintf("mantap %d", i), publisher, ctx)
		// var wg sync.WaitGroup
		// for j := 0; j < 10; j++ {
		// 	wg.Add(1)
		// 	go func(idx int) {
		// 		defer wg.Done()
		// 	}(j)
		// }
		// wg.Wait()
		i++
	}
}
