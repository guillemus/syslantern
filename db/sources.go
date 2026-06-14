package db

import (
	"context"
	"database/sql"
)

type UpsertSourceWebsitePageParams struct {
	SourceWebsiteID int64
	URL             string
	Title           string
	Content         string
	ByteCount       int64
}

func (c *Conn) UpsertSourceWebsite(ctx context.Context, workspaceID int64, rootURL string) (SourceWebsite, error) {
	row, err := c.GetSourceWebsiteQuery(ctx, GetSourceWebsiteQueryParams{WorkspaceID: workspaceID, RootURL: rootURL})
	if err == nil {
		return row.SourceWebsite, nil
	}
	if err != sql.ErrNoRows {
		return SourceWebsite{}, err
	}

	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return SourceWebsite{}, err
	}
	defer tx.Rollback()
	qtx := c.WithTx(tx)

	website, err := qtx.CreateSourceWebsiteQuery(ctx, CreateSourceWebsiteQueryParams{
		WorkspaceID: workspaceID,
		RootURL:     rootURL,
	})
	if err != nil {
		return SourceWebsite{}, err
	}

	return website, tx.Commit()
}

func (c *Conn) UpsertSourceWebsitePage(ctx context.Context, page UpsertSourceWebsitePageParams) error {
	return c.UpsertSourceWebsitePageQuery(ctx, UpsertSourceWebsitePageQueryParams{
		SourceWebsiteID: page.SourceWebsiteID,
		URL:             page.URL,
		Title:           page.Title,
		Content:         page.Content,
		ByteCount:       page.ByteCount,
	})
}
