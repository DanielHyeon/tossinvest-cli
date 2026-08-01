package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
)

type BrowserSessionVerifier func(*http.Request) (MutationIdentity, bool)
type BrowserCSRFVerifier func(*http.Request, string) bool

type MutationSecurityOptions struct {
	Boundary       *networkboundary.Boundary
	Ledger         MutationLedger
	Capability     *CapabilityVerifier
	BrowserSession BrowserSessionVerifier
	BrowserCSRF    BrowserCSRFVerifier
	MTLSIdentities map[string]MutationIdentity
	Now            func() time.Time
}

type MutationSecurity struct {
	boundary       *networkboundary.Boundary
	ledger         MutationLedger
	capability     *CapabilityVerifier
	browserSession BrowserSessionVerifier
	browserCSRF    BrowserCSRFVerifier
	mtls           map[string]MutationIdentity
	now            func() time.Time
}

type AuthorizedMutation struct {
	Identity      MutationIdentity
	Method        string
	Resource      string
	CanonicalBody []byte
	BodyDigest    string
	IfMatch       string
}

type MutationResult struct {
	Status  int
	Version string
	Body    []byte
}

// MutationCommander must be narrow and CAS-aware. A non-nil error means the
// command made no state change; ambiguous outcomes must be represented as a
// successful durable result or recovered locally, never retried by this layer.
type MutationCommander interface {
	Execute(context.Context, AuthorizedMutation) (MutationResult, error)
}

type PreconditionError struct {
	CurrentVersion string
}

func (e *PreconditionError) Error() string { return "httpapi: mutation precondition failed" }

// CommandError is a transport-safe semantic refusal from a narrow mutation
// command. Only bad-request and forbidden errors are admitted by the response
// classifier; malformed instances fail closed as COMMAND_UNAVAILABLE.
type CommandError struct {
	status  int
	code    string
	message string
}

func NewCommandError(status int, code, message string) error {
	return &CommandError{status: status, code: code, message: message}
}

func (e *CommandError) Error() string   { return "httpapi: mutation command refused" }
func (e *CommandError) StatusCode() int { return e.status }
func (e *CommandError) ErrorCode() string {
	return e.code
}

var errBrowserAntiForgery = errors.New("httpapi: browser anti-forgery check failed")

func NewMutationSecurity(options MutationSecurityOptions) (*MutationSecurity, error) {
	if options.Boundary == nil || options.Ledger == nil {
		return nil, errors.New("httpapi: mutation security requires network boundary and durable ledger")
	}
	if (options.BrowserSession == nil) != (options.BrowserCSRF == nil) {
		return nil, errors.New("httpapi: browser session and CSRF verifiers must be configured together")
	}
	if options.BrowserSession == nil && options.Capability == nil && len(options.MTLSIdentities) == 0 {
		return nil, errors.New("httpapi: mutation security requires at least one explicit identity mode")
	}
	mtls := make(map[string]MutationIdentity, len(options.MTLSIdentities))
	for fingerprint, identity := range options.MTLSIdentities {
		fingerprint = strings.ToLower(strings.TrimSpace(fingerprint))
		if !validDigest(fingerprint) || !identity.valid() || identity.Mode != AuthModeMTLS {
			return nil, errors.New("httpapi: invalid enrolled mTLS identity")
		}
		mtls[fingerprint] = identity
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	return &MutationSecurity{
		boundary: options.Boundary, ledger: options.Ledger, capability: options.Capability,
		browserSession: options.BrowserSession, browserCSRF: options.BrowserCSRF,
		mtls: mtls, now: options.Now,
	}, nil
}

func (s *MutationSecurity) Handler(resource string, command MutationCommander) http.Handler {
	if s == nil || !validResource(resource) || command == nil {
		return http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.serveMutation(w, r, resource, command)
	})
}

func (s *MutationSecurity) serveMutation(w http.ResponseWriter, r *http.Request, resource string, command MutationCommander) {
	if r.Method != http.MethodPost || r.URL.Path != resource || r.URL.RawQuery != "" {
		writeSecurityError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND")
		return
	}
	if !s.boundary.PeerAllowed(r) || !s.boundary.OriginMatches(r) {
		writeSecurityError(w, http.StatusForbidden, "BOUNDARY_REFUSED")
		return
	}
	contentType, contentTypeOK := exactRequestHeader(r.Header, "Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if !contentTypeOK || err != nil || mediaType != "application/json" || len(params) > 1 ||
		len(params) == 1 && !strings.EqualFold(params["charset"], "utf-8") {
		writeSecurityError(w, http.StatusUnsupportedMediaType, "JSON_REQUIRED")
		return
	}
	if r.ContentLength > MaxRequestBodyBytes {
		writeSecurityError(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSecurityError(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE")
		return
	}
	canonical, bodyDigest, err := canonicalJSON(body)
	if err != nil {
		writeSecurityError(w, http.StatusBadRequest, "INVALID_JSON")
		return
	}
	idempotencyKey, ok := exactRequestHeader(r.Header, "Idempotency-Key")
	if !ok || !validIdempotencyKey(idempotencyKey) {
		writeSecurityError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED")
		return
	}
	ifMatch, ok := exactRequestHeader(r.Header, "If-Match")
	if !ok || !validIfMatch(ifMatch) {
		writeSecurityError(w, http.StatusPreconditionRequired, "IF_MATCH_REQUIRED")
		return
	}
	identity, capabilityNonce, err := s.authenticate(r, resource, bodyDigest, idempotencyKey, ifMatch)
	if err != nil {
		if errors.Is(err, errBrowserAntiForgery) {
			writeSecurityError(w, http.StatusForbidden, "ANTI_FORGERY_REFUSED")
		} else {
			writeSecurityError(w, http.StatusUnauthorized, "IDENTITY_REQUIRED")
		}
		return
	}
	request := MutationLedgerRequest{
		Identity: identity, Method: r.Method, Resource: resource, BodyDigest: bodyDigest,
		IdempotencyKey: idempotencyKey, IfMatch: ifMatch, CapabilityNonce: capabilityNonce, At: s.now().UTC(),
	}
	reservation, err := s.ledger.Reserve(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrCapabilitySpent):
			writeSecurityError(w, http.StatusUnauthorized, "CAPABILITY_SPENT")
		case errors.Is(err, ErrIdempotencyConflict), errors.Is(err, ErrIdempotencyInProgress):
			writeSecurityError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT")
		default:
			writeSecurityError(w, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE")
		}
		return
	}
	if reservation.Replay != nil {
		writeStoredResponse(w, *reservation.Replay)
		return
	}
	result, commandErr := command.Execute(r.Context(), AuthorizedMutation{
		Identity: identity, Method: r.Method, Resource: resource,
		CanonicalBody: canonical, BodyDigest: bodyDigest, IfMatch: ifMatch,
	})
	response := storedCommandResponse(result, commandErr)
	completionContext, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer cancel()
	if err := completeMutationAudit(completionContext, s.ledger, reservation.ID, response); err != nil {
		writeSecurityError(w, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE")
		return
	}
	writeStoredResponse(w, response)
}

// completeMutationAudit retries only the durable outcome record, never the
// command. Complete is idempotent for the same reservation/response, including
// an ambiguous SQLite commit, so a transient audit failure cannot create a
// duplicate optimization command.
func completeMutationAudit(ctx context.Context, ledger MutationLedger, id int64, response StoredMutationResponse) error {
	var err error
	for range 3 {
		if err = ledger.Complete(ctx, id, response); err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return err
}

func (s *MutationSecurity) authenticate(r *http.Request, resource, bodyDigest, idempotencyKey, ifMatch string) (MutationIdentity, string, error) {
	type candidate struct {
		identity MutationIdentity
		nonce    string
	}
	var candidates []candidate
	if s.browserSession != nil {
		if identity, ok := s.browserSession(r); ok {
			token, exact := exactRequestHeader(r.Header, "X-CSRF-Token")
			if !exact || !identity.valid() || identity.Mode != AuthModeBrowser ||
				!s.browserCSRF(r, token) || !s.boundary.OriginMatches(r) {
				return MutationIdentity{}, "", errBrowserAntiForgery
			}
			candidates = append(candidates, candidate{identity: identity})
		}
	}
	if identity, ok := s.mtlsIdentity(r); ok {
		candidates = append(candidates, candidate{identity: identity})
	}
	if values, present := requestHeaderValues(r.Header, "Authorization"); present {
		value, ok := oneRequestHeader(values)
		if !ok || s.capability == nil {
			return MutationIdentity{}, "", ErrCapabilityInvalid
		}
		parts := strings.SplitN(value, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "TossOS-Capability") || strings.TrimSpace(parts[1]) != parts[1] {
			return MutationIdentity{}, "", ErrCapabilityInvalid
		}
		claims, err := s.capability.Verify(parts[1], CapabilityBinding{
			Method: r.Method, Resource: resource, BodyDigest: bodyDigest, IdempotencyKey: idempotencyKey,
			IfMatch: ifMatch, Audience: s.boundary.PublicOrigin().String(),
		})
		if err != nil {
			return MutationIdentity{}, "", err
		}
		candidates = append(candidates, candidate{identity: MutationIdentity{
			Actor: claims.Actor, Client: claims.Client, Mode: AuthModeCapability,
		}, nonce: claims.Nonce})
	}
	if len(candidates) != 1 {
		return MutationIdentity{}, "", ErrCapabilityInvalid
	}
	return candidates[0].identity, candidates[0].nonce, nil
}

func (s *MutationSecurity) mtlsIdentity(r *http.Request) (MutationIdentity, bool) {
	if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.VerifiedChains[0]) == 0 {
		return MutationIdentity{}, false
	}
	digest := sha256.Sum256(r.TLS.VerifiedChains[0][0].Raw)
	identity, ok := s.mtls[hex.EncodeToString(digest[:])]
	return identity, ok
}

func storedCommandResponse(result MutationResult, err error) StoredMutationResponse {
	if err != nil {
		var precondition *PreconditionError
		if errors.As(err, &precondition) && validIfMatch(precondition.CurrentVersion) {
			return StoredMutationResponse{Status: http.StatusPreconditionFailed, Version: precondition.CurrentVersion,
				Body: errorResponseBody("PRECONDITION_FAILED", "The resource version is stale; reload the current version without automatic retry.", nil, time.Now())}
		}
		var commandError *CommandError
		if errors.As(err, &commandError) && validCommandError(commandError) {
			return StoredMutationResponse{Status: commandError.status,
				Body: errorResponseBody(commandError.code, commandError.message, nil, time.Now())}
		}
		return StoredMutationResponse{Status: http.StatusServiceUnavailable,
			Body: errorResponseBody("COMMAND_UNAVAILABLE", "The narrow mutation command did not complete safely.", nil, time.Now())}
	}
	invalidVersion := result.Version != "" && !validIfMatch(result.Version)
	invalidBody := result.Status == http.StatusNoContent && len(result.Body) != 0 ||
		result.Status != http.StatusNoContent && (len(result.Body) == 0 || !json.Valid(result.Body))
	if result.Status < 200 || result.Status > 299 || int64(len(result.Body)) > MaxRequestBodyBytes || invalidVersion || invalidBody {
		return StoredMutationResponse{Status: http.StatusServiceUnavailable,
			Body: errorResponseBody("INVALID_COMMAND_RESULT", "The command returned a response outside the stable API contract.", nil, time.Now())}
	}
	return StoredMutationResponse{Status: result.Status, Version: result.Version, Body: append([]byte(nil), result.Body...)}
}

func validCommandError(err *CommandError) bool {
	if err == nil || err.status != http.StatusBadRequest && err.status != http.StatusForbidden ||
		len(err.code) < 3 || len(err.code) > 64 || len(err.message) < 1 || len(err.message) > 256 {
		return false
	}
	for _, character := range err.code {
		if character != '_' && (character < 'A' || character > 'Z') {
			return false
		}
	}
	return !strings.ContainsAny(err.message, "\r\n")
}

func writeStoredResponse(w http.ResponseWriter, response StoredMutationResponse) {
	if response.Version != "" {
		w.Header().Set("ETag", response.Version)
	}
	if len(response.Body) > 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(response.Status)
	if response.Status != http.StatusNoContent {
		_, _ = w.Write(response.Body)
	}
}

func writeSecurityError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(errorResponseBody(code, "The mutation request was refused by the stable security contract.", nil, time.Now()))
}

func requestHeaderValues(header http.Header, name string) ([]string, bool) {
	values, ok := header[http.CanonicalHeaderKey(name)]
	return values, ok
}

func exactRequestHeader(header http.Header, name string) (string, bool) {
	values, present := requestHeaderValues(header, name)
	if !present {
		return "", false
	}
	return oneRequestHeader(values)
}

func oneRequestHeader(values []string) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	value := strings.TrimSpace(values[0])
	return value, value != "" && !strings.ContainsAny(value, "\r\n,")
}

type canonicalNumber string

func (n canonicalNumber) MarshalJSON() ([]byte, error) { return []byte(n), nil }

func canonicalJSON(body []byte) ([]byte, string, error) {
	if !utf8.Valid(body) {
		return nil, "", errors.New("httpapi: mutation JSON must be valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	value, err := decodeCanonicalValue(decoder)
	if err != nil {
		return nil, "", err
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, "", errors.New("httpapi: mutation JSON must be an object")
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, "", err
	}
	canonical, err := json.Marshal(value)
	if err != nil || int64(len(canonical)) > MaxRequestBodyBytes {
		return nil, "", errors.New("httpapi: JSON cannot be canonicalized")
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}

func digestOf(body string) string {
	_, digest, _ := canonicalJSON([]byte(body))
	return digest
}

func decodeCanonicalValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token := token.(type) {
	case json.Delim:
		switch token {
		case '{':
			object := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, errors.New("httpapi: object key is not a string")
				}
				if _, duplicate := object[key]; duplicate {
					return nil, errors.New("httpapi: duplicate JSON object key")
				}
				value, err := decodeCanonicalValue(decoder)
				if err != nil {
					return nil, err
				}
				object[key] = value
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return nil, errors.New("httpapi: unterminated JSON object")
			}
			return object, nil
		case '[':
			array := []any{}
			for decoder.More() {
				value, err := decodeCanonicalValue(decoder)
				if err != nil {
					return nil, err
				}
				array = append(array, value)
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return nil, errors.New("httpapi: unterminated JSON array")
			}
			return array, nil
		}
	case json.Number:
		normalized, err := normalizeJSONNumber(string(token))
		return canonicalNumber(normalized), err
	case string, bool, nil:
		return token, nil
	}
	return nil, errors.New("httpapi: unsupported JSON token")
}

func requireJSONEOF(decoder *json.Decoder) error {
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("httpapi: trailing JSON data")
	}
	return nil
}

func normalizeJSONNumber(raw string) (string, error) {
	lower := strings.ToLower(raw)
	exponent := 0
	if index := strings.IndexByte(lower, 'e'); index >= 0 {
		parsed, err := strconv.Atoi(lower[index+1:])
		if err != nil || parsed < -10000 || parsed > 10000 {
			return "", errors.New("httpapi: JSON number exponent is out of bounds")
		}
		exponent = parsed
		lower = lower[:index]
	}
	negative := strings.HasPrefix(lower, "-")
	if negative {
		lower = lower[1:]
	}
	fractionDigits := 0
	if index := strings.IndexByte(lower, '.'); index >= 0 {
		fractionDigits = len(lower) - index - 1
		lower = lower[:index] + lower[index+1:]
	}
	lower = strings.TrimLeft(lower, "0")
	if lower == "" {
		return "0", nil
	}
	shift := exponent - fractionDigits
	for strings.HasSuffix(lower, "0") {
		lower = strings.TrimSuffix(lower, "0")
		shift++
	}
	var normalized string
	if shift >= 0 {
		if len(lower)+shift > 20000 {
			return "", errors.New("httpapi: canonical JSON number is too large")
		}
		normalized = lower + strings.Repeat("0", shift)
	} else if point := len(lower) + shift; point > 0 {
		normalized = lower[:point] + "." + lower[point:]
	} else {
		if -point > 10000 {
			return "", errors.New("httpapi: canonical JSON number is too small")
		}
		normalized = "0." + strings.Repeat("0", -point) + lower
	}
	if negative {
		normalized = "-" + normalized
	}
	return normalized, nil
}

var _ = subtle.ConstantTimeCompare
