package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Session struct {
	ID        int    `gorm:"column:id;primary_key;auto_increment"`
	ProfileID uint32 `gorm:"column:profile_id"`
	SessionID string `gorm:"column:session_id"`
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value("session").(*Session)
	if !ok {
		return nil, errors.New("session not found")
	}
	return sess, nil
}

func (sm *ServerManager) CheckSession(r *http.Request) (*Session, error) {
	sessionCookie, err := r.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, fmt.Errorf("no cookie found")
	} else if err != nil {
		return nil, err
	}

	sess, err := sm.GetSessionBySessID(sessionCookie.Value)

	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}
	return sess, nil
}

func (sm *ServerManager) CreateSession(w *http.ResponseWriter, profileID string) string {

	var session Session

	sm.db.Where("profile_id = ?", profileID).First(&session)

	cookie := &http.Cookie{
		Name:    "token",
		Value:   session.SessionID,
		Expires: time.Now().Add(90 * 25 * time.Hour),
		Path:    "/",
	}

	http.SetCookie(*w, cookie)

	return session.SessionID
}

func (sm *ServerManager) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err == nil {
		sm.DeleteSessionDB(sess.ProfileID)
	}
	cookie := &http.Cookie{
		Name:    "session_id",
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
	return nil
}
