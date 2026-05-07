package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func attestationKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

// Attestation records a human acknowledgment of a port listener.
type Attestation struct {
	Port      int       `json:"port"`
	Proto     string    `json:"proto"`
	AttestedBy string   `json:"attested_by"`
	Reason    string    `json:"reason"`
	AttestedAt time.Time `json:"attested_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// IsExpired returns true if the attestation has an expiry that has passed.
func (a Attestation) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// AttestationStore persists attestations keyed by port/proto.
type AttestationStore struct {
	dir string
}

func attestationStoreKey() string { return "attestations" }

// NewAttestationStore creates a new store rooted at dir.
func NewAttestationStore(dir string) *AttestationStore {
	return &AttestationStore{dir: dir}
}

func (s *AttestationStore) path() string {
	return filepath.Join(s.dir, "attestations.json")
}

func (s *AttestationStore) load() (map[string]Attestation, error) {
	data, err := os.ReadFile(s.path())
	if os.IsNotExist(err) {
		return map[string]Attestation{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m map[string]Attestation
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *AttestationStore) save(m map[string]Attestation) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0o644)
}

// Set stores an attestation for the given port/proto.
func (s *AttestationStore) Set(a Attestation) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[attestationKey(a.Port, a.Proto)] = a
	return s.save(m)
}

// Get retrieves an attestation. Returns false if not found.
func (s *AttestationStore) Get(port int, proto string) (Attestation, bool, error) {
	m, err := s.load()
	if err != nil {
		return Attestation{}, false, err
	}
	a, ok := m[attestationKey(port, proto)]
	return a, ok, nil
}

// Remove deletes an attestation.
func (s *AttestationStore) Remove(port int, proto string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	delete(m, attestationKey(port, proto))
	return s.save(m)
}

// All returns all attestations sorted by port.
func (s *AttestationStore) All() ([]Attestation, error) {
	m, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]Attestation, 0, len(m))
	for _, a := range m {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out, nil
}

// IsAttested returns true if the port/proto has a valid, non-expired attestation.
func (s *AttestationStore) IsAttested(port int, proto string) (bool, error) {
	a, ok, err := s.Get(port, proto)
	if err != nil || !ok {
		return false, err
	}
	return !a.IsExpired(), nil
}
