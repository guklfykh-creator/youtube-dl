package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

var (
	appID   int
	appHash string
)

func runUploader(filePath string, chatID int64, username, formatType, caption string, doAuth, doQR bool, sessionPath string) error {
	appIDStr := os.Getenv("TG_APP_ID")
	appHash = os.Getenv("TG_APP_HASH")
	sessionB64 := os.Getenv("TG_SESSION")

	if appIDStr == "" || appHash == "" {
		return fmt.Errorf("TG_APP_ID and TG_APP_HASH env vars are required")
	}

	appIDInt, err := strconv.Atoi(appIDStr)
	if err != nil {
		return fmt.Errorf("TG_APP_ID must be a valid integer")
	}
	appID = appIDInt

	if sessionB64 != "" {
		data, err := base64.StdEncoding.DecodeString(sessionB64)
		if err != nil {
			return fmt.Errorf("session base64 decode failed: %w", err)
		}
		if err := os.WriteFile(sessionPath, data, 0600); err != nil {
			return fmt.Errorf("session file write failed: %w", err)
		}
	}

	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionPath},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	return client.Run(ctx, func(ctx context.Context) error {
		if doAuth {
			return runPhoneAuth(ctx, client)
		}
		if doQR {
			return fmt.Errorf("--qr is not supported; use --auth")
		}

		me, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("not authenticated (run with --auth first): %w", err)
		}

		fmt.Printf("authenticated as user %d (%s)\n", me.ID, me.FirstName)

		if filePath == "" || chatID == 0 {
			return fmt.Errorf("--file and --chat-id are required for sending")
		}

		peer, err := resolvePeer(ctx, client, chatID, username)
		if err != nil {
			return fmt.Errorf("peer resolution failed: %w", err)
		}

		fmt.Printf("resolved peer for chat_id %d\n", chatID)

		return sendFile(ctx, client, peer, filePath, formatType, caption)
	})
}

type terminalAuth struct{}

func (t terminalAuth) Phone(ctx context.Context) (string, error) {
	phone := os.Getenv("TG_PHONE")
	if phone != "" {
		return phone, nil
	}
	fmt.Print("Enter phone number (international format, e.g. +989123456789): ")
	var input string
	_, _ = fmt.Scan(&input)
	return input, nil
}

func (t terminalAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	var code string
	_, _ = fmt.Scan(&code)
	return code, nil
}

func (t terminalAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	var password string
	_, _ = fmt.Scan(&password)
	return password, nil
}

func (t terminalAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	fmt.Println("Telegram terms of service must be accepted to continue.")
	return nil
}

func (t terminalAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	fmt.Print("Enter first name: ")
	var first string
	_, _ = fmt.Scan(&first)
	fmt.Print("Enter last name: ")
	var last string
	_, _ = fmt.Scan(&last)
	return auth.UserInfo{FirstName: first, LastName: last}, nil
}

func runPhoneAuth(ctx context.Context, client *telegram.Client) error {
	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})
	if err := flow.Run(ctx, client.Auth()); err != nil {
		return fmt.Errorf("phone auth failed: %w", err)
	}
	fmt.Println("authentication successful, session saved")
	return nil
}

func resolvePeer(ctx context.Context, client *telegram.Client, chatID int64, username string) (tg.InputPeerClass, error) {
	if username != "" {
		result, err := client.API().ContactsResolveUsername(ctx, strings.TrimPrefix(username, "@"))
		if err != nil {
			return nil, fmt.Errorf("resolve username '%s': %w", username, err)
		}

		for _, u := range result.Users {
			user, ok := u.(*tg.User)
			if !ok {
				continue
			}
			if user.ID == chatID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
		}

		for _, u := range result.Users {
			user, ok := u.(*tg.User)
			if !ok {
				continue
			}
			if strings.EqualFold(user.Username, strings.TrimPrefix(username, "@")) {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
		}
	}

	if chatID > 0 {
		result, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			Limit:         100,
			ExcludePinned: true,
		})
		if err != nil {
			return nil, fmt.Errorf("get dialogs: %w", err)
		}

		var users []tg.UserClass
		switch r := result.(type) {
		case *tg.MessagesDialogs:
			users = r.Users
		case *tg.MessagesDialogsSlice:
			users = r.Users
		}

		for _, u := range users {
			user, ok := u.(*tg.User)
			if !ok {
				continue
			}
			if user.ID == chatID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
		}

		return nil, fmt.Errorf("user %d not found in dialogs (provide --username)", chatID)
	}

	if chatID < -1000000000 {
		channelID := -chatID - 1000000000
		chResult, err := client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{ChannelID: channelID},
		})
		if err != nil {
			return nil, fmt.Errorf("get channel %d: %w", channelID, err)
		}

		for _, c := range chResult.GetChats() {
			ch, ok := c.(*tg.Channel)
			if !ok {
				continue
			}
			if ch.ID == channelID {
				return &tg.InputPeerChannel{ChannelID: ch.ID, AccessHash: ch.AccessHash}, nil
			}
		}

		return nil, fmt.Errorf("channel %d not found", channelID)
	}

	return &tg.InputPeerChat{ChatID: -chatID}, nil
}

func sendFile(ctx context.Context, client *telegram.Client, peer tg.InputPeerClass, filePath, formatType, caption string) error {
	filename := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	up := uploader.NewUploader(client.API())
	uploaded, err := up.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	fmt.Printf("uploaded %s (%d bytes)\n", filename, fileInfo.Size())

	var mimeType string
	var attrs []tg.DocumentAttributeClass

	switch formatType {
	case "video":
		mimeType = mimeTypeFromExt(ext, "video/mp4")
		attrs = []tg.DocumentAttributeClass{
			&tg.DocumentAttributeVideo{Duration: 0},
			&tg.DocumentAttributeFilename{FileName: filename},
		}
	case "audio":
		mimeType = mimeTypeFromExt(ext, "audio/mpeg")
		attrs = []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAudio{Duration: 0, Title: caption},
			&tg.DocumentAttributeFilename{FileName: filename},
		}
	default:
		mimeType = "application/octet-stream"
		attrs = []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{FileName: filename},
		}
	}

	media := &tg.InputMediaUploadedDocument{
		File:       uploaded,
		MimeType:   mimeType,
		Attributes: attrs,
	}

	randomID, err := client.RandInt64()
	if err != nil {
		return fmt.Errorf("generate random id: %w", err)
	}

	req := &tg.MessagesSendMediaRequest{
		Peer:     peer,
		Media:    media,
		Message:  caption,
		RandomID: randomID,
	}

	_, err = client.API().MessagesSendMedia(ctx, req)
	if err != nil {
		return fmt.Errorf("send media: %w", err)
	}

	fmt.Printf("sent %s to chat %d\n", filename, peerID(peer))
	return nil
}

func mimeTypeFromExt(ext, fallback string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	case ".wav":
		return "audio/wav"
	default:
		return fallback
	}
}

func peerID(peer tg.InputPeerClass) int64 {
	switch p := peer.(type) {
	case *tg.InputPeerUser:
		return int64(p.UserID)
	case *tg.InputPeerChat:
		return -p.ChatID
	case *tg.InputPeerChannel:
		return -(p.ChannelID + 1000000000)
	default:
		return 0
	}
}
