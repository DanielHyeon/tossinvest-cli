package strategyaccount

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

const (
	productionSchema       = "strategy-account-snapshot:v1"
	productionDomain       = "TossOS/strategy-account-snapshot/ed25519/v1"
	productionAlgorithm    = "Ed25519"
	productionMaximumBytes = 1 << 20
	productionWindow       = 5 * time.Second
)

var ErrProductionAccountUnavailable = errors.New("strategy account: production authority unavailable")

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type ProductionConfig struct {
	ConfigDir, AccountRef, AccountCurrency, Symbol string
	Market                                         Market
	ManifestDigest, TrustedKeyID                   string
	TrustedKey                                     ed25519.PublicKey
	ObservedAt                                     time.Time
}

type productionBody struct {
	SchemaVersion         string   `json:"schema_version"`
	Domain                string   `json:"domain"`
	SignatureAlgorithm    string   `json:"signature_algorithm"`
	KeyID                 string   `json:"key_id"`
	Generation            uint64   `json:"generation"`
	Market                Market   `json:"market"`
	AccountRef            string   `json:"account_ref"`
	AccountCurrency       string   `json:"account_currency"`
	QuoteCurrency         string   `json:"quote_currency"`
	Source                string   `json:"source"`
	SourceDigest          string   `json:"source_digest"`
	Official              bool     `json:"official"`
	ObservedAt            string   `json:"observed_at"`
	FreshUntil            string   `json:"fresh_until"`
	Revoked               bool     `json:"revoked"`
	KillSwitchActive      bool     `json:"kill_switch_active"`
	OperatingMode         string   `json:"operating_mode"`
	EntryBlockedLatch     bool     `json:"entry_blocked_latch"`
	EntryBlockedReason    string   `json:"entry_blocked_reason"`
	AllowedSymbols        []string `json:"allowed_symbols"`
	HeldQuantity          string   `json:"held_quantity"`
	CashAvailable         string   `json:"cash_available"`
	OpenExposureBase      string   `json:"open_exposure_base"`
	DailyRealizedLossBase string   `json:"daily_realized_loss_base"`
	AccountEquityBase     string   `json:"account_equity_base"`
	SameDayEntryCount     int      `json:"same_day_entry_count"`
	LastEntryAt           string   `json:"last_entry_at"`
	PendingBuy            bool     `json:"pending_buy"`
	DuplicateOrder        bool     `json:"duplicate_order"`
}

type productionManifest struct {
	productionBody
	Signature string `json:"signature"`
}

type Authority struct {
	account        risk.AccountState
	openExposure   riskcalc.Money
	market         Market
	quoteCurrency  string
	observedAt     time.Time
	freshUntil     time.Time
	generation     uint64
	manifestDigest string
	identity       string
}

func FileName(market Market) string {
	if market == MarketKR {
		return "strategy-account-snapshot-KR.json"
	}
	if market == MarketUS {
		return "strategy-account-snapshot-US.json"
	}
	return ""
}

func (a Authority) AccountState() risk.AccountState {
	state := a.account
	state.AllowedSymbols = append([]string(nil), a.account.AllowedSymbols...)
	return state
}

func (a Authority) OpenExposure() riskcalc.Money { return a.openExposure }
func (a Authority) Market() Market               { return a.market }
func (a Authority) QuoteCurrency() string        { return a.quoteCurrency }
func (a Authority) ObservedAt() time.Time        { return a.observedAt }
func (a Authority) FreshUntil() time.Time        { return a.freshUntil }
func (a Authority) Generation() uint64           { return a.generation }
func (a Authority) ManifestDigest() string       { return a.manifestDigest }
func (a Authority) Identity() string             { return a.identity }

func LoadProductionAuthority(ctx context.Context, config ProductionConfig) (Authority, error) {
	if ctx == nil || config.ObservedAt.IsZero() || ctx.Err() != nil {
		return Authority{}, ErrProductionAccountUnavailable
	}
	config = canonicalConfig(config)
	owner, ownerOK := productionOwnerUID()
	name := FileName(config.Market)
	if !ownerOK || name == "" || !filepath.IsAbs(config.ConfigDir) || !validIdentity(config.AccountRef) ||
		!validCurrency(config.AccountCurrency) || !validSymbol(config.Symbol) || !validDigest(config.ManifestDigest) ||
		!validIdentity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize {
		return Authority{}, ErrProductionAccountUnavailable
	}
	data, err := readProductionFile(filepath.Join(config.ConfigDir, name), owner, 0o400, productionMaximumBytes)
	if err != nil || digest(data) != config.ManifestDigest {
		return Authority{}, ErrProductionAccountUnavailable
	}
	manifest, err := decodeManifest(data)
	if err != nil || !verifyManifest(manifest, config) {
		return Authority{}, ErrProductionAccountUnavailable
	}
	return sealAuthority(manifest.productionBody, config.ManifestDigest)
}

func canonicalConfig(config ProductionConfig) ProductionConfig {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.AccountRef = strings.TrimSpace(config.AccountRef)
	config.AccountCurrency = strings.ToUpper(strings.TrimSpace(config.AccountCurrency))
	config.Symbol = strings.ToUpper(strings.TrimSpace(config.Symbol))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	config.ObservedAt = config.ObservedAt.UTC()
	return config
}

func decodeManifest(data []byte) (productionManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionManifest
	if err := decoder.Decode(&manifest); err != nil {
		return productionManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return productionManifest{}, ErrProductionAccountUnavailable
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionManifest{}, ErrProductionAccountUnavailable
	}
	return manifest, nil
}

func verifyManifest(manifest productionManifest, config ProductionConfig) bool {
	body := manifest.productionBody
	if body.SchemaVersion != productionSchema || body.Domain != productionDomain || body.SignatureAlgorithm != productionAlgorithm ||
		body.KeyID != config.TrustedKeyID || body.Generation == 0 || body.Market != config.Market || body.AccountRef != config.AccountRef ||
		body.AccountCurrency != config.AccountCurrency || body.Source != "toss-open-api" || !body.Official || body.Revoked ||
		!validDigest(body.SourceDigest) {
		return false
	}
	quote := "KRW"
	if config.Market == MarketUS {
		quote = "USD"
	}
	if body.QuoteCurrency != quote {
		return false
	}
	observed, observedOK := exactUTC(body.ObservedAt)
	fresh, freshOK := exactUTC(body.FreshUntil)
	if !observedOK || !freshOK || fresh.Before(config.ObservedAt) || observed.After(config.ObservedAt) || fresh.Sub(observed) > productionWindow {
		return false
	}
	payload, err := json.Marshal(body)
	signature, signatureErr := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	return err == nil && signatureErr == nil && base64.StdEncoding.EncodeToString(signature) == manifest.Signature &&
		ed25519.Verify(config.TrustedKey, payload, signature)
}

func sealAuthority(body productionBody, manifestDigest string) (Authority, error) {
	allowed := append([]string(nil), body.AllowedSymbols...)
	if len(allowed) == 0 || body.SameDayEntryCount < 0 || !validWhole(body.HeldQuantity) ||
		!validDecimal(body.CashAvailable, false) || !validDecimal(body.OpenExposureBase, false) ||
		!validDecimal(body.DailyRealizedLossBase, false) || !validDecimal(body.AccountEquityBase, true) {
		return Authority{}, ErrProductionAccountUnavailable
	}
	if !sort.StringsAreSorted(allowed) {
		return Authority{}, ErrProductionAccountUnavailable
	}
	for index, symbol := range allowed {
		if !validSymbol(symbol) || index > 0 && allowed[index-1] == symbol {
			return Authority{}, ErrProductionAccountUnavailable
		}
	}
	mode := risk.OperatingMode(body.OperatingMode)
	if mode != risk.ModeNormal && mode != risk.ModeEntryBlocked && mode != risk.ModeHaltAll {
		return Authority{}, ErrProductionAccountUnavailable
	}
	var last time.Time
	if body.LastEntryAt != "" {
		var ok bool
		last, ok = exactUTC(body.LastEntryAt)
		if !ok || body.SameDayEntryCount == 0 {
			return Authority{}, ErrProductionAccountUnavailable
		}
	} else if body.SameDayEntryCount != 0 {
		return Authority{}, ErrProductionAccountUnavailable
	}
	observed, _ := exactUTC(body.ObservedAt)
	fresh, _ := exactUTC(body.FreshUntil)
	account := risk.AccountState{KillSwitchActive: body.KillSwitchActive, Mode: mode, EntryBlockedLatch: body.EntryBlockedLatch,
		EntryBlockedReason: body.EntryBlockedReason, AllowedSymbols: allowed, HeldQuantity: body.HeldQuantity,
		CashAvailable:     riskcalc.Money{Amount: body.CashAvailable, Currency: body.QuoteCurrency},
		OpenExposure:      riskcalc.Money{Amount: body.OpenExposureBase, Currency: body.AccountCurrency},
		DailyRealizedLoss: riskcalc.Money{Amount: body.DailyRealizedLossBase, Currency: body.AccountCurrency},
		AccountEquity:     riskcalc.Money{Amount: body.AccountEquityBase, Currency: body.AccountCurrency},
		SameDayEntryCount: body.SameDayEntryCount, LastEntryAt: last, PendingBuy: body.PendingBuy, DuplicateOrder: body.DuplicateOrder}
	identityBytes, _ := json.Marshal(body)
	return Authority{account: account, openExposure: account.OpenExposure, market: body.Market, quoteCurrency: body.QuoteCurrency,
		observedAt: observed, freshUntil: fresh, generation: body.Generation, manifestDigest: manifestDigest, identity: digest(identityBytes)}, nil
}

func exactUTC(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil && parsed.Location() == time.UTC && parsed.Format(time.RFC3339Nano) == value
}

func validWhole(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	return ok && parsed.Sign() >= 0 && parsed.String() == value
}

func validDecimal(value string, positive bool) bool {
	if value == "" || strings.TrimSpace(value) != value || strings.HasPrefix(value, "+") || strings.ContainsAny(value, "eE") {
		return false
	}
	parsed, ok := new(big.Rat).SetString(value)
	if !ok || parsed.Sign() < 0 || positive && parsed.Sign() == 0 {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || len(parts[0]) > 1 && parts[0][0] == '0' || len(parts) == 2 && (parts[1] == "" || strings.HasSuffix(parts[1], "0")) {
		return false
	}
	return true
}

func validIdentity(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n\t")
}

func validSymbol(value string) bool {
	return validIdentity(value) && value == strings.ToUpper(value) && !strings.Contains(value, "/")
}

func validCurrency(value string) bool { return len(value) == 3 && value == strings.ToUpper(value) }

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+sha256.Size*2 || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
