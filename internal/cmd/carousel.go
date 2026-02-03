package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mahmoud/igpostercli/internal/config"
	"github.com/mahmoud/igpostercli/internal/graph"
	"github.com/mahmoud/igpostercli/internal/upload"
)

type CarouselCmd struct {
	Files   []string `help:"Local media files" type:"existingfile"`
	Caption string   `help:"Post caption" short:"c"`
}

func (c *CarouselCmd) Run(root *RootFlags) error {
	if len(c.Files) == 0 {
		return usage("provide at least one --files entry")
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
	client := graph.NewClient(cfg)
	childIDs := make([]string, 0, len(c.Files))

	for _, file := range c.Files {
		var isVideo bool
		isVideo, err = detectMediaType(file)
		if err != nil {
			return err
		}

		var mediaURL string
		mediaURL, err = upload.Upload(ctx, file)
		if err != nil {
			return err
		}

		var childID string
		childID, err = client.CreateCarouselChild(ctx, mediaURL, isVideo)
		if err != nil {
			return err
		}

		if isVideo {
			err = client.PollStatus(ctx, childID, cfg.PollInterval, cfg.PollTimeout)
			if err != nil {
				return err
			}
		}

		childIDs = append(childIDs, childID)
	}

	creationID, err := client.CreateCarouselContainer(ctx, childIDs, c.Caption)
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

	_, _ = fmt.Fprintf(os.Stdout, "CHILD_IDS=%s\n", strings.Join(childIDs, ","))
	_, _ = fmt.Fprintf(os.Stdout, "PUBLISHED_MEDIA_ID=%s\n", publishedID)
	return nil
}
