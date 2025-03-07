package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Data      map[string]interface{} `json:"data"`
}

// SessionRepository provides session management functionality
type SessionRepository interface {
	// Create a new session
	Create(ctx context.Context, userID string, duration time.Duration, data map[string]interface{}) (*Session, error)

	// Get a session by ID
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Update a session's data
	Update(ctx context.Context, sessionID string, data map[string]interface{}) error

	// Extend a session's expiration
	Extend(ctx context.Context, sessionID string, duration time.Duration) error

	// Delete a session
	Delete(ctx context.Context, sessionID string) error

	// Get all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*Session, error)

	// Delete all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error
}

// sessionRepository implements the SessionRepository interface
type sessionRepository struct {
	redis Repository
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(redis Repository) SessionRepository {
	return &sessionRepository{
		redis: redis,
	}
}

// Create creates a new session
func (s *sessionRepository) Create(ctx context.Context, userID string, duration time.Duration, data map[string]interface{}) (*Session, error) {
	// Generate a unique session ID
	sessionID := uuid.New().String()

	// Create the session object
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
		Data:      data,
	}

	// Serialize the session
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session: %w", err)
	}

	// Store the session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	if err := s.redis.Set(ctx, sessionKey, sessionData, duration); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Add the session to the user's session list
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
	if err := s.redis.SAdd(ctx, userSessionsKey, sessionID); err != nil {
		// Try to clean up the session if we can't add it to the user's list
		_ = s.redis.Delete(ctx, sessionKey)
		return nil, fmt.Errorf("failed to associate session with user: %w", err)
	}

	return session, nil
}

// Get retrieves a session by ID
func (s *sessionRepository) Get(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// Get the session data from Redis
	data, err := s.redis.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Deserialize the session
	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to deserialize session: %w", err)
	}

	// Check if the session has expired
	if time.Now().After(session.ExpiresAt) {
		_ = s.Delete(ctx, sessionID)
		return nil, fmt.Errorf("session has expired")
	}

	return &session, nil
}

// Update updates a session's data
func (s *sessionRepository) Update(ctx context.Context, sessionID string, data map[string]interface{}) error {
	// Get the current session
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Update the data
	session.Data = data

	// Calculate remaining TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session has expired")
	}

	// Serialize and store the updated session
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	return s.redis.Set(ctx, sessionKey, sessionData, ttl)
}

// Extend extends a session's expiration
func (s *sessionRepository) Extend(ctx context.Context, sessionID string, duration time.Duration) error {
	// Get the current session
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Update the expiration time
	session.ExpiresAt = time.Now().Add(duration)

	// Serialize and store the updated session
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	return s.redis.Set(ctx, sessionKey, sessionData, duration)
}

// Delete deletes a session
func (s *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	// Get the session to find the user ID
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		// If the session doesn't exist, we're done
		return nil
	}

	// Remove the session from the user's session list
	userSessionsKey := fmt.Sprintf("user:%s:sessions", session.UserID)
	if err := s.redis.SAdd(ctx, userSessionsKey, sessionID); err != nil {
		// Log the error but continue with deletion
		fmt.Printf("Failed to remove session from user's list: %v\n", err)
	}

	// Delete the session
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	return s.redis.Delete(ctx, sessionKey)
}

// GetByUserID gets all sessions for a user
func (s *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*Session, error) {
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	// Get all session IDs for the user
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*Session{}, nil
	}

	// Get each session
	sessions := make([]*Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := s.Get(ctx, sessionID)
		if err != nil {
			// Skip expired or invalid sessions
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteByUserID deletes all sessions for a user
func (s *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	// Get all session IDs for the user
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Delete each session
	for _, sessionID := range sessionIDs {
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		if err := s.redis.Delete(ctx, sessionKey); err != nil {
			// Log the error but continue with deletion
			fmt.Printf("Failed to delete session %s: %v\n", sessionID, err)
		}
	}

	// Delete the user's session list
	return s.redis.Delete(ctx, userSessionsKey)
}
