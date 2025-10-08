package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/okieraised/rclgo/humble"
	_ "github.com/okieraised/rclgo/humble"
)

func main() {
	err := humble.Init(nil)
	if err != nil {
		log.Default().Fatal(fmt.Sprintf("Failed to initialize ROS client: %v", err))
		return
	}
	defer func() {
		cErr := humble.Deinit()
		if cErr != nil && err == nil {
			err = cErr
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := humble.NewNode("test", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(node *humble.Node) {
		cErr := node.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}(node)

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

	typesByTopic, err := node.GetTopicNamesAndTypes(true)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println(err)
			return
		case <-ticker.C:
			typesByTopic, err = node.GetTopicNamesAndTypes(true)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(typesByTopic)
		}
	}
}
