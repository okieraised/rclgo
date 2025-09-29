package main

import (
	"context"
	"fmt"
	"time"

	"github.com/okieraised/rclgo/pkg/rclgo"
)

func main() {
	err := rclgo.Init(nil)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to initialize rclgo: %v", err))
		return
	}
	defer func() {
		cErr := rclgo.Deinit()
		if cErr != nil && err == nil {
			err = cErr
		}
	}()

	node, err := rclgo.NewNode("publisher", "")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to create node: %v", err))
		return
	}
	defer func(node *rclgo.Node) {
		cErr := node.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}(node)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()

	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	default:
		fmt.Println(node.GetTopicNamesAndTypes(true))
		time.Sleep(2 * time.Second)
	}
}
