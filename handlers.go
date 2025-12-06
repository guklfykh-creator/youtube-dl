package main

import (
	"context"
	"fmt"
	"log"
	"strings"
)

var qualityButtons = []InlineKeyboardButton{
	{Text: "ویدیو - بهترین کیفیت", CallbackData: "v_best"},
	{Text: "ویدیو - 1080p", CallbackData: "v_1080"},
	{Text: "ویدیو - 720p", CallbackData: "v_720"},
	{Text: "ویدیو - 480p", CallbackData: "v_480"},
	{Text: "ویدیو - 360p", CallbackData: "v_360"},
	{Text: "صدا - MP3", CallbackData: "a_mp3"},
	{Text: "صدا - M4A", CallbackData: "a_m4a"},
}

func onStart(ctx context.Context, bot *BotAPI, msg *Message) error {
	return bot.SendMessage(ctx, msg.Chat.ID,
		"👋 سلام! ربات دانلود یوتیوب\n\n"+
			"لینک ویدیو یوتیوب را بفرستید تا برایتان دانلود کنم.\n\n"+
			"📌 مراحل:\n"+
			"1. لینک ویدیو را بفرستید\n"+
			"2. کیفیت یا فرمت را انتخاب کنید\n"+
			"3. فایل برایتان ارسال میشود\n\n"+
			"💡 /help برای راهنما", nil)
}

func onHelp(ctx context.Context, bot *BotAPI, msg *Message) error {
	return bot.SendMessage(ctx, msg.Chat.ID,
		"📖 راهنمای ربات:\n\n"+
			"1️⃣ لینک ویدیو یوتیوب را بفرستید\n"+
			"2️⃣ کیفیت مورد نظر را از منو انتخاب کنید:\n"+
			"   • ویدیو: بهترین، 1080p، 720p، 480p، 360p\n"+
			"   • صدا: MP3 یا M4A\n"+
			"3️⃣ فایل توسط سرور دانلود و برایتان ارسال میشود\n\n"+
			"⚠️ فایلهای کمتر از 50MB مستقیم ارسال میشوند\n"+
			"⚠️ فایلهای بیشتر از 50MB با MTProto ارسال میشوند\n\n"+
			"/cancel - انصراف", nil)
}

func onCancel(ctx context.Context, bot *BotAPI, msg *Message) error {
	DelSession(msg.Chat.ID)
	return bot.SendMessage(ctx, msg.Chat.ID, "❌ نشست پاک شد. لینک یوتیوب را دوباره بفرستید.", nil)
}

func isYouTubeURL(text string) bool {
	lower := strings.ToLower(text)
	patterns := []string{
		"youtube.com/watch",
		"youtu.be/",
		"youtube.com/shorts/",
		"youtube.com/v/",
		"youtube.com/embed/",
		"m.youtube.com/watch",
		"youtube.com/live/",
		"music.youtube.com/",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func onText(ctx context.Context, bot *BotAPI, msg *Message) error {
	text := strings.TrimSpace(msg.Text)

	if !isYouTubeURL(text) {
		if msg.Chat.Type == "private" {
			return bot.SendMessage(ctx, msg.Chat.ID, "❌ لطفا لینک یوتیوب معتبر بفرستید.", nil)
		}
		return nil
	}

	SetSession(msg.Chat.ID, text)

	markup := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{qualityButtons[0], qualityButtons[1]},
			{qualityButtons[2], qualityButtons[3]},
			{qualityButtons[4]},
			{qualityButtons[5], qualityButtons[6]},
		},
	}

	return bot.SendMessage(ctx, msg.Chat.ID, "🎵 فرمت و کیفیت مورد نظر را انتخاب کنید:", markup)
}

func parseCallbackData(data string) (formatType, quality string) {
	if strings.HasPrefix(data, "v_") {
		return "video", strings.TrimPrefix(data, "v_")
	}
	if strings.HasPrefix(data, "a_") {
		return "audio", strings.TrimPrefix(data, "a_")
	}
	return "", ""
}

func onQualitySelect(ctx context.Context, bot *BotAPI, callback *CallbackQuery) error {
	_ = bot.AnswerCallback(ctx, callback.ID)
	if callback.Message == nil {
		return nil
	}

	chatID := callback.Message.Chat.ID
	session := GetSession(chatID)
	if session == nil {
		return bot.SendMessage(ctx, chatID, "❌ نشست منقضی شده. لینک یوتیوب را دوباره بفرستید.", nil)
	}

	formatType, quality := parseCallbackData(callback.Data)
	if formatType == "" {
		return bot.SendMessage(ctx, chatID, "❌ انتخاب نامعتبر. دوباره تلاش کنید.", nil)
	}

	chatIDStr := fmt.Sprintf("%d", chatID)
	username := callback.From.Username

	cfg := LoadConfig()
	if err := TriggerWorkflow(cfg, session.URL, formatType, quality, chatIDStr, username); err != nil {
		log.Printf("workflow trigger failed: %v (format=%s quality=%s chatID=%s)",
			err, formatType, quality, chatIDStr)
		return bot.SendMessage(ctx, chatID, "❌ خطا در شروع دانلود. لطفا دوباره تلاش کنید.", nil)
	}

	DelSession(chatID)

	log.Printf("workflow triggered: format=%s quality=%s chatID=%s username=%s",
		formatType, quality, chatIDStr, username)

	return bot.SendMessage(ctx, chatID, "⏳ در حال دانلود... لطفا صبر کنید. فایل به زودی برایتان ارسال خواهد شد.", nil)
}
