package config

// adoption_io.go is the console's config seam (change console-adoption-controls
// task 1.2): read the engine.adoption block as the file spells it, and write it
// back without touching anything else.
//
// # Why "as the file spells it" is a separate read
//
// The merge path zeroes a refused block (mergeAdoption), which is right for the
// engine — it must not run on half a setting — and catastrophic for an editor:
// a settings screen that round-trips the zeroed struct erases the exclude list
// that was protecting a long-term holding (review round 1, P1-1). So the screen
// reads the raw block plus a verdict, and the zeroing stays where it belongs.
//
// # Why the write is a byte splice
//
// config.json is a hand-edited file. Round-tripping it through the typed File
// would drop every key this build does not know and re-order the rest. The save
// therefore locates the engine.adoption *value span* in the original bytes and
// replaces exactly that; a key this code has never heard of survives to the
// byte (spec: 블록 밖 바이트 보존). Insertion (no engine block yet, or no
// adoption key yet) necessarily adds text, but still leaves every existing
// value's bytes alone.
//
// # Concurrency
//
// One advisory flock (config.json.lock) around the read-modify-rename, plus a
// unique temp file, so two saves serialize instead of installing each other's
// stale reads (review round 1, P2-4). The lock file rides beside the config so
// an isolated --config-dir profile locks its own file.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Path is where this service reads and writes its config file.
func (s *Service) Path() string { return s.path }

// LoadRawEngineAdoption reads the engine.adoption block as written — before the
// merge's zeroing — plus the validation verdict ("" when the block is usable).
// A missing file is the zero block with no verdict.
func (s *Service) LoadRawEngineAdoption() (Adoption, string, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Adoption{}, "", nil
	}
	if err != nil {
		return Adoption{}, "", err
	}

	var doc struct {
		Engine struct {
			Adoption rawAdoption `json:"adoption"`
		} `json:"engine"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return Adoption{}, "", fmt.Errorf("config: parsing %s: %w", s.path, err)
	}

	raw := doc.Engine.Adoption
	block := Adoption{DefaultStopPct: raw.DefaultStopPct}
	if raw.Enabled != nil {
		block.Enabled = *raw.Enabled
	}
	block.ExcludeSymbols = normaliseSymbols(raw.ExcludeSymbols)
	block.IncludeSymbols = normaliseSymbols(raw.IncludeSymbols)
	return block, block.validate(), nil
}

// SaveEngineAdoption writes the block into config.json, surgically.
//
// A block the merge would refuse is refused here instead: writing it would
// leave a file that says on and an engine that runs off, with nobody told
// (spec: 엔진이 zeroing할 블록을 기록하는 것은 금지된다).
func (s *Service) SaveEngineAdoption(a Adoption) error {
	next := Adoption{Enabled: a.Enabled, DefaultStopPct: a.DefaultStopPct}
	next.ExcludeSymbols = normaliseSymbols(a.ExcludeSymbols)
	next.IncludeSymbols = normaliseSymbols(a.IncludeSymbols)
	next.Rejected = ""
	if why := next.validate(); why != "" {
		return fmt.Errorf("config: refusing to save an adoption block the engine would zero: %s", why)
	}

	lock, err := acquireConfigLock(s.path + ".lock")
	if err != nil {
		return err
	}
	defer lock.release()

	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Skeleton creation is for a file that does not exist, only. A file that
		// exists but cannot be parsed is somebody's edit or damage — refused below.
		file := DefaultFile()
		file.Engine.Adoption = next
		skeleton, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return err
		}
		return s.installBytes(append(skeleton, '\n'))
	case err != nil:
		return err
	}

	if !json.Valid(data) {
		return fmt.Errorf("config: %s is not valid JSON; refusing to overwrite a file that may be "+
			"a hand-edit in progress", s.path)
	}

	block, err := json.Marshal(next)
	if err != nil {
		return err
	}

	out, err := spliceAdoption(data, block)
	if err != nil {
		return err
	}
	return s.installBytes(out)
}

// installBytes writes atomically: a unique temp file in the same directory,
// then rename.
func (s *Service) installBytes(data []byte) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Chmod(name, 0o600); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Rename(name, s.path)
}

// spliceAdoption replaces (or inserts) the engine.adoption value in data.
func spliceAdoption(data, block []byte) ([]byte, error) {
	aStart, aEnd, found, err := adoptionValueSpan(data)
	if err != nil {
		return nil, err
	}
	if found {
		out := make([]byte, 0, len(data)+len(block))
		out = append(out, data[:aStart]...)
		out = append(out, block...)
		out = append(out, data[aEnd:]...)
		return out, nil
	}

	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil {
		return nil, err
	}
	if eFound {
		// Insert the adoption key just inside the engine object.
		return insertKey(data, eStart, eEnd, "adoption", block)
	}

	// No engine block at all: insert one inside the top-level object.
	engine := append([]byte(`{"adoption": `), block...)
	engine = append(engine, '}')
	return insertKey(data, 0, int64(len(data)), "engine", engine)
}

// insertKey inserts `"name": value` inside the object whose span is
// [objStart, objEnd) of data. The object's existing bytes are preserved; only
// the new member (and a comma, when the object is not empty) is added.
func insertKey(data []byte, objStart, objEnd int64, name string, value []byte) ([]byte, error) {
	if objStart >= int64(len(data)) || data[objStart] != '{' {
		return nil, fmt.Errorf("config: the %s block is not an object; refusing to rewrite it", name)
	}
	// Empty when nothing but whitespace sits between the braces.
	empty := true
	for _, b := range data[objStart+1 : objEnd-1] {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			empty = false
			break
		}
	}
	member := append([]byte(`"`+name+`": `), value...)
	if !empty {
		member = append(member, ',', ' ')
	}
	out := make([]byte, 0, len(data)+len(member))
	out = append(out, data[:objStart+1]...)
	out = append(out, member...)
	out = append(out, data[objStart+1:]...)
	return out, nil
}

// adoptionValueSpan locates the byte span of the engine.adoption value.
func adoptionValueSpan(data []byte) (start, end int64, found bool, err error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil || !eFound {
		return 0, 0, false, err
	}
	aStart, aEnd, aFound, err := valueSpan(data[eStart:eEnd], "adoption")
	if err != nil || !aFound {
		return 0, 0, false, err
	}
	return eStart + aStart, eStart + aEnd, true, nil
}

// valueSpan finds the byte span of the value for a top-level key of the object
// in data. Offsets bound exactly the value's own bytes.
func valueSpan(data []byte, key string) (start, end int64, found bool, err error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return 0, 0, false, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return 0, 0, false, errors.New("config: the document is not a JSON object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return 0, 0, false, err
		}
		name, _ := keyTok.(string)
		after := dec.InputOffset()
		if err := skipValue(dec); err != nil {
			return 0, 0, false, err
		}
		if name != key {
			continue
		}
		valueEnd := dec.InputOffset()
		// The value begins after the colon and any whitespace that follows the
		// key token.
		s := after
		for s < valueEnd {
			switch data[s] {
			case ':', ' ', '\t', '\n', '\r':
				s++
				continue
			}
			break
		}
		return s, valueEnd, true, nil
	}
	return 0, 0, false, nil
}

// skipValue consumes one complete JSON value from the decoder.
func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); ok && (d == '{' || d == '[') {
		depth := 1
		for depth > 0 {
			t, err := dec.Token()
			if err != nil {
				return err
			}
			if dd, ok := t.(json.Delim); ok {
				switch dd {
				case '{', '[':
					depth++
				case '}', ']':
					depth--
				}
			}
		}
	}
	return nil
}
