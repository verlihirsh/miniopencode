package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"opencode-tty/internal/config"
	"opencode-tty/internal/proxy"
	"opencode-tty/internal/tui"
)

func main() {
	headless := flag.Bool("headless", false, "run in headless stdin/stdout mode")
	configPath := flag.String("config", "", "path to config file (default: ~/.config/miniopencode.yaml)")
	flag.Parse()

	cfg, err := config.Load(*configPath, config.Options{})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if *headless {
		p := proxy.NewProxy(proxy.Config{Host: cfg.Server.Host, Port: fmt.Sprintf("%d", cfg.Server.Port)})
		p.RunHeadless()
		return
	}

	ctx := context.Background()
	if err := tui.Run(ctx, cfg); err != nil {
		log.Fatalf("tui: %v", err)
	}
}
