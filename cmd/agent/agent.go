package main

import (
	agentdb "app/cmd/agent/db"
	"context"
	"fmt"
	"time"
)

func StartAgent(ctx context.Context) {
	store, err := agentdb.Connect("data/openlogs-agent.db")
	if err != nil {
		fmt.Printf("connect sqlite db: %v\n", err)
		return
	}
	defer store.Close()

	client := NewClient()

	cmdC := client.Connect(ctx)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("ticked, sending batch")
			collectSaveSend(ctx, store, client)
		case cmd := <-cmdC:
			fmt.Printf("received command: %+v\n", cmd)
			collectSaveSend(ctx, store, client)
		}
	}
}

func collectSaveSend(ctx context.Context, store *agentdb.Conn, client *Client) {
	batch, err := collectBatch()
	if err != nil {
		fmt.Printf("collect batch: %v\n", err)
		return
	}

	if err := store.SaveBatch(ctx, batch); err != nil {
		fmt.Printf("save batch: %v\n", err)
		return
	}

	if err := client.SendBatch(ctx, batch); err != nil {
		fmt.Printf("send batch: %v\n", err)
	}
}
