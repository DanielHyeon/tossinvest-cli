package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const openAPICredentialSavedAuditAction = "openapi.credentials.saved"
const openAPIOnboardingPendingName = "openapi-onboarding.pending"

type consoleOpenAPIState string

const (
	consoleOpenAPIReady               consoleOpenAPIState = "ready"
	consoleOpenAPIMissing             consoleOpenAPIState = "missing"
	consoleOpenAPIRejected            consoleOpenAPIState = "rejected"
	consoleOpenAPIEnvironmentRejected consoleOpenAPIState = "environment_rejected"
	consoleOpenAPIUnavailable         consoleOpenAPIState = "unavailable"
	consoleOpenAPISavedNotStarted     consoleOpenAPIState = "saved_not_started"
)

type consoleOpenAPIResult struct {
	State   consoleOpenAPIState
	Message string
}

type consoleOpenAPIDeps struct {
	getenv   func(string) string
	validate func(context.Context, official.Credentials, string) (probeResult, error)
	save     func(string, string, string) error
	remove   func(string) error
	temp     func(string) (string, func() error, error)
	audit    func() error
	pending  func(string) (bool, error)
	mark     func(string) error
	clear    func(string) error
}

func defaultConsoleOpenAPIDeps() consoleOpenAPIDeps {
	return consoleOpenAPIDeps{
		getenv: os.Getenv,
		validate: func(ctx context.Context, creds official.Credentials, tokenFile string) (probeResult, error) {
			return validateOpenAPICredentials(ctx, creds, tokenFile)
		},
		save: saveOpenAPICredentials,
		remove: func(path string) error {
			err := os.Remove(path)
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		},
		temp:    isolatedOpenAPITokenPath,
		pending: openAPIOnboardingPending,
		mark:    writeOpenAPIOnboardingPending,
		clear:   clearOpenAPIOnboardingPending,
		audit: func() error {
			log := openAuditLog()
			if log == nil {
				return errors.New("audit log unavailable")
			}
			return log.RecordAction(
				openAPICredentialSavedAuditAction,
				"openapi.credentials",
				"saved",
				"operator console",
			)
		},
	}
}

type consoleOpenAPISeam struct {
	credentialFile string
	tokenFile      string
	pendingFile    string
	deps           consoleOpenAPIDeps
}

func newConsoleOpenAPISeam(root *rootOptions, deps consoleOpenAPIDeps) (*consoleOpenAPISeam, error) {
	credentialFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, err
	}
	return &consoleOpenAPISeam{
		credentialFile: credentialFile,
		tokenFile:      tokenFile,
		pendingFile:    filepath.Join(filepath.Dir(credentialFile), openAPIOnboardingPendingName),
		deps:           deps,
	}, nil
}

func (s *consoleOpenAPISeam) Check(ctx context.Context) consoleOpenAPIResult {
	environment := s.environmentManaged()
	if !environment {
		pending, err := s.deps.pending(s.pendingFile)
		if err != nil {
			return openAPIUnavailable("Open API 자격증명 설정 상태를 읽지 못했다. soak은 시작하지 않았다.")
		}
		if pending {
			return savedNotStarted(
				"Open API 자격증명 설정이 완료되지 않아 soak은 시작되지 않음 — 설정 화면에서 다시 저장하라.")
		}
	}
	creds, err := official.LoadCredentials(s.deps.getenv, s.credentialFile)
	if err != nil {
		return openAPIUnavailable("Open API 자격증명 파일을 읽지 못했다. 파일 권한과 형식을 확인하라.")
	}
	if creds == nil {
		return consoleOpenAPIResult{
			State:   consoleOpenAPIMissing,
			Message: "저장된 Open API 자격증명이 없다.",
		}
	}
	validationToken := s.tokenFile
	var cleanup func() error
	if environment {
		validationToken, cleanup, err = s.deps.temp(s.tokenFile)
		if err != nil {
			return openAPIUnavailable("환경변수 Open API 자격증명 검증용 임시 저장소를 준비하지 못했다. soak은 시작하지 않았다.")
		}
	}
	result, err := s.deps.validate(ctx, *creds, validationToken)
	if cleanup != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			return openAPIUnavailable("환경변수 Open API 자격증명 검증용 임시 토큰을 제거하지 못했다. soak은 시작하지 않았다.")
		}
	}
	if err != nil {
		return openAPIUnavailable("Open API 자격증명 확인을 완료하지 못했다. 잠시 후 다시 시도하라.")
	}
	classified := classifyConsoleOpenAPI(result, environment)
	if environment && classified.State == consoleOpenAPIReady {
		if err := s.deps.remove(s.tokenFile); err != nil {
			return openAPIUnavailable("환경변수 Open API 자격증명은 확인했지만 기존 토큰 캐시를 비우지 못했다. soak은 시작하지 않았다.")
		}
	}
	return classified
}

func (s *consoleOpenAPISeam) Save(
	ctx context.Context,
	key string,
	secret string,
) consoleOpenAPIResult {
	if s.environmentManaged() {
		return environmentCredentialRejected()
	}
	key = strings.TrimSpace(key)
	secret = strings.TrimSpace(secret)
	if key == "" || secret == "" {
		return consoleOpenAPIResult{
			State:   consoleOpenAPIRejected,
			Message: "Open API 키와 시크릿을 모두 입력하라.",
		}
	}

	isolatedToken, cleanup, err := s.deps.temp(s.tokenFile)
	if err != nil {
		return openAPIUnavailable("자격증명 검증용 임시 저장소를 준비하지 못했다. 아무것도 저장하지 않았다.")
	}
	result, validateErr := s.deps.validate(ctx, official.Credentials{
		APIKey:    key,
		SecretKey: secret,
	}, isolatedToken)
	cleanupErr := cleanup()
	if cleanupErr != nil {
		return openAPIUnavailable("자격증명 검증용 임시 토큰을 제거하지 못했다. 아무것도 저장하거나 시작하지 않았다.")
	}
	if validateErr != nil {
		return openAPIUnavailable("Open API 자격증명 확인을 완료하지 못했다. 아무것도 저장하지 않았다.")
	}
	classified := classifyConsoleOpenAPI(result, false)
	if classified.State != consoleOpenAPIReady {
		return classified
	}
	if err := s.deps.mark(s.pendingFile); err != nil {
		return openAPIUnavailable("Open API 자격증명 설정 상태를 보호 저장하지 못했다. 아무것도 저장하거나 시작하지 않았다.")
	}
	if err := s.deps.save(s.credentialFile, key, secret); err != nil {
		return savedNotStarted(
			"Open API 자격증명 보호 저장을 완료하지 못해 soak은 시작되지 않음 — 설정 화면에서 다시 저장하라.")
	}
	if err := s.deps.remove(s.tokenFile); err != nil {
		return savedNotStarted("자격증명은 저장됨, soak은 시작되지 않음 — 기존 토큰 캐시를 비우지 못했다. Open API 설정 화면에서 다시 저장하라.")
	}
	if err := s.deps.audit(); err != nil {
		return savedNotStarted("자격증명은 저장됨, soak은 시작되지 않음 — 감사 기록을 남기지 못했다. 문제를 확인한 뒤 Open API 설정 화면에서 다시 저장하라.")
	}
	if err := s.deps.clear(s.pendingFile); err != nil {
		return savedNotStarted("자격증명은 저장됨, soak은 시작되지 않음 — 설정 완료 상태를 기록하지 못했다. 설정 화면에서 다시 저장하라.")
	}
	return consoleOpenAPIResult{State: consoleOpenAPIReady}
}

func (s *consoleOpenAPISeam) environmentManaged() bool {
	return s.deps.getenv("TOSSCTL_OPENAPI_KEY") != "" &&
		s.deps.getenv("TOSSCTL_OPENAPI_SECRET") != ""
}

func (s *consoleOpenAPISeam) PrepareSpawn() error {
	return s.deps.remove(s.tokenFile)
}

func openAPIOnboardingPending(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, err
	}
}

func writeOpenAPIOnboardingPending(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if err := f.Chmod(0o600); err != nil {
		return err
	}
	if _, err := f.WriteString("pending\n"); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func clearOpenAPIOnboardingPending(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func isolatedOpenAPITokenPath(normalTokenFile string) (string, func() error, error) {
	dir := filepath.Dir(normalTokenFile)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", nil, err
	}
	f, err := os.CreateTemp(dir, ".openapi-login-token-*")
	if err != nil {
		return "", nil, err
	}
	name := f.Name()
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", nil, err
	}
	if err := os.Remove(name); err != nil {
		return "", nil, err
	}
	cleanup := func() error {
		err := os.Remove(name)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return name, cleanup, nil
}

func classifyConsoleOpenAPI(result probeResult, environment bool) consoleOpenAPIResult {
	if result.OK {
		return consoleOpenAPIResult{State: consoleOpenAPIReady}
	}
	switch result.ErrorKind {
	case "auth":
		if environment {
			return environmentCredentialRejected()
		}
		return consoleOpenAPIResult{
			State:   consoleOpenAPIRejected,
			Message: "Open API 키 또는 시크릿이 거부되었다. 새 자격증명을 입력하라.",
		}
	case "ip_not_allowed":
		return openAPIUnavailable("현재 접속 IP가 Open API 허용 목록에 없다. 토스증권 개발자 포털의 허용 IP를 확인하라.")
	case "rate_limited":
		return openAPIUnavailable("Open API 호출 한도에 도달했다. 잠시 후 soak 재시작을 다시 누르라.")
	case "server_error":
		return openAPIUnavailable("Open API 서버 오류로 자격증명을 확인하지 못했다. 잠시 후 다시 시도하라.")
	case "transport_error":
		return openAPIUnavailable("Open API 네트워크 연결을 확인하지 못했다. VPN과 인터넷 연결을 확인하라.")
	default:
		return openAPIUnavailable("Open API 자격증명 확인이 실패했다. soak은 시작하지 않았다.")
	}
}

func environmentCredentialRejected() consoleOpenAPIResult {
	return consoleOpenAPIResult{
		State: consoleOpenAPIEnvironmentRejected,
		Message: "환경변수 자격증명이 거부되었다. TOSSCTL_OPENAPI_KEY와 " +
			"TOSSCTL_OPENAPI_SECRET을 갱신한 뒤 컨테이너를 다시 생성하라. 값은 이 화면에 표시하지 않는다.",
	}
}

func openAPIUnavailable(message string) consoleOpenAPIResult {
	return consoleOpenAPIResult{
		State:   consoleOpenAPIUnavailable,
		Message: message,
	}
}

func savedNotStarted(message string) consoleOpenAPIResult {
	return consoleOpenAPIResult{
		State:   consoleOpenAPISavedNotStarted,
		Message: message,
	}
}
