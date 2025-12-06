package main

import (
	"fmt"
	"log"
	"strings"

	tb "gopkg.in/telebot.v3"
)

var qualityButtons []*tb.InlineButton

var (
	btnVideoBest = &tb.InlineButton{Text: "ویدیو - بهترین کیفیت", Data: "v_best"}
	btnVideo1080 = &tb.InlineButton{Text: "ویدیو - 1080p", Data: "v_1080"}
	btnVideo720  = &tb.InlineButton{Text: "ویدیو - 720p", Data: "v_720"}
	btnVideo480  = &tb.InlineButton{Text: "ویدیو - 480p", Data: "v_480"}
	btnVideo360  = &tb.InlineButton{Text: "ویدیو - 360p", Data: "v_360"}
	btnAudioMP3  = &tb.InlineButton{Text: "صدا - MP3", Data: "a_mp3"}
	btnAudioM4A  = &tb.InlineButton{Text: "صدا - M4A", Data: "a_m4a"}
)

func init() {
	qualityButtons = []*tb.InlineButton{
		btnVideoBest, btnVideo1080, btnVideo720, btnVideo480, btnVideo360,
		btnAudioMP3, btnAudioM4A,
	}
}

func onStart(c tb.Context) error {
	return c.Send(
		"👋 سلام! ربات دانلود یوتیوب\n\n"+
			"لینک ویدیو یوتیوب را بفرستید تا برایتان دانلود کنم.\n\n"+
			"📌 مراحل:\n"+
			"1. لینک ویدیو را بفرستید\n"+
			"2. کیفیت یا فرمت را انتخاب کنید\n"+
			"3. فایل برایتان ارسال میشود\n\n"+
			"💡 /help برای راهنما")
}

func onHelp(c tb.Context) error {
	return c.Send(
		"📖 راهنمای ربات:\n\n"+
			"1️⃣ لینک ویدیو یوتیوب را بفرستید\n"+
			"2️⃣ کیفیت مورد نظر را از منو انتخاب کنید:\n"+
			"   • ویدیو: بهترین، 1080p، 720p، 480p، 360p\n"+
			"   • صدا: MP3 یا M4A\n"+
			"3️⃣ فایل توسط سرور دانلود و برایتان ارسال میشود\n\n"+
			"⚠️ فایلهای کمتر از 50MB مستقیم ارسال میشوند\n"+
			"⚠️ فایلهای بیشتر از 50MB لینک دانلود مستقیم دریافت خواهید کرد\n\n"+
			"/cancel - انصراف")
}

func onCancel(c tb.Context) error {
	chatID := c.Chat().ID
	DelSession(chatID)
	return c.Send("❌ نشست پاک شد. لینک یوتیوب را دوباره بفرستید.")
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

func onText(c tb.Context) error {
	text := strings.TrimSpace(c.Text())

	if !isYouTubeURL(text) {
		if c.Chat().Type == "private" {
			return c.Send("❌ لطفا لینک یوتیوب معتبر بفرستید.")
		}
		return nil
	}

	chatID := c.Chat().ID
	SetSession(chatID, text)

	markup := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{btnVideoBest, btnVideo1080},
			{btnVideo720, btnVideo480},
			{btnVideo360},
			{btnAudioMP3, btnAudioM4A},
		},
	}

	return c.Send("🎵 فرمت و کیفیت مورد نظر را انتخاب کنید:", markup)
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

func onQualitySelect(c tb.Context) error {
	c.Respond()

	chatID := c.Chat().ID
	session := GetSession(chatID)
	if session == nil {
		return c.Send("❌ نشست منقضی شده. لینک یوتیوب را دوباره بفرستید.")
	}

	formatType, quality := parseCallbackData(c.Data())
	if formatType == "" {
		return c.Send("❌ انتخاب نامعتبر. دوباره تلاش کنید.")
	}

	chatIDStr := fmt.Sprintf("%d", chatID)
	username := c.Sender().Username

	cfg := LoadConfig()
	if err := TriggerWorkflow(cfg, session.URL, formatType, quality, chatIDStr, username); err != nil {
		log.Printf("workflow trigger failed: %v (url=%s format=%s quality=%s chatID=%s)",
			err, session.URL, formatType, quality, chatIDStr)
		return c.Send("❌ خطا در شروع دانلود. لطفا دوباره تلاش کنید.")
	}

	DelSession(chatID)

	log.Printf("workflow triggered: url=%s format=%s quality=%s chatID=%s username=%s",
		session.URL, formatType, quality, chatIDStr, username)

	return c.Send("⏳ در حال دانلود... لطفا صبر کنید. فایل به زودی برایتان ارسال خواهد شد.")
}