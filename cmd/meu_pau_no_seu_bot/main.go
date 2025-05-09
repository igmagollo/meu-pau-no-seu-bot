package main

import (
	"context"
	"flag"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/igmagollo/meu-pau-no-seu-bot/internal"
	"github.com/joho/godotenv"
)

var (
	Name    = "meu-pau-no-seu-bot"
	Version string

	configPath = flag.String("config", "", "path to config file")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if err := godotenv.Load(); err != nil {
		logger.Printf("failed to load .env file: %v", err)
	}

	if *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	telegram, err := internal.NewTelegram(logger)
	if err != nil {
		logger.Fatalf("failed to create telegram: %v", err)
	}

	config, err := internal.NewConfig(*configPath)
	if err != nil {
		logger.Fatalf("failed to create config: %v", err)
	}

	bot, err := internal.NewBot(config, logger, rand.IntN, rand.Float64, telegram)
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bot.Run(ctx); err != nil {
			logger.Printf("failed to run bot: %v", err)
		}
	}()

	logger.Printf("bot started")

	<-signalChan
	if err := bot.Stop(ctx); err != nil {
		logger.Printf("failed to stop bot: %v", err)
		return
	}

	awaitTimeout(wg, 10*time.Second)
}

func awaitTimeout(wg *sync.WaitGroup, timeout time.Duration) {
	timeoutChan := time.After(timeout)
	waitChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-timeoutChan:
		return
	case <-waitChan:
		return
	}
}
