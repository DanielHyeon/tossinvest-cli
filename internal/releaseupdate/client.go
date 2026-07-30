// Package releaseupdate downloads only the fixed official tossctl release asset,
// verifies its Sigstore provenance, and returns executable bytes for a separate
// fixed-path publisher. It never installs or executes those bytes.
package releaseupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/updatecheck"
	"github.com/JungHoonGhae/tossinvest-cli/internal/version"
)

const (
	productionAPIBase = "https://api.github.com"
	apiVersion        = "2026-03-10"
	userAgent         = "tossctl-signed-release"
)

var stableTag = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type requestKind uint8

const (
	requestMetadata requestKind = iota
	requestAsset
	requestAttestationList
	requestBundle
	requestTUF
)

type responseLimits struct {
	metadata        int64
	archive         int64
	attestationList int64
	bundle          int64
	expanded        int64
	binary          int64
	entries         int
	pages           int
	bundles         int
}

func defaultLimits() responseLimits {
	return responseLimits{
		metadata: 2 << 20, archive: 256 << 20, attestationList: 2 << 20,
		bundle: 10 << 20, expanded: 512 << 20, binary: 200 << 20,
		entries: 512, pages: 3, bundles: 100,
	}
}

type Provenance struct {
	WorkflowIdentity string
	SourceCommit     string
}

type Release struct {
	Tag              string
	AssetName        string
	ArchiveSHA256    string
	WorkflowIdentity string
	SourceCommit     string
	Binary           []byte
	Bootstrap        bool
}

type provenanceVerifier interface {
	Verify(
		ctx context.Context,
		bundle []byte,
		digest, tag, asset string,
	) (Provenance, error)
}

type clientConfig struct {
	apiBase        string
	httpClient     *http.Client
	goos           string
	goarch         string
	currentVersion string
	verifier       provenanceVerifier
	allowURL       func(*url.URL, requestKind) bool
	limits         responseLimits
}

type Client struct {
	apiBase        string
	httpClient     *http.Client
	goos           string
	goarch         string
	currentVersion string
	verifier       provenanceVerifier
	allowURL       func(*url.URL, requestKind) bool
	limits         responseLimits
}

func NewProduction(cacheDir, currentVersion string) (*Client, error) {
	verifier, err := newProductionVerifier(cacheDir)
	if err != nil {
		return nil, err
	}
	return newClient(clientConfig{
		apiBase:        productionAPIBase,
		httpClient:     &http.Client{Timeout: 45 * time.Second},
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		currentVersion: currentVersion,
		verifier:       verifier,
		allowURL:       productionURLAllowed,
		limits:         defaultLimits(),
	})
}

func newClient(cfg clientConfig) (*Client, error) {
	if strings.TrimSpace(cfg.apiBase) == "" || cfg.httpClient == nil ||
		cfg.verifier == nil || cfg.allowURL == nil {
		return nil, errors.New("releaseupdate: incomplete client configuration")
	}
	if cfg.limits.metadata == 0 {
		cfg.limits = defaultLimits()
	}
	base, err := url.Parse(cfg.apiBase)
	if err != nil || !cfg.allowURL(base, requestMetadata) {
		return nil, fmt.Errorf("releaseupdate: invalid fixed API base %q", cfg.apiBase)
	}
	return &Client{
		apiBase: strings.TrimRight(cfg.apiBase, "/"), httpClient: cfg.httpClient,
		goos: cfg.goos, goarch: cfg.goarch, currentVersion: cfg.currentVersion,
		verifier: cfg.verifier, allowURL: cfg.allowURL, limits: cfg.limits,
	}, nil
}

func (c *Client) Fetch(ctx context.Context) (Release, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if c.goos != "linux" && c.goos != "darwin" {
		return Release{}, fmt.Errorf("releaseupdate: signed staging is unsupported on %s/%s", c.goos, c.goarch)
	}
	if c.goarch != "amd64" && c.goarch != "arm64" {
		return Release{}, fmt.Errorf("releaseupdate: no official asset for %s/%s", c.goos, c.goarch)
	}

	releaseURL := fmt.Sprintf("%s/repos/%s/releases/latest", c.apiBase, version.Repo)
	body, _, err := c.get(ctx, releaseURL, requestMetadata, c.limits.metadata)
	if err != nil {
		return Release{}, fmt.Errorf("releaseupdate: discovering latest release: %w", err)
	}
	var latest struct {
		TagName    string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name               string `json:"name"`
			Size               int64  `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &latest); err != nil {
		return Release{}, fmt.Errorf("releaseupdate: decoding latest release: %w", err)
	}
	if latest.Draft || latest.Prerelease || !stableTag.MatchString(latest.TagName) {
		return Release{}, fmt.Errorf("releaseupdate: latest release %q is not a canonical stable tag", latest.TagName)
	}
	bootstrap := !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).
		MatchString(c.currentVersion)
	if !bootstrap && !updatecheck.IsNewer(strings.TrimPrefix(latest.TagName, "v"), c.currentVersion) {
		return Release{}, fmt.Errorf(
			"releaseupdate: refusing equal or older release %s while running %s",
			latest.TagName, c.currentVersion)
	}

	assetName := fmt.Sprintf("tossctl-%s-%s.tar.gz", c.goos, c.goarch)
	assetURL := ""
	assetSize := int64(-1)
	for _, asset := range latest.Assets {
		if asset.Name != assetName {
			continue
		}
		if assetURL != "" {
			return Release{}, fmt.Errorf("releaseupdate: duplicate release asset %q", assetName)
		}
		assetURL, assetSize = asset.BrowserDownloadURL, asset.Size
	}
	if assetURL == "" {
		return Release{}, fmt.Errorf("releaseupdate: release %s has no exact asset %s", latest.TagName, assetName)
	}
	if assetSize < 0 || assetSize > c.limits.archive {
		return Release{}, fmt.Errorf("releaseupdate: asset size %d exceeds policy", assetSize)
	}
	archive, _, err := c.get(ctx, assetURL, requestAsset, c.limits.archive)
	if err != nil {
		return Release{}, fmt.Errorf("releaseupdate: downloading %s: %w", assetName, err)
	}
	if assetSize != 0 && int64(len(archive)) != assetSize {
		return Release{}, fmt.Errorf(
			"releaseupdate: asset size changed: release says %d, response has %d",
			assetSize, len(archive))
	}
	sum := sha256.Sum256(archive)
	digest := hex.EncodeToString(sum[:])

	provenance, err := c.verifyAttestations(ctx, digest, latest.TagName, assetName)
	if err != nil {
		return Release{}, err
	}
	binary, err := extractTossctl(archive, archiveLimits{
		maxArchive: c.limits.archive, maxBinary: c.limits.binary,
		maxExpanded: c.limits.expanded, maxEntries: c.limits.entries,
	})
	if err != nil {
		return Release{}, fmt.Errorf("releaseupdate: extracting verified archive: %w", err)
	}
	return Release{
		Tag: latest.TagName, AssetName: assetName, ArchiveSHA256: digest,
		WorkflowIdentity: provenance.WorkflowIdentity, SourceCommit: provenance.SourceCommit,
		Binary: binary, Bootstrap: bootstrap,
	}, nil
}

func (c *Client) verifyAttestations(
	ctx context.Context,
	digest, tag, asset string,
) (Provenance, error) {
	next := fmt.Sprintf(
		"%s/repos/%s/attestations/sha256:%s?per_page=100&predicate_type=provenance",
		c.apiBase, version.Repo, digest)
	totalBytes := int64(0)
	bundleURLs := make([]string, 0)
	for page := 0; next != ""; page++ {
		if page >= c.limits.pages {
			return Provenance{}, fmt.Errorf("releaseupdate: attestation list exceeds %d pages", c.limits.pages)
		}
		body, header, err := c.get(ctx, next, requestAttestationList, c.limits.attestationList)
		if err != nil {
			return Provenance{}, fmt.Errorf("releaseupdate: listing attestations: %w", err)
		}
		totalBytes += int64(len(body))
		if totalBytes > c.limits.attestationList {
			return Provenance{}, errors.New("releaseupdate: aggregate attestation list exceeds limit")
		}
		var pageResult struct {
			Attestations []struct {
				BundleURL string `json:"bundle_url"`
			} `json:"attestations"`
		}
		if err := json.Unmarshal(body, &pageResult); err != nil {
			return Provenance{}, fmt.Errorf("releaseupdate: decoding attestation list: %w", err)
		}
		for _, attestation := range pageResult.Attestations {
			bundleURLs = append(bundleURLs, attestation.BundleURL)
			if len(bundleURLs) > c.limits.bundles {
				return Provenance{}, fmt.Errorf(
					"releaseupdate: more than %d attestation bundles", c.limits.bundles)
			}
		}
		next, err = nextLink(header.Get("Link"))
		if err != nil {
			return Provenance{}, err
		}
	}
	if len(bundleURLs) == 0 {
		return Provenance{}, errors.New("releaseupdate: no provenance attestation exists for archive digest")
	}

	var failures []string
	for _, bundleURL := range bundleURLs {
		bundle, _, err := c.get(ctx, bundleURL, requestBundle, c.limits.bundle)
		if err == nil {
			var provenance Provenance
			provenance, err = c.verifier.Verify(ctx, bundle, digest, tag, asset)
			if err == nil {
				return provenance, nil
			}
		}
		failures = append(failures, err.Error())
	}
	return Provenance{}, fmt.Errorf(
		"releaseupdate: no attestation passed complete verification: %s",
		strings.Join(failures, "; "))
}

func (c *Client) get(
	ctx context.Context,
	raw string,
	kind requestKind,
	limit int64,
) ([]byte, http.Header, error) {
	target, err := url.Parse(raw)
	if err != nil || !c.allowURL(target, kind) {
		return nil, nil, fmt.Errorf("URL is outside the fixed allowlist: %q", raw)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Del("Authorization")

	client := *c.httpClient
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if kind == requestBundle {
			return errors.New("releaseupdate: attestation bundle redirects are forbidden")
		}
		if len(via) >= 5 {
			return errors.New("releaseupdate: more than five redirects")
		}
		if !c.allowURL(next.URL, kind) {
			return fmt.Errorf("releaseupdate: redirect escaped allowlist: %s", next.URL)
		}
		next.Header.Del("Authorization")
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, nil, fmt.Errorf("Content-Length %d exceeds %d", resp.ContentLength, limit)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, nil, err
	}
	if int64(len(body)) > limit {
		return nil, nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return body, resp.Header.Clone(), nil
}

func nextLink(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	for _, part := range strings.Split(raw, ",") {
		fields := strings.Split(part, ";")
		if len(fields) < 2 {
			continue
		}
		isNext := false
		for _, field := range fields[1:] {
			if strings.TrimSpace(field) == `rel="next"` {
				isNext = true
			}
		}
		if !isNext {
			continue
		}
		value := strings.TrimSpace(fields[0])
		if len(value) < 3 || value[0] != '<' || value[len(value)-1] != '>' {
			return "", errors.New("releaseupdate: malformed attestation pagination link")
		}
		return value[1 : len(value)-1], nil
	}
	return "", nil
}

func productionURLAllowed(u *url.URL, kind requestKind) bool {
	if u == nil || u.Scheme != "https" || u.User != nil {
		return false
	}
	if port := u.Port(); port != "" && port != "443" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || net.ParseIP(host) != nil {
		return false
	}
	switch kind {
	case requestMetadata, requestAttestationList:
		return host == "api.github.com"
	case requestAsset:
		return host == "github.com" ||
			host == "release-assets.githubusercontent.com" ||
			host == "objects.githubusercontent.com"
	case requestBundle:
		return host == "tmaproduction.blob.core.windows.net" &&
			strings.HasPrefix(u.EscapedPath(), "/attestations/")
	case requestTUF:
		return host == "tuf-repo-cdn.sigstore.dev"
	default:
		return false
	}
}
