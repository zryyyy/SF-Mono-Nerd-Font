package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/app"
)

func main() {
	configPath := flag.String("config", "font.json", "path to font config")
	dryRun := flag.Bool("dry-run", false, "validate config and required commands without downloading or patching")
	flag.Parse()

	if err := app.Run(*configPath, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
