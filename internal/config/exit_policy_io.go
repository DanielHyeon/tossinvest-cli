package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func (s *Service) LoadRawEngineExitPolicy() (ExitPolicy, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return ExitPolicy{}, nil
	}
	if err != nil {
		return ExitPolicy{}, err
	}
	var doc struct {
		Engine struct {
			ExitPolicy rawExitPolicy `json:"exit_policy"`
		} `json:"engine"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return ExitPolicy{}, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}
	out := ExitPolicy{CommonPolicy: doc.Engine.ExitPolicy.CommonPolicy}
	out.Rejected = out.validate()
	return out, nil
}

func (s *Service) SaveEngineExitPolicy(policy ExitPolicy) error {
	next := ExitPolicy{CommonPolicy: policy.CommonPolicy}
	next.CommonPolicy = strings.TrimSpace(next.CommonPolicy)
	if why := next.validate(); why != "" {
		return fmt.Errorf("config: refusing to save exit policy: %s", why)
	}

	lock, err := acquireConfigLock(s.path + ".lock")
	if err != nil {
		return err
	}
	defer lock.release()

	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		file := DefaultFile()
		file.Engine.ExitPolicy = next
		skeleton, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return err
		}
		return s.installBytes(append(skeleton, '\n'))
	case err != nil:
		return err
	}
	if !json.Valid(data) {
		return fmt.Errorf("config: %s is not valid JSON; refusing to overwrite it", s.path)
	}
	block, err := json.Marshal(next)
	if err != nil {
		return err
	}
	out, err := spliceExitPolicy(data, block)
	if err != nil {
		return err
	}
	return s.installBytes(out)
}

func spliceExitPolicy(data, block []byte) ([]byte, error) {
	start, end, found, err := exitPolicyValueSpan(data)
	if err != nil {
		return nil, err
	}
	if found {
		out := make([]byte, 0, len(data)+len(block))
		out = append(out, data[:start]...)
		out = append(out, block...)
		out = append(out, data[end:]...)
		return out, nil
	}
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil {
		return nil, err
	}
	if eFound {
		return insertKey(data, eStart, eEnd, "exit_policy", block)
	}
	engine := append([]byte(`{"exit_policy": `), block...)
	engine = append(engine, '}')
	return insertKey(data, 0, int64(len(data)), "engine", engine)
}

func exitPolicyValueSpan(data []byte) (start, end int64, found bool, err error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil || !eFound {
		return 0, 0, false, err
	}
	pStart, pEnd, pFound, err := valueSpan(data[eStart:eEnd], "exit_policy")
	if err != nil || !pFound {
		return 0, 0, false, err
	}
	return eStart + pStart, eStart + pEnd, true, nil
}
