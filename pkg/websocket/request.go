package websocket

import (
	"sync"

	"github.com/google/uuid"
)

type RegistrationRequest struct {
	InitialToken string `json:"initial_token"`
	DeviceMetadata
}

type RegistrationResponse struct {
	PermanentToken string `json:"permanent_token"`
	Message        string `json:"message"`
}

func formatUUID(id string) uuid.UUID {
	u, _ := uuid.Parse(id)
	return u
}

type CredentialStore struct {
	initialTokens   map[string]string    // deviceID string → initial token
	permanentTokens map[uuid.UUID]string // deviceID UUID → permanent token
	mu              sync.RWMutex
}

// NewCredentialStore initializes the store with pre-authorized initial tokens.
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		initialTokens: map[string]string{
			"5d679008-4117-4c06-8bdf-509489f82c70": "initial-1",
		},
		permanentTokens: make(map[uuid.UUID]string),
	}
}

// ValidateInitialToken checks if the initial token matches the expected one for the device.
func (s *CredentialStore) ValidateInitialToken(deviceID uuid.UUID, initToken string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedToken, ok := s.initialTokens[deviceID.String()]
	return ok && storedToken == initToken
}

// IsRegistered returns true if the device has already been registered.
func (s *CredentialStore) IsRegistered(deviceID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.permanentTokens[deviceID]
	return ok
}

// StoreAsPermanent stores the permanent token for a registered device.
func (s *CredentialStore) StoreAsPermanent(deviceID uuid.UUID, permanentToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.permanentTokens[deviceID] = permanentToken
}

// GetPermanentToken retrieves the permanent token for a device, if registered.
func (s *CredentialStore) GetPermanentToken(deviceID uuid.UUID) (*string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.permanentTokens[deviceID]
	if !ok {
		return nil, false
	}
	copy := token // avoid exposing internal reference
	return &copy, true
}

// DeleteInitialToken removes an initial token after it's been used (optional but recommended).
func (s *CredentialStore) DeleteInitialToken(deviceID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.initialTokens, deviceID.String())
}
