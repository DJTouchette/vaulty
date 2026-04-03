package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ApprovalStatus represents the state of a pending approval.
type ApprovalStatus string

const (
	StatusPending  ApprovalStatus = "pending"
	StatusApproved ApprovalStatus = "approved"
	StatusDenied   ApprovalStatus = "denied"
)

// PendingApproval holds a pending secret injection request awaiting user decision.
type PendingApproval struct {
	ID         string          `json:"id"`
	SecretName string          `json:"secret_name"`
	Target     string          `json:"target"`
	Action     string          `json:"action"` // "proxy" or "exec"
	Status     ApprovalStatus  `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  time.Time       `json:"expires_at"`
	Args       json.RawMessage `json:"-"` // original request args for re-execution
}

// ApprovalStore is an in-memory store of pending approvals.
type ApprovalStore struct {
	mu      sync.Mutex
	entries map[string]*PendingApproval
	counter int
	timeout time.Duration
}

// NewApprovalStore creates an approval store with the given expiration timeout.
func NewApprovalStore(timeout time.Duration) *ApprovalStore {
	return &ApprovalStore{
		entries: make(map[string]*PendingApproval),
		timeout: timeout,
	}
}

// Create adds a new pending approval and returns it.
func (s *ApprovalStore) Create(secretName, target, action string, args json.RawMessage) *PendingApproval {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	id := fmt.Sprintf("approval-%d", s.counter)
	now := time.Now()

	pa := &PendingApproval{
		ID:         id,
		SecretName: secretName,
		Target:     target,
		Action:     action,
		Status:     StatusPending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.timeout),
		Args:       args,
	}
	s.entries[id] = pa
	return pa
}

// Get returns a pending approval by ID.
func (s *ApprovalStore) Get(id string) (*PendingApproval, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pa, ok := s.entries[id]
	if !ok {
		return nil, false
	}
	return pa, true
}

// Approve marks a pending approval as approved.
func (s *ApprovalStore) Approve(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pa, ok := s.entries[id]
	if !ok {
		return fmt.Errorf("approval %q not found", id)
	}
	if pa.Status != StatusPending {
		return fmt.Errorf("approval %q already %s", id, pa.Status)
	}
	if time.Now().After(pa.ExpiresAt) {
		return fmt.Errorf("approval %q has expired", id)
	}
	pa.Status = StatusApproved
	return nil
}

// Deny marks a pending approval as denied.
func (s *ApprovalStore) Deny(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pa, ok := s.entries[id]
	if !ok {
		return fmt.Errorf("approval %q not found", id)
	}
	if pa.Status != StatusPending {
		return fmt.Errorf("approval %q already %s", id, pa.Status)
	}
	pa.Status = StatusDenied
	return nil
}

// Cleanup removes expired entries from the store.
func (s *ApprovalStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, pa := range s.entries {
		if now.After(pa.ExpiresAt) {
			delete(s.entries, id)
		}
	}
}

// ListPending returns all pending (non-expired) approvals.
func (s *ApprovalStore) ListPending() []*PendingApproval {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var result []*PendingApproval
	for _, pa := range s.entries {
		if pa.Status == StatusPending && now.Before(pa.ExpiresAt) {
			result = append(result, pa)
		}
	}
	return result
}
