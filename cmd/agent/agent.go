package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func StartAgent(ctx context.Context) {
	client := NewClient()

	cmdC := client.Connect(ctx)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// send current state back to server
		case cmd := <-cmdC:
			fmt.Fprintf(os.Stderr, "received command: %+v\n", cmd)

			batch, err := collectBatch()
			if err != nil {
				fmt.Fprintf(os.Stderr, "collect batch: %v\n", err)
				continue
			}

			if err := client.SendBatch(ctx, batch); err != nil {
				fmt.Fprintf(os.Stderr, "send batch: %v\n", err)
			}
		}
	}
}
