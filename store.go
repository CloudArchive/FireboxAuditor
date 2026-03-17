package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ── Models ───────────────────────────────────────────────────────────────────

type AuditRecord struct {
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	FileName   string      `json:"file_name"`
	DeviceName string      `json:"device_name"`
	Score      int         `json:"score"`
	Report     AuditReport `json:"report"`
	// SSH enrichment data (nil until enriched)
	Enrichment *EnrichData `json:"enrichment,omitempty"`
}

type EnrichData struct {
	SerialNumber string         `json:"serial_number"`
	UpTime       string         `json:"up_time"`
	MemoryUsage  string         `json:"memory_usage"`
	CPUUsage     string         `json:"cpu_usage"`
	FeatureKey   *ParsedFeatureKey `json:"feature_key,omitempty"`
	EnrichedAt   time.Time      `json:"enriched_at"`
}

type ParsedFeatureKey struct {
	Features []FeatureEntry `json:"features"`
	Raw      string         `json:"raw"`
}

type FeatureEntry struct {
	Name       string `json:"name"`
	Expiration string `json:"expiration,omitempty"`
	Active     bool   `json:"active"`
}

// ── Store ────────────────────────────────────────────────────────────────────

const (
	maxAuditsPerUser = 3
	historyDir       = "data/history"
)

var storeMu sync.RWMutex

func userHistoryDir(username string) string {
	return filepath.Join(historyDir, username)
}

func recordPath(username, id string) string {
	return filepath.Join(userHistoryDir(username), id+".json")
}

func ensureUserDir(username string) error {
	return os.MkdirAll(userHistoryDir(username), 0755)
}

// SaveAudit saves a new audit record and rotates old ones (keeps max 3).
func SaveAudit(username string, record *AuditRecord) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	if err := ensureUserDir(username); err != nil {
		return fmt.Errorf("dizin oluşturulamadı: %w", err)
	}

	// Rotate: if already at max, delete the oldest
	existing, err := listAuditsUnsafe(username)
	if err == nil && len(existing) >= maxAuditsPerUser {
		// Sort oldest first
		sort.Slice(existing, func(i, j int) bool {
			return existing[i].CreatedAt.Before(existing[j].CreatedAt)
		})
		for i := 0; i <= len(existing)-maxAuditsPerUser; i++ {
			os.Remove(recordPath(username, existing[i].ID))
		}
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON encode hatası: %w", err)
	}
	return os.WriteFile(recordPath(username, record.ID), data, 0644)
}

// ListAudits returns all audits for a user, newest first.
func ListAudits(username string) ([]AuditRecord, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return listAuditsUnsafe(username)
}

func listAuditsUnsafe(username string) ([]AuditRecord, error) {
	dir := userHistoryDir(username)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []AuditRecord{}, nil
	}
	if err != nil {
		return nil, err
	}

	var records []AuditRecord
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var rec AuditRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}

	// Newest first
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

// GetAudit returns a single audit record by ID.
func GetAudit(username, id string) (*AuditRecord, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	data, err := os.ReadFile(recordPath(username, id))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec AuditRecord
	return &rec, json.Unmarshal(data, &rec)
}

// UpdateEnrichment updates the SSH enrichment data on an existing record.
func UpdateEnrichment(username, id string, enrich *EnrichData) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	data, err := os.ReadFile(recordPath(username, id))
	if err != nil {
		return fmt.Errorf("kayıt bulunamadı: %w", err)
	}
	var rec AuditRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return err
	}
	rec.Enrichment = enrich

	out, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(recordPath(username, id), out, 0644)
}

// DeleteAudit removes an audit record.
func DeleteAudit(username, id string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	path := recordPath(username, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("kayıt bulunamadı")
	}
	return os.Remove(path)
}

// ── Feature Key Parser ───────────────────────────────────────────────────────

func ParseFeatureKey(raw string) *ParsedFeatureKey {
	fk := &ParsedFeatureKey{Raw: raw}
	var current *FeatureEntry

	for _, line := range splitLines(raw) {
		trimmed := trimSpace(line)

		if hasPrefix(trimmed, "Feature:") {
			if current != nil {
				fk.Features = append(fk.Features, *current)
			}
			name := trimSpace(trimmed[len("Feature:"):])
			current = &FeatureEntry{Name: name, Active: true}
			continue
		}

		if current != nil && hasPrefix(trimmed, "Expiration:") {
			exp := trimSpace(trimmed[len("Expiration:"):])
			current.Expiration = exp
			// Mark expired if date is in the past (simple string check)
			if exp == "0" || exp == "None" || exp == "" {
				current.Active = false
			}
		}
	}
	if current != nil {
		fk.Features = append(fk.Features, *current)
	}
	return fk
}

// small helpers to avoid importing strings in parser
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
