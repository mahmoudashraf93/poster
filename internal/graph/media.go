package graph

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (c *Client) CreatePhotoContainer(ctx context.Context, imageURL, caption string) (string, error) {
	params := map[string]string{
		"image_url": imageURL,
	}
	if caption != "" {
		params["caption"] = caption
	}

	resp, err := c.post(ctx, fmt.Sprintf("%s/media", c.igUserID), params)
	if err != nil {
		return "", err
	}

	return extractID(resp)
}

func (c *Client) CreateReelContainer(ctx context.Context, videoURL, caption string) (string, error) {
	params := map[string]string{
		"media_type": "REELS",
		"video_url":  videoURL,
	}
	if caption != "" {
		params["caption"] = caption
	}

	resp, err := c.post(ctx, fmt.Sprintf("%s/media", c.igUserID), params)
	if err != nil {
		return "", err
	}

	return extractID(resp)
}

func (c *Client) CreateCarouselChild(ctx context.Context, mediaURL string, isVideo bool) (string, error) {
	params := map[string]string{
		"is_carousel_item": "true",
	}
	if isVideo {
		params["video_url"] = mediaURL
	} else {
		params["image_url"] = mediaURL
	}

	resp, err := c.post(ctx, fmt.Sprintf("%s/media", c.igUserID), params)
	if err != nil {
		return "", err
	}

	return extractID(resp)
}

func (c *Client) CreateCarouselContainer(ctx context.Context, childIDs []string, caption string) (string, error) {
	params := map[string]string{
		"media_type": "CAROUSEL",
		"children":   strings.Join(childIDs, ","),
	}
	if caption != "" {
		params["caption"] = caption
	}

	resp, err := c.post(ctx, fmt.Sprintf("%s/media", c.igUserID), params)
	if err != nil {
		return "", err
	}

	return extractID(resp)
}

func (c *Client) PollStatus(ctx context.Context, creationID string, interval, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("%w after %s", ErrPollTimeout, timeout)
		}

		resp, err := c.get(ctx, creationID, map[string]string{"fields": "status_code"})
		if err != nil {
			return err
		}

		status, ok := resp["status_code"].(string)
		if ok {
			switch status {
			case "FINISHED":
				return nil
			case "ERROR":
				return ErrMediaProcessing
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("poll canceled: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
}

func (c *Client) Publish(ctx context.Context, creationID string) (string, error) {
	resp, err := c.post(ctx, fmt.Sprintf("%s/media_publish", c.igUserID), map[string]string{
		"creation_id": creationID,
	})
	if err != nil {
		return "", err
	}

	return extractID(resp)
}

func extractID(resp JSON) (string, error) {
	value, ok := resp["id"]
	if !ok {
		return "", ErrMissingID
	}

	switch typed := value.(type) {
	case string:
		if typed == "" {
			return "", ErrEmptyID
		}

		return typed, nil
	case float64:
		return fmt.Sprintf("%.0f", typed), nil
	default:
		return "", ErrUnexpectedIDType
	}
}
