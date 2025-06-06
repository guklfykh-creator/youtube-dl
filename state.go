package main

import (
	"sync"
	"time"
)

type Session struct {
	URL       string
	CreatedAt time.Time
}

var sessions = struct {
	sync.RWMutex
	m map[int64]*Session
}{m: make(map[int64]*Session)}

const sessionTimeout = 5 * time.Minute

func SetSession(chatID int64, url string) {
	sessions.Lock()
	sessions.m[chatID] = &Session{URL: url, CreatedAt: time.Now()}
	sessions.Unlock()
}

func GetSession(chatID int64) *Session {
	sessions.RLock()
	s, ok := sessions.m[chatID]
	sessions.RUnlock()
	if !ok {
		return nil
	}
	if time.Since(s.CreatedAt) > sessionTimeout {
		DelSession(chatID)
		return nil
	}
	return s
}

func DelSession(chatID int64) {
	sessions.Lock()
	delete(sessions.m, chatID)
	sessions.Unlock()
}