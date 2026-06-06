package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type BotAPI struct {
	token  string
	client *http.Client
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      Chat   `json:"chat"`
	From      *User  `json:"from"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	Data    string   `json:"data"`
	Message *Message `json:"message"`
	From    User     `json:"from"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func main() {
	filePath := flag.String("file", "", "path to file to send")
	chatID := flag.Int64("chat-id", 0, "target telegram chat id")
	username := flag.String("username", "", "target telegram username (without @)")
	formatType := flag.String("format", "video", "format: video, audio, document")
	caption := flag.String("caption", "", "message caption")
	doSetupSession := flag.Bool("setup-session", false, "run interactive session setup: authenticate via phone number and output TG_SESSION value")
	sessionPath := flag.String("session", "session.json", "session file path")
	flag.Parse()

	cfg := LoadConfig()

	if *doSetupSession {
		if err := runSessionSetup(cfg, *sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *filePath != "" || *chatID != 0 {
		if err := runUploader(cfg, *filePath, *chatID, *username, *formatType, *caption, *sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runBot(); err != nil {
		log.Fatalf("bot stopped: %v", err)
	}
}

func runBot() error {
	cfg := LoadConfig()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	bot := &BotAPI{
		token:  cfg.BotToken,
		client: &http.Client{Timeout: 35 * time.Second},
	}

	log.Println("bot started successfully")
	return bot.Poll(context.Background())
}

func (b *BotAPI) Poll(ctx context.Context) error {
	var offset int64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.GetUpdates(ctx, offset)
		if err != nil {
			log.Printf("get updates failed: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			if err := b.HandleUpdate(ctx, update); err != nil {
				log.Printf("handle update failed: %v", err)
			}
		}
	}
}

func (b *BotAPI) GetUpdates(ctx context.Context, offset int64) ([]Update, error) {
	var result struct {
		OK          bool     `json:"ok"`
		Description string   `json:"description"`
		Result      []Update `json:"result"`
	}

	err := b.Call(ctx, "getUpdates", map[string]any{
		"offset":  offset,
		"timeout": 25,
	}, &result)
	if err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", result.Description)
	}
	return result.Result, nil
}

func (b *BotAPI) HandleUpdate(ctx context.Context, update Update) error {
	if update.CallbackQuery != nil {
		return onQualitySelect(ctx, b, update.CallbackQuery)
	}
	if update.Message == nil {
		return nil
	}

	text := update.Message.Text
	switch text {
	case "/start":
		return onStart(ctx, b, update.Message)
	case "/help":
		return onHelp(ctx, b, update.Message)
	case "/cancel":
		return onCancel(ctx, b, update.Message)
	default:
		return onText(ctx, b, update.Message)
	}
}

func (b *BotAPI) SendMessage(ctx context.Context, chatID int64, text string, markup *InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	return b.Call(ctx, "sendMessage", payload, nil)
}

func (b *BotAPI) AnswerCallback(ctx context.Context, callbackID string) error {
	return b.Call(ctx, "answerCallbackQuery", map[string]any{
		"callback_query_id": callbackID,
	}, nil)
}

func (b *BotAPI) Call(ctx context.Context, method string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.URL(method), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(respBody))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	return nil
}

func (b *BotAPI) URL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}