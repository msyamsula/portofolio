package main

import (
	pubsub "github.com/msyamsula/portofolio/pub-sub"
)

func main() {
	pubsub.PullMsgs("1")
	// go pubsub.PullMsgs("2")
}
