package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

type Telegram struct {
	bot              *tgbotapi.BotAPI
	whitelistedChats []int64
	logger           *log.Logger
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

	t := &Telegram{bot: bot, logger: logger}

	whitelistedChats := strings.Split(os.Getenv("TG_CHAT_WHITELIST"), ",")
	if len(whitelistedChats) > 0 {
		for _, chat := range whitelistedChats {
			if chat == "" {
				continue
			}

			chatID, err := strconv.ParseInt(chat, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse TG_CHAT_WHITELIST environment variable: %w", err)
			}
			t.whitelistedChats = append(t.whitelistedChats, chatID)
		}
	}

	return t, nil
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

			t.logger.Printf("[telegram] message received. chat_id: %d, message_id: %d, text: %s",
				update.Message.Chat.ID,
				update.Message.MessageID,
				update.Message.Text,
			)

			if len(t.whitelistedChats) > 0 && !slices.Contains(t.whitelistedChats, update.Message.Chat.ID) {
				t.logger.Printf("[telegram] chat is not allowed: %d", update.Message.Chat.ID)
				continue
			}

			msg, err := newTelegramMessage(&update, t)
			if err != nil {
				t.logger.Printf("[telegram] failed to create telegram message: %v", err)
				continue
			}

			ch <- msg
		}

		t.logger.Printf("[telegram] updates channel closed")
	}()

	return ch, nil
}

func (t *Telegram) Stop(ctx context.Context) error {
	t.logger.Printf("[telegram] stopping")
	t.bot.StopReceivingUpdates()
	t.logger.Printf("[telegram] stopped")
	return nil
}

func (t *Telegram) clearMessages() int {
	t.logger.Printf("[telegram] clearing messages")
	offset := 0
	for {
		t.logger.Printf("[telegram] clearing messages with offset %d", offset)

		updates, err := t.bot.GetUpdates(tgbotapi.UpdateConfig{
			Offset:  offset,
			Limit:   0,
			Timeout: 0,
		})
		if err != nil {
			t.logger.Printf("[telegram] failed to get updates: %v", err)
			return offset
		}

		for _, update := range updates {
			if update.Message != nil {
				t.logger.Printf("[telegram] message: %s", update.Message.Text)
			} else {
				t.logger.Printf("[telegram] discarting update")
			}
		}

		if len(updates) == 0 {
			break
		}

		offset = updates[len(updates)-1].UpdateID + 1
	}

	t.logger.Printf("[telegram] cleared %d messages", offset)

	return offset
}

type telegramMessage struct {
	update      *tgbotapi.Update
	integration *Telegram
}

func newTelegramMessage(update *tgbotapi.Update, integration *Telegram) (Message, error) {
	if update.Message == nil {
		return nil, fmt.Errorf("message is nil")
	}

	return &telegramMessage{update: update, integration: integration}, nil
}

func (m *telegramMessage) Text() string {
	return cleanMessage(m.update.Message.Text)
}

func (m *telegramMessage) Reply(ctx context.Context, message string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: m.update.Message.Chat.ID,
			},
			ReplyParameters: tgbotapi.ReplyParameters{
				MessageID: m.update.Message.MessageID,
			},
		},
		Text: message,
	}
	if _, err := m.integration.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (m *telegramMessage) Parent() Message {
	if m.update.Message.ReplyToMessage == nil {
		return nil
	}

	update := tgbotapi.Update{
		Message: m.update.Message.ReplyToMessage,
	}

	return &telegramMessage{update: &update, integration: m.integration}
}

func (m *telegramMessage) IsMessageToMe() bool {
	if m.update.Message == nil {
		return false
	}

	for _, entity := range m.update.Message.Entities {
		if entity.Type == "mention" {
			mention := []rune(m.update.Message.Text)[entity.Offset : entity.Offset+entity.Length]
			if strings.Contains(string(mention), m.integration.bot.Self.UserName) {
				return true
			}
		}
	}

	return false
}

// TODO: Implement
func (m *telegramMessage) React(ctx context.Context, emoji string) error {
	return nil
}
