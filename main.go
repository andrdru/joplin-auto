package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/andrdru/joplin-auto/cmd/app"
)

type (
	flags struct {
		isHelp     *bool
		configPath *string
	}
)

const (
	// serviceName name of service
	// redefine here or with ldflags
	// go build -ldflags="-X 'main.serviceName=my_service'"
	serviceName = "joplin_auto"
)

func main() {
	f := initFlags()
	if *f.isHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	initLogger()

	code := app.Run(slog.Default(), *f.configPath)

	os.Exit(code)
}

func initFlags() (fv flags) {
	fv.isHelp = flag.Bool("help", false, "Print help and exit")
	fv.configPath = flag.String("config", "config.yaml", "path to config.yml")

	flag.Parse()
	return fv
}

func initLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		"service",
		serviceName,
	)

	slog.SetDefault(logger)
}
