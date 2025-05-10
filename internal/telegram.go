package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	bot    *tgbotapi.BotAPI
	logger *log.Logger
}

func NewTelegram(logger *log.Logger) (*Telegram, error) {
	apiKey := os.Getenv("TG_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TG_API_KEY environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	if !bot.Self.CanReadAllGroupMessages {
		return nil, fmt.Errorf("bot is not allowed to read all group messages")
	}

	return &Telegram{bot: bot, logger: logger}, nil
}

func (t *Telegram) Subscribe(ctx context.Context) (<-chan Message, error) {
	offset := t.clearMessages()

	// Set up update configuration
	updateConfig := tgbotapi.NewUpdate(offset)
	updateConfig.Timeout = 60

	ch := make(chan Message)
	updates := t.bot.GetUpdatesChan(updateConfig)

	go func() {
		defer close(ch)

		for update := range updates {
			if update.Message == nil {
				continue
			}

			msg, err := newTelegramMessage(&update, t.bot)
			if err != nil {
				t.logger.Printf("failed to create telegram message: %v", err)
				continue
			}

			ch <- msg
		}

		t.logger.Printf("telegram updates channel closed")
	}()

	return ch, nil
}

func (t *Telegram) Stop(ctx context.Context) error {
	t.logger.Printf("stopping telegram")
	t.bot.StopReceivingUpdates()
	t.logger.Printf("telegram stopped")
	return nil
}

func (t *Telegram) clearMessages() int {
	t.logger.Printf("clearing messages")
	offset := 0
	for {
		t.logger.Printf("clearing messages with offset %d", offset)

		updates, err := t.bot.GetUpdates(tgbotapi.UpdateConfig{
			Offset:  offset,
			Limit:   0,
			Timeout: 0,
		})
		if err != nil {
			t.logger.Printf("failed to get updates: %v", err)
			return offset
		}

		for _, update := range updates {
			t.logger.Printf("update: %+v", update)
			if update.Message != nil {
				t.logger.Printf("message: %s", update.Message.Text)
			} else {
				t.logger.Printf("discarting update")
			}
		}

		if len(updates) == 0 {
			break
		}

		offset = updates[len(updates)-1].UpdateID + 1
	}

	t.logger.Printf("cleared %d messages", offset)

	return offset
}

type telegramMessage struct {
	update *tgbotapi.Update
	bot    *tgbotapi.BotAPI
}

func newTelegramMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI) (Message, error) {
	if update.Message == nil {
		return nil, fmt.Errorf("message is nil")
	}

	return &telegramMessage{update: update, bot: bot}, nil
}

func (m *telegramMessage) Text() string {
	return m.update.Message.Text
}

func (m *telegramMessage) Reply(ctx context.Context, message string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           m.update.Message.Chat.ID,
			ReplyToMessageID: m.update.Message.MessageID,
		},
		Text:                  message,
		DisableWebPagePreview: false,
	}
	if _, err := m.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
