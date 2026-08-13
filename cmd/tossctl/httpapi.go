package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performancejournal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/spf13/cobra"
)

const (
	defaultHTTPAPIPort            = 37086
	defaultHTTPAPIBind            = "127.0.0.1"
	defaultHTTPAPIPublicURL       = "https://127.0.0.1:37086"
	defaultHTTPAPIPublishInterval = 30 * time.Second
)

type httpAPIOptions struct {
	port                int
	bind                string
	allowedCIDRs        []string
	publicURL           string
	tlsCert             string
	tlsKey              string
	trustedProxy        string
	tlsForwarded        bool
	securityDB          string
	capabilityPublicKey string
}

func newHTTPAPICmd(root *rootOptions) *cobra.Command {
	opts := &httpAPIOptions{}
	cmd := &cobra.Command{
		Use:   "httpapi",
		Short: "Serve the private read-only operator REST and SSE API",
		Long: strings.TrimSpace(`
Serve the versioned private operator API on loopback or an explicitly allowed
VPN address. Reads need no application login after the private peer and TLS
boundary is validated. The service has no order, LIVE, gate, kill-switch, engine
start/stop, or activation-manifest route and never starts the engine.

All controls are server-owned. Read routes accept no query or request body, and
the daemon provides no login form, symbol field, numeric field, reason field, or
typed confirmation UI.`),
		Annotations:  map[string]string{"source": "both"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHTTPAPI(cmd, root, opts)
		},
	}
	cmd.Flags().IntVar(&opts.port, "port", defaultHTTPAPIPort, "Private HTTPS port")
	cmd.Flags().StringVar(&opts.bind, "bind", defaultHTTPAPIBind, "Loopback or private/VPN bind IP literal")
	cmd.Flags().StringSliceVar(&opts.allowedCIDRs, "allowed-cidr", []string{"127.0.0.0/8"},
		"Private client CIDR accepted by the daemon; repeat for multiple networks")
	cmd.Flags().StringVar(&opts.publicURL, "public-url", defaultHTTPAPIPublicURL,
		"Canonical private HTTPS origin, including effective port")
	cmd.Flags().StringVar(&opts.tlsCert, "tls-cert", "", "PEM TLS certificate for direct TLS termination")
	cmd.Flags().StringVar(&opts.tlsKey, "tls-key", "", "PEM TLS private key for direct TLS termination")
	cmd.Flags().StringVar(&opts.trustedProxy, "trusted-proxy", "",
		"Exact private proxy IP allowed to supply one-hop forwarded identity")
	cmd.Flags().BoolVar(&opts.tlsForwarded, "tls-forwarded", false,
		"Accept TLS termination only from --trusted-proxy")
	cmd.Flags().StringVar(&opts.securityDB, "security-db", "",
		"Separate 0600 nonce, idempotency and mutation-audit database")
	cmd.Flags().StringVar(&opts.capabilityPublicKey, "capability-public-key", "",
		"Ed25519 public-key file enabling signed one-time optimization mutations")
	return cmd
}

type httpAPIResources struct {
	reader       *httpAPIReader
	journal      *journal.ReadOnly
	performance  consolePerformanceCapabilities
	optimization *optimization.Store
	security     *httpapi.SecurityStore
}

func (r *httpAPIResources) Close() error {
	if r == nil {
		return nil
	}
	var errs []error
	if r.optimization != nil {
		errs = append(errs, r.optimization.Close())
	}
	if r.security != nil {
		errs = append(errs, r.security.Close())
	}
	errs = append(errs, r.performance.Close())
	if r.journal != nil {
		errs = append(errs, r.journal.Close())
	}
	return errors.Join(errs...)
}

func runHTTPAPI(cmd *cobra.Command, root *rootOptions, opts *httpAPIOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, stop := signal.NotifyContext(ctx, consoleTerminationSignals()...)
	defer stop()

	boundary, directTLS, err := httpAPIBoundary(opts)
	if err != nil {
		return err
	}
	var serverCertificate tls.Certificate
	if directTLS {
		serverCertificate, err = networkboundary.LoadServerCertificate(
			opts.tlsCert, opts.tlsKey, boundary.PublicOrigin().Host, time.Now(),
		)
		if err != nil {
			return err
		}
	}
	resources, err := openHTTPAPIResources(ctx, root, time.Now)
	if err != nil {
		return err
	}
	defer resources.Close()
	// a108 D4 — descriptor 부재와 dial 실패는 같은 강등이다.
	//
	// 이 데몬은 descriptor 가 없으면 전략 화면 없이 뜨도록 **설계돼 있었다.** 그런데
	// descriptor 가 있는데 dial 이 실패하면 fatal 이었고, 2026-08-13 의 반쪽 잔재
	// (endpoint.json 은 남고 runtime.sock 은 unlink 된 상태) 가 정확히 그쪽으로
	// 떨어져 컨테이너가 `Restarting (1)` 로 돌았다. 설계된 강등과 crash loop 사이의
	// 차이가 「파일 하나가 남아 있는가」일 이유가 없다.
	//
	// fatal 로 남는 것은 descriptor 를 **조사할 수조차 없는 경우**뿐이다. 그것은
	// 잔재가 아니라 환경 이상이고, 삼키면 잘못 배치된 데몬이 조용히 반쪽으로 뜬다.
	var strategyRuntime httpapi.StrategyRuntimeReader
	if dir, dirErr := engineJournalDir(root); dirErr == nil {
		descriptor := strategyprojectionrpc.DescriptorPath(dir)
		if _, statErr := os.Stat(descriptor); statErr == nil {
			client, dialErr := strategyprojectionrpc.Dial(ctx, descriptor)
			if dialErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"note: strategy runtime projection에 연결하지 못했다 (%s): %v\n"+
						"      데몬은 전략 화면 없이 뜬다 — 엔진이 다시 뜬 뒤 이 데몬을 재시작하면 붙는다.\n",
					descriptor, dialErr)
			} else {
				strategyRuntime = client
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("httpapi: inspect strategy runtime projection: %w", statErr)
		}
	}
	resources.reader.strategyRuntime = strategyRuntime

	snapshots, err := newHTTPAPISnapshotCache(resources.reader)
	if err != nil {
		return err
	}
	stream, err := httpapi.NewStream(httpapi.StreamOptions{}, snapshots.Snapshot)
	if err != nil {
		return err
	}
	defer stream.Close()
	mutationRoutes, err := httpAPIMutationRoutes(boundary, resources, opts, time.Now)
	if err != nil {
		return err
	}
	router, err := httpapi.NewRouter(httpapi.Options{
		Reader: resources.reader, StrategyRuntime: strategyRuntime, Stream: stream, MutationRoutes: mutationRoutes, Now: time.Now,
	})
	if err != nil {
		return err
	}
	address := net.JoinHostPort(boundary.Bind().String(), fmt.Sprintf("%d", opts.port))
	server, err := httpapi.NewServer(address, boundary.HandlerWithRejection(router, httpapi.WriteBoundaryRejection))
	if err != nil {
		return err
	}
	if directTLS {
		server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{serverCertificate}}
	}

	publisherCtx, cancelPublisher := context.WithCancel(ctx)
	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		publishHTTPAPISnapshots(publisherCtx, stream, snapshots, defaultHTTPAPIPublishInterval)
	}()
	defer func() {
		cancelPublisher()
		<-publisherDone
	}()

	serveErr := make(chan error, 1)
	go func() {
		if directTLS {
			serveErr <- server.ListenAndServeTLS("", "")
			return
		}
		serveErr <- server.ListenAndServe()
	}()
	fmt.Fprintf(cmd.OutOrStdout(), "httpapi private endpoint %s/api/v1\n", strings.TrimSuffix(opts.publicURL, "/"))

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-serveErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func httpAPIMutationRoutes(
	boundary *networkboundary.Boundary,
	resources *httpAPIResources,
	opts *httpAPIOptions,
	now func() time.Time,
) (map[string]http.Handler, error) {
	securityPath := strings.TrimSpace(opts.securityDB)
	publicKeyPath := strings.TrimSpace(opts.capabilityPublicKey)
	if securityPath == "" && publicKeyPath == "" {
		return nil, nil
	}
	if securityPath == "" || publicKeyPath == "" {
		return nil, errors.New("httpapi: --security-db and --capability-public-key must be configured together")
	}
	publicKey, err := loadHTTPAPICapabilityPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}
	verifier, err := httpapi.NewCapabilityVerifier(publicKey, now)
	if err != nil {
		return nil, err
	}
	securityStore, err := httpapi.OpenSecurityStore(securityPath)
	if err != nil {
		return nil, err
	}
	resources.security = securityStore
	security, err := httpapi.NewMutationSecurity(httpapi.MutationSecurityOptions{
		Boundary: boundary, Ledger: securityStore, Capability: verifier, Now: now,
	})
	if err != nil {
		return nil, err
	}
	return map[string]http.Handler{
		"/api/v1/optimization/previews": security.Handler(
			"/api/v1/optimization/previews",
			httpAPIOptimizationMutation{store: resources.optimization, kind: httpAPIOptimizationPreview},
		),
		"/api/v1/optimization/applications": security.Handler(
			"/api/v1/optimization/applications",
			httpAPIOptimizationMutation{store: resources.optimization, kind: httpAPIOptimizationApply},
		),
	}, nil
}

func loadHTTPAPICapabilityPublicKey(path string) (ed25519.PublicKey, error) {
	parentInfo, err := os.Lstat(filepath.Dir(filepath.Clean(path)))
	if err != nil || !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 || parentInfo.Mode().Perm()&0o002 != 0 {
		return nil, errors.New("httpapi: capability public-key directory must be a non-world-writable real directory")
	}
	file, err := openHTTPAPIRegularNoFollow(path)
	if err != nil {
		return nil, fmt.Errorf("httpapi: securely opening capability public key: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("httpapi: reading capability public-key metadata: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o022 != 0 || info.Size() > 4096 {
		return nil, errors.New("httpapi: capability public key must be a small regular file, not a symlink")
	}
	body, err := io.ReadAll(io.LimitReader(file, 4097))
	if err != nil {
		return nil, fmt.Errorf("httpapi: reading capability public key: %w", err)
	}
	if len(body) > 4096 {
		return nil, errors.New("httpapi: capability public key exceeds 4096 bytes")
	}
	encoded := strings.TrimSpace(string(body))
	var decoded []byte
	for _, decode := range []func(string) ([]byte, error){
		hex.DecodeString,
		base64.RawStdEncoding.DecodeString,
		base64.StdEncoding.DecodeString,
		base64.RawURLEncoding.DecodeString,
	} {
		candidate, decodeErr := decode(encoded)
		if decodeErr == nil && len(candidate) == ed25519.PublicKeySize {
			decoded = candidate
			break
		}
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, errors.New("httpapi: capability public key must encode exactly one Ed25519 public key")
	}
	return ed25519.PublicKey(append([]byte(nil), decoded...)), nil
}

type httpAPIOptimizationMutationKind uint8

const (
	httpAPIOptimizationPreview httpAPIOptimizationMutationKind = iota + 1
	httpAPIOptimizationApply
)

type httpAPIOptimizationMutation struct {
	store *optimization.Store
	kind  httpAPIOptimizationMutationKind
}

// Remote applications are a deny-by-default transition registry. a051 ships
// no remotely applicable effective transition: the current exit-protection
// owner is contextual. A future neutral preset requires an explicit code review
// here in addition to owner metadata; metadata drift alone cannot widen remote
// authority.
var remoteApplicableOptimizationTransitions = map[string]map[string]struct{}{}

func (m httpAPIOptimizationMutation) Execute(
	ctx context.Context,
	request httpapi.AuthorizedMutation,
) (httpapi.MutationResult, error) {
	if m.store == nil {
		return httpapi.MutationResult{}, errors.New("httpapi: optimization command is unavailable")
	}
	commander, err := m.store.ForActor(httpAPIOptimizationActor(request.Identity))
	if err != nil {
		return httpapi.MutationResult{}, err
	}
	view, err := commander.Read(ctx)
	if err != nil {
		return httpapi.MutationResult{}, err
	}
	currentETag := quoteHTTPAPIVersion(view.Snapshot.Version)
	if request.IfMatch != currentETag {
		return httpapi.MutationResult{}, &httpapi.PreconditionError{CurrentVersion: currentETag}
	}
	switch m.kind {
	case httpAPIOptimizationPreview:
		return executeHTTPAPIOptimizationPreview(ctx, commander, request.CanonicalBody, view.Snapshot.Version)
	case httpAPIOptimizationApply:
		return executeHTTPAPIOptimizationApply(ctx, commander, request.CanonicalBody, currentETag)
	default:
		return httpapi.MutationResult{}, errors.New("httpapi: unknown optimization command")
	}
}

func httpAPIOptimizationActor(identity httpapi.MutationIdentity) string {
	return fmt.Sprintf("httpapi:%d:%s:%d:%s", len(identity.Actor), identity.Actor, len(identity.Client), identity.Client)
}

type httpAPIOptimizationPreviewBody struct {
	Category    string `json:"category"`
	BaseVersion uint64 `json:"baseVersion"`
	Changes     []struct {
		Key      string `json:"key"`
		OptionID string `json:"optionId"`
	} `json:"changes"`
}

func executeHTTPAPIOptimizationPreview(
	ctx context.Context,
	commander optimization.Commander,
	body []byte,
	currentVersion uint64,
) (httpapi.MutationResult, error) {
	var input httpAPIOptimizationPreviewBody
	if err := decodeHTTPAPIMutationBody(body, &input); err != nil || input.BaseVersion != currentVersion ||
		input.Category != string(optimization.CategoryExitProtection) || len(input.Changes) < 1 || len(input.Changes) > 16 {
		return httpapi.MutationResult{}, invalidHTTPAPISelectionError()
	}
	changes := make(map[string]string, len(input.Changes))
	for _, change := range input.Changes {
		key, optionID := strings.TrimSpace(change.Key), strings.TrimSpace(change.OptionID)
		if key == "" || optionID == "" {
			return httpapi.MutationResult{}, invalidHTTPAPISelectionError()
		}
		if _, duplicate := changes[key]; duplicate {
			return httpapi.MutationResult{}, invalidHTTPAPISelectionError()
		}
		changes[key] = optionID
	}
	preview, err := commander.Preview(ctx, optimization.PreviewRequest{
		BaseVersion: input.BaseVersion, Category: optimization.CategoryExitProtection, Changes: changes,
		Source: optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		if errors.Is(err, optimization.ErrVersionConflict) {
			latest, readErr := commander.Read(ctx)
			if readErr == nil {
				return httpapi.MutationResult{}, &httpapi.PreconditionError{CurrentVersion: quoteHTTPAPIVersion(latest.Snapshot.Version)}
			}
		}
		if errors.Is(err, optimization.ErrInvalidCandidate) || errors.Is(err, optimization.ErrCapabilityInvalid) {
			return httpapi.MutationResult{}, invalidHTTPAPISelectionError()
		}
		return httpapi.MutationResult{}, err
	}
	response, err := json.Marshal(struct {
		SchemaVersion string               `json:"schemaVersion"`
		Resource      string               `json:"resource"`
		Data          optimization.Preview `json:"data"`
	}{httpapi.SchemaVersion, "optimization-preview", preview})
	return httpapi.MutationResult{Status: http.StatusCreated, Version: quoteHTTPAPIVersion(currentVersion), Body: response}, err
}

type httpAPIOptimizationApplyBody struct {
	Capability string `json:"capability"`
	Confirmed  bool   `json:"confirmed"`
}

func executeHTTPAPIOptimizationApply(
	ctx context.Context,
	commander optimization.Commander,
	body []byte,
	currentETag string,
) (httpapi.MutationResult, error) {
	var input httpAPIOptimizationApplyBody
	if err := decodeHTTPAPIMutationBody(body, &input); err != nil || strings.TrimSpace(input.Capability) == "" {
		return httpapi.MutationResult{}, invalidHTTPAPIApplicationError()
	}
	conflict, err := commander.RecoverConflict(ctx, input.Capability)
	if err != nil {
		if errors.Is(err, optimization.ErrCapabilityInvalid) {
			return httpapi.MutationResult{}, invalidHTTPAPIApplicationError()
		}
		return httpapi.MutationResult{}, err
	}
	// Current a050 exposes only contextual exit protection. A native request may
	// create and inspect its preview, but remote apply is impossible until an
	// owner descriptor explicitly proves a neutral direction. This is the
	// compile-time-independent barrier that keeps a generic applications route
	// from becoming a protection-weakening endpoint.
	for _, change := range conflict.Attempted {
		options, fieldAllowed := remoteApplicableOptimizationTransitions[change.Key]
		_, transitionAllowed := options[change.AfterOptionID]
		if change.Safety != settingmeta.SafetyNeutral || !fieldAllowed || !transitionAllowed {
			return httpapi.MutationResult{}, httpapi.NewCommandError(http.StatusForbidden, "REMOTE_TRANSITION_REFUSED",
				"The requested transition is reserved for the local human approval channel.")
		}
	}
	result, err := commander.Apply(ctx, optimization.ApplyRequest{Capability: input.Capability, Confirmed: input.Confirmed})
	if err != nil {
		if errors.Is(err, optimization.ErrVersionConflict) {
			latest, readErr := commander.Read(ctx)
			if readErr == nil {
				return httpapi.MutationResult{}, &httpapi.PreconditionError{CurrentVersion: quoteHTTPAPIVersion(latest.Snapshot.Version)}
			}
		}
		if errors.Is(err, optimization.ErrCapabilityInvalid) || errors.Is(err, optimization.ErrCapabilityTooEarly) ||
			errors.Is(err, optimization.ErrCapabilityExpired) || errors.Is(err, optimization.ErrConfirmationRequired) {
			return httpapi.MutationResult{}, invalidHTTPAPIApplicationError()
		}
		return httpapi.MutationResult{}, err
	}
	response, err := json.Marshal(struct {
		SchemaVersion string                   `json:"schemaVersion"`
		Resource      string                   `json:"resource"`
		Data          optimization.ApplyResult `json:"data"`
	}{httpapi.SchemaVersion, "optimization-application", result})
	return httpapi.MutationResult{Status: http.StatusCreated,
		Version: quoteHTTPAPIVersion(result.Snapshot.Version), Body: response}, err
}

func invalidHTTPAPISelectionError() error {
	return httpapi.NewCommandError(http.StatusBadRequest, "INVALID_SELECTION",
		"Choose one of the current server-owned optimization presets and reload stale choices.")
}

func invalidHTTPAPIApplicationError() error {
	return httpapi.NewCommandError(http.StatusBadRequest, "INVALID_APPLICATION",
		"The preview application is invalid, incomplete, expired, or not yet available.")
}

func decodeHTTPAPIMutationBody(body []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("httpapi: mutation body has trailing content")
	}
	return nil
}

func quoteHTTPAPIVersion(version uint64) string {
	return `"` + strconv.FormatUint(version, 10) + `"`
}

func httpAPIBoundary(opts *httpAPIOptions) (*networkboundary.Boundary, bool, error) {
	if opts == nil || opts.port < 1 || opts.port > 65535 {
		return nil, false, errors.New("httpapi: port must be between 1 and 65535")
	}
	cert, key := strings.TrimSpace(opts.tlsCert), strings.TrimSpace(opts.tlsKey)
	if (cert == "") != (key == "") {
		return nil, false, errors.New("httpapi: --tls-cert and --tls-key must be configured together")
	}
	directTLS := cert != ""
	boundary, err := networkboundary.New(networkboundary.ServerConfig{
		Bind: opts.bind, AllowedCIDRs: opts.allowedCIDRs, PublicOrigin: opts.publicURL,
		TrustedProxy: opts.trustedProxy, TLSConfigured: directTLS, TLSForwarded: opts.tlsForwarded,
	})
	if err != nil {
		return nil, false, err
	}
	return boundary, directTLS, nil
}

func openHTTPAPIResources(ctx context.Context, root *rootOptions, now func() time.Time) (*httpAPIResources, error) {
	journalPath, err := consoleJournalPath(root)
	if err != nil {
		return nil, err
	}
	resources := &httpAPIResources{}
	journalReader, journalErr := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: journalPath})
	if journalErr == nil {
		resources.journal = journalReader
	}

	performanceCaps, performanceErr := openConsolePerformanceCapabilities(filepath.Dir(journalPath), now)
	if performanceErr == nil {
		resources.performance = performanceCaps
	}
	registry, err := optimization.CoreRegistry(ctx)
	if err != nil {
		resources.Close()
		return nil, err
	}
	store, err := optimization.Open(ctx, optimization.Options{
		Path:     filepath.Join(filepath.Dir(journalPath), optimization.DatabaseFileName),
		Registry: registry, Actor: "httpapi-daemon", Evidence: resources.performance.Evidence,
	})
	if err != nil {
		resources.Close()
		return nil, err
	}
	resources.optimization = store

	shared := newConsoleBroker(root)
	adoptionSettings := newAdoptionSettingsSeam(root)
	reader := &httpAPIReader{
		now: now, holdings: newConsoleHoldings(shared), orders: consoleOrdersSeam(shared),
		signals: consoleSignalsSeam(root), journal: journalReader, journalErr: journalErr,
		performance: resources.performance.Performance, performanceErr: performanceErr,
		optimization: store,
		accountRef: func() (string, error) {
			_, accountRef, err := shared.resolve()
			return accountRef, err
		},
	}
	if adoptionSettings != nil {
		reader.adoptionDesired = adoptionSettings.Load
	}
	if dir, err := engineJournalDir(root); err == nil {
		reader.engineMarker = enginelock.MarkerPath(dir)
		reader.managementRuntime = positionPolicyRuntimeDescriptorReader{
			descriptorPath: positionpolicyrpc.RuntimeDescriptorPath(dir),
		}
	}
	if journalReader != nil {
		reader.performanceSource = performancejournal.New(journalReader)
	}
	if err := reader.validate(); err != nil {
		resources.Close()
		return nil, err
	}
	resources.reader = reader
	return resources, nil
}

func publishHTTPAPISnapshots(ctx context.Context, stream *httpapi.Stream, snapshots *httpAPISnapshotCache, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var previous [sha256.Size]byte
	var known bool
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			body, err := snapshots.Refresh(ctx)
			if err != nil {
				continue
			}
			digest := sha256.Sum256(body)
			if known && digest == previous {
				continue
			}
			if _, err := stream.Publish(httpapi.StreamEventSnapshot, body); err != nil {
				return
			}
			previous, known = digest, true
		}
	}
}
