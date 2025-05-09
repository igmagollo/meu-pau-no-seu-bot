package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Message interface {
	Text() string
	Reply(ctx context.Context, message string) error
}

type Integration interface {
	Subscribe(ctx context.Context) (<-chan Message, error)
	Stop(ctx context.Context) error
}

type Bot struct {
	config      *Config
	logger      *log.Logger
	randIntN    func(n int) int
	randFloat64 func() float64
	integration []Integration
}

func NewBot(
	config *Config,
	logger *log.Logger,
	randIntN func(n int) int,
	randFloat64 func() float64,
	integration ...Integration,
) (*Bot, error) {
	if len(integration) == 0 {
		return nil, fmt.Errorf("at least one integration is required")
	}

	return &Bot{
		config:      config,
		logger:      logger,
		randIntN:    randIntN,
		randFloat64: randFloat64,
		integration: integration,
	}, nil
}

func (b *Bot) answer(message string) (string, bool) {
	if b.config.AnswerRate < b.randFloat64() {
		return "", false
	}

	for suffix, answers := range b.config.Suffixes {
		if strings.HasSuffix(message, suffix) {
			return answers[b.randIntN(len(answers))], true
		}
	}

	return "", false
}

func (b *Bot) Run(ctx context.Context) error {
	messageSubscriptions := make([]<-chan Message, len(b.integration))
	for _, integration := range b.integration {
		messages, err := integration.Subscribe(ctx)
		if err != nil {
			return fmt.Errorf("failed to subscribe to messages: %w", err)
		}
		messageSubscriptions = append(messageSubscriptions, messages)
	}

	for message := range messagesFanIn(messageSubscriptions) {
		b.logger.Printf("received message: %s", message.Text())
		answer, ok := b.answer(message.Text())
		if !ok {
			continue
		}

		if err := message.Reply(ctx, answer); err != nil {
			return fmt.Errorf("failed to reply to message: %w", err)
		}
	}

	b.logger.Printf("all messages processed")

	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	b.logger.Printf("stopping bot")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}
	for _, integration := range b.integration {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := integration.Stop(ctx); err != nil {
				b.logger.Printf("failed to stop integration: %v", err)
			}
		}()
	}

	closeChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(closeChan)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout while stopping bot")
	case <-closeChan:
		b.logger.Printf("bot stopped")
		return nil
	}
}

func messagesFanIn(channels []<-chan Message) <-chan Message {
	out := make(chan Message)

	wg := &sync.WaitGroup{}
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan Message) {
			defer wg.Done()
			for message := range c {
				out <- message
			}
		}(ch)
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}
