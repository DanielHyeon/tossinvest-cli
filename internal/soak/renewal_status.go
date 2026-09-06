package soak

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const RenewalStatusSuffix = ".renewal-status.json"
const renewalStatusMaxBytes = 4096

var createRenewalStatusTemp = os.CreateTemp

type RenewalOutcome string

const (
	RenewalIssued  RenewalOutcome = "issued"
	RenewalRefused RenewalOutcome = "refused"
	RenewalFailed  RenewalOutcome = "failed"
)

type RenewalReasonCode string

const (
	ReasonRecordEmpty      RenewalReasonCode = "record_empty"
	ReasonRecordStale      RenewalReasonCode = "record_stale"
	ReasonAccountAmbiguous RenewalReasonCode = "account_ambiguous"
	ReasonAccountMissing   RenewalReasonCode = "account_missing"
	ReasonStreak           RenewalReasonCode = "insufficient_streak"
	ReasonTokenObservation RenewalReasonCode = "token_observation_missing"
	ReasonTokenRefresh     RenewalReasonCode = "token_refresh_missing"
	ReasonEndpoint         RenewalReasonCode = "required_endpoint_missing"
	ReasonCompleteness     RenewalReasonCode = "completeness_failed"
	ReasonInput            RenewalReasonCode = "input_error"
	ReasonSupervised       RenewalReasonCode = "supervised_evidence_invalid"
)

type RenewalStatus struct {
	FormatVersion int                 `json:"format_version"`
	AttemptedAt   time.Time           `json:"attempted_at"`
	Outcome       RenewalOutcome      `json:"outcome"`
	ReasonCodes   []RenewalReasonCode `json:"reason_codes"`
	ExpiresAt     time.Time           `json:"expires_at,omitempty"`
}

func RenewalStatusPath(attestationPath string) string { return attestationPath + RenewalStatusSuffix }

func (s RenewalStatus) Validate(now time.Time) error {
	if s.FormatVersion != 1 || s.AttemptedAt.IsZero() || s.AttemptedAt.After(now) || !isUTCOffset(s.AttemptedAt) {
		return errors.New("invalid renewal status")
	}
	if s.Outcome != RenewalIssued && s.Outcome != RenewalRefused && s.Outcome != RenewalFailed {
		return errors.New("invalid renewal status")
	}
	if len(s.ReasonCodes) > 16 {
		return errors.New("invalid renewal status")
	}
	seen := map[RenewalReasonCode]bool{}
	for _, code := range s.ReasonCodes {
		if seen[code] || !validReasonCode(code) {
			return errors.New("invalid renewal status")
		}
		seen[code] = true
	}
	if s.Outcome == RenewalIssued && (!s.ReasonCodesEmpty() || s.ExpiresAt.IsZero() || !isUTCOffset(s.ExpiresAt) || !s.ExpiresAt.After(s.AttemptedAt)) {
		return errors.New("invalid renewal status")
	}
	if s.Outcome != RenewalIssued && s.ReasonCodesEmpty() {
		return errors.New("invalid renewal status")
	}
	return nil
}

func isUTCOffset(t time.Time) bool {
	_, offset := t.Zone()
	return offset == 0
}
func (s RenewalStatus) ReasonCodesEmpty() bool { return len(s.ReasonCodes) == 0 }
func validReasonCode(code RenewalReasonCode) bool {
	switch code {
	case ReasonRecordEmpty, ReasonRecordStale, ReasonAccountAmbiguous, ReasonAccountMissing, ReasonStreak, ReasonTokenObservation, ReasonTokenRefresh, ReasonEndpoint, ReasonCompleteness, ReasonInput, ReasonSupervised:
		return true
	}
	return false
}

func SaveRenewalStatus(path string, status RenewalStatus) error {
	if err := status.Validate(time.Now().UTC()); err != nil {
		return err
	}
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := createRenewalStatusTemp(filepath.Dir(path), ".renewal-status-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func LoadRenewalStatus(path string, now time.Time) (RenewalStatus, error) {
	file, err := openRenewalStatus(path)
	if err != nil {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > renewalStatusMaxBytes || !renewalStatusOwnedByCurrentUser(info) {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	data, err := io.ReadAll(io.LimitReader(file, renewalStatusMaxBytes+1))
	if err != nil {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	if len(data) > renewalStatusMaxBytes {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var status RenewalStatus
	if err := dec.Decode(&status); err != nil {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return RenewalStatus{}, errors.New("invalid renewal status")
	}
	if err := status.Validate(now.UTC()); err != nil {
		return RenewalStatus{}, err
	}
	return status, nil
}

func SortedReasonCodes(codes []RenewalReasonCode) []RenewalReasonCode {
	out := append([]RenewalReasonCode(nil), codes...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
