package main

import (
	"app/cmd/agent/db"
	"context"
	"fmt"
	"time"
)

func StartAgent(ctx context.Context) {
	store, err := db.Connect("data/openlogs-agent.db")
	if err != nil {
		fmt.Printf("connect sqlite db: %v\n", err)
		return
	}
	defer store.Close()

	client := NewClient()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("ticked, sending batch")
			if err := collectSaveSend(ctx, store, client); err != nil {
				fmt.Println("collectSaveSend err:", err)
			}
		}
	}
}

func collectSaveSend(ctx context.Context, store *db.Conn, client *Client) error {
	batch, err := collectBatch()
	if err != nil {
		return fmt.Errorf("collect batch: %w", err)
	}

	if err := store.SaveBatch(ctx, batch); err != nil {
		return fmt.Errorf("save batch: %w", err)
	}

	if err := client.SendBatch(ctx, batch); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}
