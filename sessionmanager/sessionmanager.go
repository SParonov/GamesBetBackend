package sessionmanager

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dchest/uniuri"
)

type Session struct {
	ID        string
	UserEmail string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions              map[string]*Session
	db                    *sql.DB
	mutex                 *sync.Mutex
	sessionExtensionDelta time.Duration
}

func New(db *sql.DB, sessionExtensionDelta time.Duration) *SessionManager {
	return &SessionManager{
		sessions:              make(map[string]*Session),
		db:                    db,
		mutex:                 &sync.Mutex{},
		sessionExtensionDelta: sessionExtensionDelta,
	}
}

func (s *SessionManager) GetSession(sessionID string) (string, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, ok := s.sessions[sessionID]

	if ok && session.ExpiresAt.Before(time.Now()) {
		s.removeSession(session)
		session = nil
		return "", fmt.Errorf("session expired")
	}

	if session != nil {
		s.updateSession(session)
		return session.UserEmail, nil
	}

	return "", fmt.Errorf("session not found")
}

func (s *SessionManager) GetSessionData(sessionID string) (Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, ok := s.sessions[sessionID]

	if ok {
		return *session, nil
	}

	return Session{}, fmt.Errorf("session not found")
}

func (s *SessionManager) CreateSession(ctx context.Context, userEmail string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if session, ok := s.sessions[userEmail]; ok {
		s.removeSession(session)
	}

	newSessionID := uniuri.New()

	newSession := &Session{
		ID:        newSessionID,
		UserEmail: userEmail,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.sessionExtensionDelta),
	}

	s.sessions[newSession.ID] = newSession

	return newSession.ID, nil
}

func (s *SessionManager) CleanExpiredSessions() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	cnt := 0

	log.Println("start clean expired sessions")
	for _, session := range s.sessions {
		if session.ExpiresAt.Before(now) {
			s.removeSession(session)
			cnt += 1
		}
	}
	log.Printf("end clean expired sessions. removed %d sessions\n", cnt)

}

func (s *SessionManager) updateSession(session *Session) {
	session.ExpiresAt = time.Now().Add(s.sessionExtensionDelta)
}

func (s *SessionManager) removeSession(session *Session) {
	delete(s.sessions, session.ID)
}
