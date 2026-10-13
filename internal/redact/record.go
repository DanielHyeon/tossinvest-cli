package redact

// record.go — soak·검증 기록 줄의 계좌 봉인(design D4).
//
// 두 기록의 `account_ref` 는 로그이면서 인터록 증거의 계좌 결속 입력임(attestation 계좌의 출처 · 감독 증거 마스크 비교 ·
// 검증 대사 계좌 결속). 지우면 결속이 사라지고, 소금 없는 해시는 11자리 대입으로 되돌릴 수 있음. 그래서 로컬 키로
// 봉인(AES-256-GCM, 무작위 nonce)해 쓰고 읽는 쪽이 열어 종전과 같은 값을 봄. 키 파일은 계좌 원문을 이미 든
// attestation·원장과 같은 디렉터리(0600)에 있으므로 키를 가진 사람은 이미 계좌를 가짐 — 봉인은 「로그만 샌 경우」를 막음.
//
// 쓰기 경로(Recorder.Append · 대사 appendLine)와 일회성 정리 명령은 **같은 함수** SealRecordLine 을 부름.

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// PlainField 는 a131 이전 판본이 계좌를 평문으로 쓰던 필드임.
	PlainField = "account_ref"
	// SealedField 는 봉인된 계좌 필드임.
	SealedField = "account_ref_sealed"
	// KeyFileName 은 기록 디렉터리 안 봉인 키 파일 이름임.
	KeyFileName = "account-seal.key"

	sealPrefix = "v1."
	keyBytes   = 32
)

// sealAAD 는 봉인 값을 이 용도(기록의 계좌 필드)에 묶음 — 다른 곳의 같은 키 암호문을 여기 옮겨도 열리지 않게.
var sealAAD = []byte("tossos/a131/account_ref/v1")

// SealedFormatVersion 은 봉인 줄의 기록 형식 번호임. soak · verifylive 의 RecordFormatVersion 이 이 값과 같아야 함(시험이 고정) —
// 형식 1 까지만 아는 옛 읽기는 봉인 줄을 「newer format」으로 거절함(평문 옛 줄 + 봉인 새 줄로 발급하는 길을 닫음).
const SealedFormatVersion = 2

// SealKey 는 한 기록 디렉터리의 준비된 봉인 키임. PrepareSealKey 로만 만듦.
type SealKey struct{ key []byte }

// PrepareSealKey 는 기록 recordPath 옆의 봉인 키를 읽거나 만들고 시험 봉인을 한 번 해 봄. 기록을 여는 시점(브로커 호출 전)에
// 부르는 것이 목적 — 키 실패가 실주문 뒤의 Append 에서 드러나지 않게.
//
// 키가 없는데 기록에 이미 봉인 줄이 있으면 만들지 않고 거절함 — 새 키로 쓰면 옛 줄은 영영 열리지 않고, 새 줄만 계좌를 가진 기록이 됨.
func PrepareSealKey(recordPath string) (*SealKey, error) {
	dir := filepath.Dir(recordPath)
	key, err := loadKey(dir, false)
	if errors.Is(err, fs.ErrNotExist) {
		// 같은 디렉터리의 기록들(soak · verify KR · verify US)은 키 하나를 나눠 씀 — 어느 기록에든 봉인 줄이 있으면 새 키를 만들지 않음
		// (구현 리뷰 I-5: 여는 기록 하나만 보면 다른 기록의 봉인 줄이 영영 안 열림).
		if sealed := sealedRecordIn(dir, recordPath); sealed != "" {
			return nil, fmt.Errorf("redact: %s has sealed lines but the seal key %s is missing — refusing to start a new key "+
				"(the old lines could never be opened again); restore the key or start fresh records",
				sealed, filepath.Join(dir, KeyFileName))
		}
		key, err = loadKey(dir, true)
	}
	if err != nil {
		return nil, err
	}
	probe, err := seal(key, "probe")
	if err != nil {
		return nil, err
	}
	if got, ok := open(key, probe); !ok || got != "probe" {
		return nil, fmt.Errorf("redact: the seal key in %s does not round-trip", dir)
	}
	return &SealKey{key: key}, nil
}

// sealedRecordIn 은 dir 안 기록 파일(recordPath 와 *.jsonl) 중 봉인 줄이 있는 첫 경로를 돌려줌. 없으면 "".
func sealedRecordIn(dir, recordPath string) string {
	candidates, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	candidates = append([]string{recordPath}, candidates...)
	for _, c := range candidates {
		if data, err := os.ReadFile(c); err == nil && bytes.Contains(data, []byte(`"`+SealedField+`"`)) {
			return c
		}
	}
	return ""
}

// SealRecordLine 은 기록 한 줄(JSON 객체, 끝 개행 선택)을 규칙에 통과시킴: 평문 `account_ref` 를 떼어 봉인해 `account_ref_sealed` 로
// 두고 `format_version` 을 SealedFormatVersion 이상으로 올리며, 나머지 문자열 값을 String 으로 치환함(식별자·다이제스트 키는 제외).
// 바뀐 것이 없으면 받은 바이트를 그대로 돌려줌. 쓰기 경로(Recorder.Append · appendLine)와 정리 명령이 같은 이 함수를 부름.
func SealRecordLine(line []byte, key *SealKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("redact: no seal key was prepared for this record")
	}
	out, _, err := transformRecordLine(line, func(plain string) (string, error) { return seal(key.key, plain) })
	return out, err
}

// RecordLineNeedsRewrite 는 SealRecordLine 이 이 줄을 바꿀지 판정함 — 키를 읽지도 만들지도 않음(미리보기용, 쓰기 0).
func RecordLineNeedsRewrite(line []byte) (bool, error) {
	_, changed, err := transformRecordLine(line, func(string) (string, error) { return "", nil })
	return changed, err
}

// transformRecordLine 은 봉인 함수만 바꿔 끼우는 공통 변환임 — 쓰기·정리·미리보기가 같은 판정을 씀.
func transformRecordLine(line []byte, sealer func(string) (string, error)) ([]byte, bool, error) {
	body := bytes.TrimRight(line, "\r\n")
	trailing := line[len(body):]
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // 숫자 원문 보존(정밀도)
	var obj map[string]any
	if err := dec.Decode(&obj); err != nil {
		return line, false, fmt.Errorf("redact: the record line is not a JSON object: %w", err)
	}
	if obj == nil {
		return line, false, errors.New("redact: the record line is not a JSON object")
	}
	// 줄 하나는 객체 하나 — 붙은 두 번째 객체나 꼬리 쓰레기를 버리면 읽는 쪽(json.Unmarshal)이 거절하던 줄이 줄 하나를 잃은 채
	// 읽히게 됨(구현 리뷰 I-1). 그런 줄은 해독 불가로 두고 그대로 남김.
	if _, err := dec.Token(); err != io.EOF {
		return line, false, errors.New("redact: the record line holds more than one JSON value")
	}
	changed := false
	if plain, ok := obj[PlainField].(string); ok && strings.TrimSpace(plain) != "" {
		sealed, err := sealer(plain)
		if err != nil {
			return line, false, err
		}
		delete(obj, PlainField)
		obj[SealedField] = sealed
		if v, ok := obj["format_version"].(json.Number); ok {
			if n, err := v.Int64(); err == nil && n < SealedFormatVersion {
				obj["format_version"] = json.Number(fmt.Sprint(SealedFormatVersion))
			}
		}
		changed = true
	}
	for k, v := range obj {
		if k == SealedField || identityKey(k) {
			continue
		}
		if s, c := scrubValue(v); c {
			obj[k] = s
			changed = true
		}
	}
	if !changed {
		return line, false, nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // 원문의 <, >, & 를 그대로 둠
	if err := enc.Encode(obj); err != nil {
		return line, false, fmt.Errorf("redact: re-encoding the record line: %w", err)
	}
	out := bytes.TrimRight(buf.Bytes(), "\n")
	return append(append([]byte(nil), out...), trailing...), true, nil
}

// scrubValue 는 디코드된 JSON 값 안의 모든 문자열을 String 으로 치환함. 두 번째 반환은 바뀌었는지.
func scrubValue(v any) (any, bool) {
	switch x := v.(type) {
	case string:
		s := String(x)
		return s, s != x
	case map[string]any:
		changed := false
		for k, e := range x {
			if identityKey(k) {
				continue
			}
			if s, c := scrubValue(e); c {
				x[k] = s
				changed = true
			}
		}
		return x, changed
	case []any:
		changed := false
		for i, e := range x {
			if s, c := scrubValue(e); c {
				x[i] = s
				changed = true
			}
		}
		return x, changed
	default:
		return v, false
	}
}

// identityKey 는 값이 식별자·다이제스트인 키임 — 치환하지 않음(대사·정리·재생이 이 값으로 대상을 찾음). 식별자에는 계좌가 들어가지 않음.
func identityKey(k string) bool {
	k = strings.ToLower(k)
	return strings.HasSuffix(k, "id") || strings.HasSuffix(k, "ids") || strings.Contains(k, "digest")
}

// Opener 는 한 기록 디렉터리의 봉인을 여는 것임. 키는 처음 한 번만 읽음(줄마다 파일을 열지 않게). 키를 만들지 않음.
type Opener struct {
	dir    string
	key    []byte
	loaded bool
}

// NewOpener 는 dir 의 봉인 키로 여는 Opener 를 만듦.
func NewOpener(dir string) *Opener { return &Opener{dir: dir} }

// Open 은 봉인 값을 연다. 키가 없거나 값이 위조·손상이면 ("", false) — 읽는 쪽은 계좌가 없는 것으로 판정(fail-closed).
func (o *Opener) Open(sealed string) (string, bool) {
	if o == nil || strings.TrimSpace(sealed) == "" {
		return "", false
	}
	if !o.loaded {
		o.loaded = true
		if k, err := loadKey(o.dir, false); err == nil {
			o.key = k
		}
	}
	if o.key == nil {
		return "", false
	}
	return open(o.key, sealed)
}

// OpenAccount 는 dir 의 키로 봉인 값 하나를 연다.
func OpenAccount(sealed, dir string) (string, bool) { return NewOpener(dir).Open(sealed) }

func seal(key []byte, plain string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("redact: crypto/rand is unavailable: %w", err)
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), sealAAD)
	return sealPrefix + base64.RawURLEncoding.EncodeToString(out), nil
}

func open(key []byte, sealed string) (string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(strings.TrimSpace(sealed), sealPrefix))
	if err != nil || !strings.HasPrefix(strings.TrimSpace(sealed), sealPrefix) {
		return "", false
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", false
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(raw) < gcm.NonceSize() {
		return "", false
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], sealAAD)
	if err != nil {
		return "", false
	}
	return string(plain), true
}

// loadKey 는 dir 의 봉인 키를 읽음. create 면 없을 때 만듦: 임시 파일(0600)에 쓰고 동기화한 뒤 link 로 게시 —
// 두 프로세스가 동시에 만들면 먼저 게시한 쪽 키를 둘 다 씀(rename 은 덮어쓰므로 쓰지 않음).
func loadKey(dir string, create bool) ([]byte, error) {
	path := filepath.Join(dir, KeyFileName)
	if key, err := readKey(path); err == nil {
		return key, nil
	} else if !errors.Is(err, fs.ErrNotExist) || !create {
		return nil, err
	}
	key := make([]byte, keyBytes)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("redact: crypto/rand is unavailable: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".account-seal-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("redact: creating the seal key in %s: %w", dir, err)
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("redact: securing the seal key: %w", err)
	}
	if _, err := tmp.WriteString(hex.EncodeToString(key) + "\n"); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("redact: writing the seal key: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("redact: syncing the seal key: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("redact: closing the seal key: %w", err)
	}
	if err := os.Link(tmp.Name(), path); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return readKey(path)
		}
		return nil, fmt.Errorf("redact: publishing the seal key %s: %w", path, err)
	}
	// 디렉터리 항목도 동기화 — 정전 뒤 봉인 줄만 남고 키가 없으면 그 기록은 영영 못 염(구현 리뷰 I-4).
	if d, err := os.Open(dir); err == nil {
		serr := d.Sync()
		_ = d.Close()
		if serr != nil {
			return nil, fmt.Errorf("redact: syncing %s after publishing the seal key: %w", dir, serr)
		}
	}
	return key, nil
}

func readKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	key, err := hex.DecodeString(strings.TrimSpace(string(data)))
	if err != nil || len(key) != keyBytes {
		return nil, fmt.Errorf("redact: the seal key %s is malformed", path)
	}
	return key, nil
}
