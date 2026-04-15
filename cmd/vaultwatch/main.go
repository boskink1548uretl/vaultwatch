package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/youorg/vaultwatch/internal/alert"
	"github.com/youorg/vaultwatch/internal/config"
	"github.com/youorg/vaultwatch/internal/monitor"
	"github.com/youorg/vaultwatch/internal/scheduler"
	"github.com/youorg/vaultwatch/internal/vault"
)

func main() {
	configPath := flag.String("config", "configs/vaultwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	vaultClient, err := vault.NewClient(cfg.Vault.Address, cfg.Vault.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create vault client: %v\n", err)
		os.Exit(1)
	}

	checker := monitor.NewChecker(cfg.Alerts.WarningThreshold, cfg.Alerts.CriticalThreshold)
	notifier := alert.NewStdoutNotifier()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("shutting down vaultwatch...")
		cancel()
	}()

	taskFn := func(ctx context.Context) error {
		for _, path := range cfg.Secrets {
			meta, err := vaultClient.GetSecretMetadata(ctx, path)
			if err != nil {
				log.Printf("error fetching metadata for %s: %v", path, err)
				continue
			}
			result := checker.Evaluate(path, meta.ExpiresAt)
			if n := alert.FromCheckResult(result); n != nil {
				if err := notifier.Notify(ctx, *n); err != nil {
					log.Printf("notify error for %s: %v", path, err)
				}
			}
		}
		return nil
	}

	sched := scheduler.New(cfg.Scheduler.Interval, taskFn)
	if err := sched.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("scheduler error: %v", err)
	}
}
