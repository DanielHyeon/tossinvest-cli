//go:build unix

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// unixReceipt owns a single freshly created 0700 receipt directory below a
// verified /tmp root. Every child is created through its directory FD.
type unixReceipt struct {
	rootFD       int
	runFD        int
	files        map[string]string
	manifest     []byte
	live         func() error
	payloadLimit int64
}

const maxReceiptPayloadBytes int64 = 64 << 20

func openReceipt(root string) (receiptStore, error) {
	if !isTemporaryRoot(root) {
		return nil, errors.New("receipt root must be beneath /tmp")
	}
	rootFD, err := openPrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	var filesystem unix.Statfs_t
	if err := unix.Fstatfs(rootFD, &filesystem); err != nil {
		_ = unix.Close(rootFD)
		return nil, err
	}
	// FUSE cannot provide the POSIX ownership semantics required for raw bodies.
	if uint64(filesystem.Type) == 0x65735546 { // FUSE_SUPER_MAGIC
		_ = unix.Close(rootFD)
		return nil, errors.New("FUSE receipt root is not accepted")
	}
	for attempt := 0; attempt < 5; attempt++ {
		name, err := receiptDirectoryName()
		if err != nil {
			_ = unix.Close(rootFD)
			return nil, err
		}
		if err := unix.Mkdirat(rootFD, name, 0o700); err != nil {
			if errors.Is(err, unix.EEXIST) {
				continue
			}
			_ = unix.Close(rootFD)
			return nil, err
		}
		runFD, err := unix.Openat(rootFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if err != nil {
			_ = unix.Close(rootFD)
			return nil, err
		}
		if err := verifyPrivateDirectory(runFD); err != nil {
			_ = unix.Close(runFD)
			_ = unix.Close(rootFD)
			return nil, err
		}
		if err := unix.Fsync(rootFD); err != nil {
			_ = unix.Close(runFD)
			_ = unix.Close(rootFD)
			return nil, err
		}
		return &unixReceipt{rootFD: rootFD, runFD: runFD, files: make(map[string]string), payloadLimit: maxReceiptPayloadBytes}, nil
	}
	_ = unix.Close(rootFD)
	return nil, errors.New("could not create an exclusive receipt directory")
}

func isTemporaryRoot(root string) bool {
	clean := filepath.Clean(root)
	if !filepath.IsAbs(clean) {
		return false
	}
	rel, err := filepath.Rel("/tmp", clean)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func receiptDirectoryName() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "a112-mb-us-" + hex.EncodeToString(bytes), nil
}

func openPrivateDirectory(path string) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return -1, err
	}
	components := strings.Split(strings.TrimPrefix(filepath.Clean(abs), string(filepath.Separator)), string(filepath.Separator))
	fd, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, err
	}
	for index, component := range components {
		if component == "" || component == "." || component == ".." {
			_ = unix.Close(fd)
			return -1, errors.New("unsafe receipt root component")
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, openErr
		}
		fd = next
		// Ancestors are opened no-follow to pin path resolution. Only the exact
		// configured root is the raw-body capability and therefore must be 0700.
		if index == len(components)-1 {
			if err := verifyPrivateDirectory(fd); err != nil {
				_ = unix.Close(fd)
				return -1, err
			}
		}
	}
	return fd, nil
}

func verifyPrivateDirectory(fd int) error {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o777 != 0o700 {
		return errors.New("receipt directory is not current-UID exact 0700")
	}
	return nil
}

func (r *unixReceipt) writeJSON(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.write(name, append(data, '\n'))
}

func (r *unixReceipt) writeResult(prefix string, result sourceResult) error {
	if !safeRawBody(result.body) {
		return errors.New("raw body contains a secret-like field or malformed JSON")
	}
	if !safeCanonicalQuery(result.query) {
		return errors.New("canonical query contains unsafe material")
	}
	if err := r.write(prefix+".raw.json", result.body); err != nil {
		return err
	}
	if len(result.cursorJSON) != 0 {
		if err := r.write(prefix+".cursor.raw.json", result.cursorJSON); err != nil {
			return err
		}
	}
	headers, err := json.Marshal(result.rateHeaders)
	if err != nil {
		return err
	}
	if err := r.write(prefix+".headers.raw.json", append(headers, '\n')); err != nil {
		return err
	}
	meta, err := marshalReceiptMeta(result)
	if err != nil {
		return err
	}
	return r.write(prefix+".meta.json", append(meta, '\n'))
}

func (r *unixReceipt) write(name string, data []byte) error {
	if r.live != nil {
		if err := r.live(); err != nil {
			return err
		}
	}
	if !validReceiptName(name) {
		return errors.New("invalid receipt filename")
	}
	fd, err := unix.Openat(r.runFD, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		_ = unix.Close(fd)
		return errors.New("receipt descriptor could not become file")
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o777 != 0o600 {
		return errors.New("receipt payload is not current-UID exact 0600")
	}
	if err := unix.Fsync(r.runFD); err != nil {
		return err
	}
	r.files[name] = digestBytes(data)
	if r.live != nil {
		if err := r.live(); err != nil {
			return err
		}
	}
	return nil
}

func (r *unixReceipt) setLive(live func() error) { r.live = live }

func validReceiptName(name string) bool {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.HasPrefix(name, ".") {
		return false
	}
	for _, char := range name {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '.' {
			return false
		}
	}
	return true
}

func (r *unixReceipt) seal(seal receiptSeal) error {
	if err := r.verifyExpected(false); err != nil {
		return err
	}
	names := make([]string, 0, len(r.files))
	for name := range r.files {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]receiptManifestEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, receiptManifestEntry{Name: name, SHA256: r.files[name]})
	}
	manifest, err := json.Marshal(receiptManifest{Schema: "a112-mb-us-receipt:v1", Identity: seal, Files: entries})
	if err != nil {
		return err
	}
	manifest = append(manifest, '\n')
	if err := r.write("manifest.json", manifest); err != nil {
		return err
	}
	delete(r.files, "manifest.json")
	r.manifest = append([]byte(nil), manifest...)
	return r.verifyExpected(true)
}

func (r *unixReceipt) verifySealed() error { return r.verifyExpected(true) }

func (r *unixReceipt) taint() error {
	if _, exists := r.files["TAINTED"]; exists {
		return nil
	}
	live := r.live
	r.live = nil
	defer func() { r.live = live }()
	return r.write("TAINTED", []byte("HOLD\n"))
}

func (r *unixReceipt) verifyExpected(withManifest bool) error {
	if err := verifyPrivateDirectory(r.rootFD); err != nil {
		return errors.New("receipt root mode changed")
	}
	if err := verifyPrivateDirectory(r.runFD); err != nil {
		return errors.New("receipt run directory mode changed")
	}
	fd, err := unix.Openat(r.runFD, ".", unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(fd), "receipt-dir")
	if dir == nil {
		_ = unix.Close(fd)
		return errors.New("receipt directory descriptor is invalid")
	}
	entries, readErr := dir.ReadDir(-1)
	closeErr := dir.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	want := make(map[string]struct{}, len(r.files)+1)
	for name := range r.files {
		want[name] = struct{}{}
	}
	if withManifest {
		want["manifest.json"] = struct{}{}
	}
	if len(entries) != len(want) {
		return errors.New("receipt directory entry set changed")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return errors.New("receipt directory contains nested entry")
		}
		if _, exists := want[entry.Name()]; !exists {
			return errors.New("receipt directory contains unexpected entry")
		}
		var stat unix.Stat_t
		if err := unix.Fstatat(r.runFD, entry.Name(), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return errors.New("receipt payload metadata is unavailable")
		}
		if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Mode&0o777 != 0o600 || stat.Uid != uint32(os.Geteuid()) {
			return errors.New("receipt payload mode changed")
		}
	}
	for name, wantDigest := range r.files {
		data, err := r.readRegular(name)
		if err != nil || digestBytes(data) != wantDigest {
			return errors.New("receipt payload content changed")
		}
	}
	if withManifest {
		data, err := r.readRegular("manifest.json")
		if err != nil || len(r.manifest) == 0 || !strings.EqualFold(digestBytes(data), digestBytes(r.manifest)) || !bytesEqual(data, r.manifest) {
			return errors.New("receipt manifest content changed")
		}
		var parsed receiptManifest
		if !strictUnmarshal(data, &parsed) || parsed.Schema != "a112-mb-us-receipt:v1" || len(parsed.Files) != len(r.files) {
			return errors.New("receipt manifest is invalid")
		}
		for index, item := range parsed.Files {
			if index > 0 && parsed.Files[index-1].Name >= item.Name {
				return errors.New("receipt manifest entries are not canonical")
			}
			if item.Name == "manifest.json" || r.files[item.Name] != item.SHA256 {
				return errors.New("receipt manifest digest set changed")
			}
		}
	}
	if err := verifyPrivateDirectory(r.rootFD); err != nil {
		return errors.New("receipt root mode changed during verification")
	}
	if err := verifyPrivateDirectory(r.runFD); err != nil {
		return errors.New("receipt run directory mode changed during verification")
	}
	return nil
}

type receiptManifestEntry struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type receiptManifest struct {
	Schema   string                 `json:"schema"`
	Identity receiptSeal            `json:"identity"`
	Files    []receiptManifestEntry `json:"files"`
}

func (r *unixReceipt) readRegular(name string) ([]byte, error) {
	return r.readRegularWithLimit(name, r.payloadLimit)
}

func (r *unixReceipt) readRegularWithLimit(name string, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, errors.New("receipt payload limit is invalid")
	}
	fd, err := unix.Openat(r.runFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		_ = unix.Close(fd)
		return nil, errors.New("receipt payload descriptor is invalid")
	}
	defer file.Close()
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o777 != 0o600 || stat.Size < 0 {
		return nil, errors.New("receipt payload is not a private regular file")
	}
	if stat.Size > limit {
		return nil, errors.New("receipt payload exceeds size limit")
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("receipt payload exceeds size limit")
	}
	var after unix.Stat_t
	if err := unix.Fstat(fd, &after); err != nil || after.Mode&unix.S_IFMT != unix.S_IFREG || after.Uid != uint32(os.Geteuid()) || after.Mode&0o777 != 0o600 || after.Dev != stat.Dev || after.Ino != stat.Ino || after.Size != stat.Size || int64(len(data)) != after.Size {
		return nil, errors.New("receipt payload changed or truncated during read")
	}
	return data, nil
}

func bytesEqual(left, right []byte) bool {
	return len(left) == len(right) && digestBytes(left) == digestBytes(right)
}

func (r *unixReceipt) close() error {
	var errorsOut []error
	if r.runFD >= 0 {
		errorsOut = append(errorsOut, unix.Close(r.runFD))
		r.runFD = -1
	}
	if r.rootFD >= 0 {
		errorsOut = append(errorsOut, unix.Close(r.rootFD))
		r.rootFD = -1
	}
	return errors.Join(errorsOut...)
}

func safeRawBody(data []byte) bool {
	value, ok := decodeStrictJSON(data)
	if !ok {
		return false
	}
	return noSecretLikeFields(value)
}

var surrogateEscape = regexp.MustCompile(`(?i)\\uD[89A-F][0-9A-F]{2}`)

func decodeStrictJSON(data []byte) (any, bool) {
	if !utf8.Valid(data) || surrogateEscape.Match(data) {
		return nil, false
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder)
	if err != nil {
		return nil, false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, false
	}
	return value, true
}

func decodeStrictValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return token, nil
	}
	switch delim {
	case '{':
		object := map[string]any{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, errors.New("JSON object key is not a string")
			}
			if _, duplicate := object[key]; duplicate {
				return nil, errors.New("duplicate JSON object key")
			}
			child, err := decodeStrictValue(decoder)
			if err != nil {
				return nil, err
			}
			object[key] = child
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return nil, errors.New("unterminated JSON object")
		}
		return object, nil
	case '[':
		array := []any{}
		for decoder.More() {
			child, err := decodeStrictValue(decoder)
			if err != nil {
				return nil, err
			}
			array = append(array, child)
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return nil, errors.New("unterminated JSON array")
		}
		return array, nil
	default:
		return nil, errors.New("unexpected JSON delimiter")
	}
}

func strictUnmarshal(data []byte, target any) bool {
	if _, ok := decodeStrictJSON(data); !ok {
		return false
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target) == nil
}

func safeCanonicalQuery(query string) bool {
	values, err := url.ParseQuery(query)
	if err != nil {
		return false
	}
	for key, items := range values {
		if secretLikeKey(key) || len(items) != 1 || secretLikeValue(items[0]) {
			return false
		}
	}
	return true
}

func noSecretLikeFields(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if secretLikeKey(key) {
				return false
			}
			if !noSecretLikeFields(child) {
				return false
			}
		}
	case []any:
		for _, child := range typed {
			if !noSecretLikeFields(child) {
				return false
			}
		}
	case string:
		return !secretLikeValue(typed)
	}
	return true
}

func secretLikeKey(key string) bool {
	var normalized strings.Builder
	for _, char := range strings.ToLower(key) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			normalized.WriteRune(char)
		}
	}
	compact := normalized.String()
	for _, marker := range []string{"authorization", "apikey", "token", "secret", "cookie", "password", "credential", "session", "accesskey", "privatekey", "account"} {
		if strings.Contains(compact, marker) {
			return true
		}
	}
	return false
}

var (
	bearerValue = regexp.MustCompile(`(?i)^\s*bearer\s+[a-z0-9._~+/=-]{8,}\s*$`)
	jwtValue    = regexp.MustCompile(`^[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{6,}$`)
	apiKeyValue = regexp.MustCompile(`(?i)^(?:sk|pk|api|token|key)[_-][A-Za-z0-9_-]{12,}$`)
)

func secretLikeValue(value string) bool {
	return bearerValue.MatchString(value) || jwtValue.MatchString(value) || apiKeyValue.MatchString(value)
}

var _ receiptStore = (*unixReceipt)(nil)
