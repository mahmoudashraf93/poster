package main

import (
	"os"

	"github.com/mahmoud/igpostercli/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
