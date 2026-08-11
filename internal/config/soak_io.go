package config

// soak_io.go is the operator's process-lifecycle approval for the read-only
// capability survey (openspec change a101).
//
// It is engine.autostart's shape, deliberately and completely: same default,
// same surgical write, same separation. The reason it is a second key rather
// than a reuse of the engine's is that the two answer different questions and
// neither implies the other — a machine may be approved to bring its engine up
// after a reboot without being the machine that runs the survey, and far more
// commonly the reverse, because the survey has to keep running for the
// attestation to stay renewable whether or not anything trades.
//
// # Why an approval at all, for a tool that cannot trade
//
// The survey issues nothing but GETs and structurally cannot reach a mutation
// (internal/soak's package doc and its import-graph test). So this is not a
// capability grant, and audit.ActionEngineAutostart already says what that makes
// it: "the separate process-lifecycle approval. It does not grant order
// capability". Starting a process on somebody's machine at boot is a decision
// they get to record, and the record is what survives the container it ran in.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// LoadSoakAutostart reads the operator's approval to start the survey when the
// console starts.
//
// Missing files, missing soak blocks and missing keys all mean false, which is
// exactly the behaviour that existed before this key did. A file that cannot be
// parsed is an error and not a false: "there is no approval here" and "this
// configuration is unreadable" are different facts, and the second one is worth
// telling the operator about.
func (s *Service) LoadSoakAutostart() (bool, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var doc struct {
		Soak struct {
			Autostart bool `json:"autostart"`
		} `json:"soak"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return false, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}
	return doc.Soak.Autostart, nil
}

// SaveSoakAutostart writes soak.autostart and no other key.
func (s *Service) SaveSoakAutostart(on bool) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		return spliceMembers(data, soakValueSpan, insertEmptySoak,
			[]gateMember{{"autostart", []byte(boolLiteral(on))}})
	})
}

func soakValueSpan(data []byte) (start, end int64, found bool, err error) {
	return valueSpan(data, "soak")
}

// insertEmptySoak creates the top-level soak block. The caller splices the one
// requested key afterwards, exactly as insertEmptyEngine's caller does.
func insertEmptySoak(data []byte) ([]byte, error) {
	return insertKey(data, 0, int64(len(data)), "soak", []byte("{}"))
}
