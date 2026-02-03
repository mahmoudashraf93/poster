package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mahmoud/igpostercli/internal/config"
	"github.com/mahmoud/igpostercli/internal/graph"
)

type TokenCmd struct {
	Exchange ExchangeCmd `cmd:"" help:"Exchange short-lived token"`
	Debug    DebugCmd    `cmd:"" help:"Debug access token"`
}

type ExchangeCmd struct {
	ShortToken string `help:"Short-lived user token" required:""`
}

func (c *ExchangeCmd) Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	err = cfg.ValidateForTokenExchange()
	if err != nil {
		return err
	}
	if c.ShortToken == "" {
		return usage("provide --short-token")
	}

	ctx := context.Background()
	longToken, err := graph.ExchangeToken(ctx, cfg.AppID, cfg.AppSecret, c.ShortToken)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "ACCESS_TOKEN=%s\n", longToken)
	return nil
}

type DebugCmd struct {
	Token string `help:"Token to debug (defaults to IG_ACCESS_TOKEN)"`
}

func (c *DebugCmd) Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	token := c.Token
	if token == "" {
		err = cfg.ValidateForTokenDebug()
		if err != nil {
			return err
		}
		token = cfg.AccessToken
	}

	ctx := context.Background()
	info, err := graph.DebugToken(ctx, token)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "IS_VALID=%t\n", info.IsValid)
	_, _ = fmt.Fprintf(os.Stdout, "APP_ID=%s\n", info.AppID)
	_, _ = fmt.Fprintf(os.Stdout, "APPLICATION=%s\n", info.Application)
	_, _ = fmt.Fprintf(os.Stdout, "TYPE=%s\n", info.Type)
	_, _ = fmt.Fprintf(os.Stdout, "EXPIRES_AT=%d\n", info.ExpiresAt)
	_, _ = fmt.Fprintf(os.Stdout, "DATA_ACCESS_EXPIRES_AT=%d\n", info.DataAccessExpiresAt)
	_, _ = fmt.Fprintf(os.Stdout, "USER_ID=%s\n", info.UserID)
	return nil
}
