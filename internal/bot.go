package internal

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Message interface {
	Text() string
	Reply(ctx context.Context, message string) error
	React(ctx context.Context, emoji string) error
	Parent() Message
	IsMessageToMe() bool
}

type Integration interface {
	Subscribe(ctx context.Context) (<-chan Message, error)
	Stop(ctx context.Context) error
}

type Bot struct {
	trie        *SuffixTrie
	logger      *log.Logger
	answerRate  float64
	randIntN    func(n int) int
	randFloat64 func() float64
	integration []Integration
}

type command struct {
	Command commandType
	Args    []string
}

type commandType int

const (
	commandTypeUnknown commandType = iota
	commandTypeReply
)

func NewBot(
	answers *Answers,
	logger *log.Logger,
	answerRate float64,
	randIntN func(n int) int,
	randFloat64 func() float64,
	integration ...Integration,
) (*Bot, error) {
	if len(integration) == 0 {
		return nil, fmt.Errorf("at least one integration is required")
	}

	trie := NewSuffixTrie()
	for _, answer := range answers.Answers {
		trie.Insert(answer)
	}

	return &Bot{
		trie:        trie,
		logger:      logger,
		answerRate:  answerRate,
		randIntN:    randIntN,
		randFloat64: randFloat64,
		integration: integration,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	messageSubscriptions := make([]<-chan Message, len(b.integration))
	for _, integration := range b.integration {
		messages, err := integration.Subscribe(ctx)
		if err != nil {
			return fmt.Errorf("[bot] failed to subscribe to messages: %w", err)
		}
		messageSubscriptions = append(messageSubscriptions, messages)
	}

	for message := range messagesFanIn(messageSubscriptions) {
		b.logger.Printf("[bot] received message: %s", message.Text())

		if err := b.handleMessage(ctx, message); err != nil {
			b.logger.Printf("[bot] failed to handle message: %v", err)
		}
	}

	b.logger.Printf("[bot] all messages processed")

	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	b.logger.Printf("[bot] stopping")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}
	for _, integration := range b.integration {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := integration.Stop(ctx); err != nil {
				b.logger.Printf("[bot] failed to stop integration: %v", err)
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
		b.logger.Printf("[bot] stopped")
		return nil
	}
}

func (b *Bot) handleMessage(ctx context.Context, message Message) error {
	command := b.parseCommand(message)
	if command == nil {
		return nil
	}

	switch command.Command {
	case commandTypeReply:
		if err := b.handleReply(ctx, message); err != nil {
			return fmt.Errorf("failed to handle reply message: %w", err)
		}
	default:
		if err := b.handleDefault(ctx, message); err != nil {
			return fmt.Errorf("failed to handle default message: %w", err)
		}
	}

	return nil
}

func (b *Bot) handleReply(ctx context.Context, message Message) error {
	parentMessage := message.Parent()
	if parentMessage == nil {
		return nil
	}

	messageText := parentMessage.Text()
	answer, ok := b.answer(messageText)
	if !ok {
		answer = "meu pau ficou sem rimas"
	}

	if err := message.React(ctx, "👍"); err != nil {
		return fmt.Errorf("failed to react to message: %w", err)
	}

	if err := parentMessage.Reply(ctx, answer); err != nil {
		return fmt.Errorf("failed to reply to message: %w", err)
	}

	return nil
}

func (b *Bot) handleDefault(ctx context.Context, message Message) error {
	if b.answerRate < b.randFloat64() {
		return nil
	}

	messageText := message.Text()
	answer, ok := b.answer(messageText)
	if !ok {
		return nil
	}

	if err := message.Reply(ctx, answer); err != nil {
		return fmt.Errorf("failed to reply to message: %w", err)
	}

	return nil
}

func (b *Bot) parseCommand(message Message) *command {
	if message.IsMessageToMe() {
		return &command{Command: commandTypeReply, Args: []string{}}
	}

	return nil
}

func (b *Bot) answer(message string) (string, bool) {
	answers := b.trie.Search(message)
	if len(answers) == 0 {
		return "", false
	}

	return answers[b.randIntN(len(answers))], true
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

func cleanMessage(message string) string {
	// remove all special characters
	replacePattern := regexp.MustCompile(`[!@#$%^&*()+={}|\\:;<>,.?/]`)
	message = replacePattern.ReplaceAllString(message, "")

	// remove all extra spaces
	replacePattern = regexp.MustCompile(`\s+`)
	message = replacePattern.ReplaceAllString(message, " ")

	message = strings.ToLower(message)
	message = strings.TrimSpace(message)
	return message
}
