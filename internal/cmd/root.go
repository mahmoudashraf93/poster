package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/mahmoud/igpostercli/internal/errfmt"
)

type RootFlags struct {
	Verbose    bool   `help:"Verbose logging" short:"v"`
	UserID     string `help:"Instagram user ID (overrides IG_USER_ID)"`
	PageID     string `help:"Facebook Page ID (overrides IG_PAGE_ID)"`
	BusinessID string `help:"Business ID (overrides IG_BUSINESS_ID)"`
}

type CLI struct {
	RootFlags `embed:""`

	Version  kong.VersionFlag `help:"Print version"`
	Photo    PhotoCmd         `cmd:"" help:"Post a photo"`
	Reel     ReelCmd          `cmd:"" help:"Post a reel"`
	Carousel CarouselCmd      `cmd:"" help:"Post a carousel"`
	Token    TokenCmd         `cmd:"" help:"Token management"`
	Account  AccountCmd       `cmd:"" help:"Account utilities"`
	Owned    OwnedPagesCmd    `cmd:"" name:"owned-pages" help:"List pages owned by a business"`
}

type exitPanic struct{ code int }

func Execute() int {
	if err := executeArgs(os.Args[1:]); err != nil {
		return ExitCode(err)
	}
	return 0
}

func executeArgs(args []string) (err error) {
	parser, cli, err := newParser()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil
					return
				}
				err = &ExitError{Code: ep.code, Err: errors.New("exited")}
				return
			}
			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(parsedErr))
		return parsedErr
	}

	logLevel := slog.LevelWarn
	if cli.Verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	kctx.Bind(&cli.RootFlags)

	err = kctx.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(err))
		return err
	}

	return nil
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: 2, Err: parseErr}
	}
	return err
}

func newParser() (*kong.Kong, *CLI, error) {
	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("igpost"),
		kong.Description("Instagram Graph API posting CLI"),
		kong.Vars{"version": VersionString()},
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return nil, nil, err
	}
	return parser, cli, nil
}
