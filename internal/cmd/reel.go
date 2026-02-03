package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mahmoud/igpostercli/internal/config"
	"github.com/mahmoud/igpostercli/internal/graph"
	"github.com/mahmoud/igpostercli/internal/upload"
)

type ReelCmd struct {
	File    string `help:"Local video file" type:"existingfile"`
	URL     string `help:"Public HTTPS video URL (skip upload)"`
	Caption string `help:"Post caption" short:"c"`
}

func (c *ReelCmd) Run(root *RootFlags) error {
	if c.File == "" && c.URL == "" {
		return usage("provide --file or --url")
	}
	if c.File != "" && c.URL != "" {
		return usage("provide only one of --file or --url")
	}

	cfg, err := config.LoadWithProfile(root.Profile)
	if err != nil {
		return err
	}

	if root != nil && root.UserID != "" {
		cfg.IGUserID = root.UserID
	}

	err = cfg.ValidateForAccessToken()
	if err != nil {
		return err
	}

	ctx := context.Background()
	mediaURL := c.URL
	if mediaURL != "" {
		mediaURL, err = ensureHTTPS(mediaURL)
		if err != nil {
			return err
		}
	} else {
		mediaURL, err = upload.Upload(ctx, c.File)
		if err != nil {
			return err
		}
	}

	client := graph.NewClient(cfg)
	creationID, err := client.CreateReelContainer(ctx, mediaURL, c.Caption)
	if err != nil {
		return err
	}

	err = client.PollStatus(ctx, creationID, cfg.PollInterval, cfg.PollTimeout)
	if err != nil {
		return err
	}

	publishedID, err := client.Publish(ctx, creationID)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "PUBLISHED_MEDIA_ID=%s\n", publishedID)
	return nil
}
