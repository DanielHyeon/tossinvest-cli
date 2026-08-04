package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const strategyActivationEnvPrefix = "TOSSOS_STRATEGY_ACTIVATION_"

type strategyScheduleCalendarReader interface {
	TypedMarketCalendar(context.Context, string, string) (official.MarketCalendarResponse, error)
}

type strategyScheduleAuthorityLoader struct {
	configDir string
	clock     clock.Clock
	calendar  strategyScheduleCalendarReader
	getenv    func(string) string
}

func newStrategyScheduleAuthorityLoader(configDir string, clk clock.Clock, calendar strategyScheduleCalendarReader, getenv func(string) string) *strategyScheduleAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	return &strategyScheduleAuthorityLoader{configDir: filepath.Clean(strings.TrimSpace(configDir)), clock: clk, calendar: calendar, getenv: getenv}
}

// StrategyScheduleMarketSnapshot is an observation only. In particular, the
// activation and calendar are retained in the private authority pair below.
type StrategyScheduleMarketSnapshot struct {
	Market                   StrategyMarket
	DesiredEnabled           bool
	DesiredAutostart         bool
	Ready                    bool
	Reason                   scheduler.ResumeReason
	CalendarVersion          string
	ActivationManifestDigest string
	ConfigDigest             string
}

type PairedStrategyScheduleSnapshot struct {
	ObservedAt time.Time
	KR         StrategyScheduleMarketSnapshot
	US         StrategyScheduleMarketSnapshot
}

func (snapshot PairedStrategyScheduleSnapshot) For(market StrategyMarket) StrategyScheduleMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyScheduleMarketSnapshot{Market: market, Reason: scheduler.ResumeVerificationFailed, ConfigDigest: strategyRuntimeConfigDigest()}
}

type strategyScheduleMarketAuthority struct {
	market   StrategyMarket
	desired  scheduler.DesiredState
	calendar scheduler.CalendarSnapshot
	restore  scheduler.RestoreResult
	snapshot StrategyScheduleMarketSnapshot
}

type strategyScheduleAuthorityPair struct {
	observedAt time.Time
	kr         strategyScheduleMarketAuthority
	us         strategyScheduleMarketAuthority
}

func (pair strategyScheduleAuthorityPair) forMarket(market StrategyMarket) strategyScheduleMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyScheduleMarketAuthority{market: market}
}

func (pair strategyScheduleAuthorityPair) Snapshot() PairedStrategyScheduleSnapshot {
	return PairedStrategyScheduleSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type preparedStrategySchedule struct {
	market         StrategyMarket
	desired        scheduler.DesiredState
	calendar       scheduler.CalendarSnapshot
	current        scheduler.CurrentBinding
	verifier       scheduler.ActivationVerifier
	manifestDigest string
	failed         bool
}

func (loader *strategyScheduleAuthorityLoader) collect(ctx context.Context) strategyScheduleAuthorityPair {
	if loader == nil || loader.clock == nil || ctx == nil {
		return failedStrategySchedulePair(time.Time{})
	}
	observedAt := loader.clock.Now().UTC()
	if observedAt.IsZero() {
		return failedStrategySchedulePair(time.Time{})
	}
	prepared := make(chan preparedStrategySchedule, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() { prepared <- loader.prepare(ctx, market, observedAt) }()
	}
	byMarket := make(map[StrategyMarket]preparedStrategySchedule, 2)
	for index := 0; index < 2; index++ {
		value := <-prepared
		byMarket[value.market] = value
	}
	requests := [2]scheduler.PairedRestoreRequest{}
	for index, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		value := byMarket[market]
		requests[index] = scheduler.PairedRestoreRequest{Market: strategySchedulerMarket(market), Desired: value.desired,
			Current: value.current, Verifier: value.verifier}
	}
	restored, err := scheduler.RestorePairedProduction(ctx, requests, func() time.Time { return observedAt })
	if err != nil {
		return failedStrategySchedulePair(observedAt)
	}
	pair := strategyScheduleAuthorityPair{observedAt: observedAt}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		value := byMarket[market]
		result := restored.For(strategySchedulerMarket(market))
		if value.failed {
			result = scheduler.RestoreResult{Reason: scheduler.ResumeVerificationFailed}
		}
		snapshot := StrategyScheduleMarketSnapshot{Market: market, DesiredEnabled: value.desired.Enabled,
			DesiredAutostart: value.desired.AutoStart, Ready: result.Restored && result.Activation != nil,
			Reason: result.Reason, CalendarVersion: value.calendar.Version,
			ActivationManifestDigest: value.manifestDigest, ConfigDigest: strategyRuntimeConfigDigest()}
		authority := strategyScheduleMarketAuthority{market: market, desired: value.desired, calendar: value.calendar, restore: result, snapshot: snapshot}
		if market == StrategyMarketKR {
			pair.kr = authority
		} else {
			pair.us = authority
		}
	}
	return pair
}

// collectMarket is the final no-byte-sent revalidator. Unlike collect it never
// starts or waits for the peer market, so a blocked KR source cannot delay US
// dispatch and a blocked US source cannot delay KR dispatch.
func (loader *strategyScheduleAuthorityLoader) collectMarket(ctx context.Context, market StrategyMarket) strategyScheduleMarketAuthority {
	failed := strategyScheduleMarketAuthority{market: market, snapshot: StrategyScheduleMarketSnapshot{Market: market,
		Reason: scheduler.ResumeVerificationFailed, ConfigDigest: strategyRuntimeConfigDigest()}}
	if loader == nil || loader.clock == nil || ctx == nil || (market != StrategyMarketKR && market != StrategyMarketUS) {
		return failed
	}
	observedAt := loader.clock.Now().UTC()
	if observedAt.IsZero() {
		return failed
	}
	prepared := loader.prepare(ctx, market, observedAt)
	result := scheduler.Restore(ctx, prepared.desired, prepared.current, prepared.verifier, observedAt)
	if prepared.failed {
		result = scheduler.RestoreResult{Reason: scheduler.ResumeVerificationFailed}
	}
	snapshot := StrategyScheduleMarketSnapshot{Market: market, DesiredEnabled: prepared.desired.Enabled,
		DesiredAutostart: prepared.desired.AutoStart, Ready: result.Restored && result.Activation != nil,
		Reason: result.Reason, CalendarVersion: prepared.calendar.Version,
		ActivationManifestDigest: prepared.manifestDigest, ConfigDigest: strategyRuntimeConfigDigest()}
	return strategyScheduleMarketAuthority{market: market, desired: prepared.desired, calendar: prepared.calendar,
		restore: result, snapshot: snapshot}
}

func (loader *strategyScheduleAuthorityLoader) prepare(ctx context.Context, market StrategyMarket, observedAt time.Time) preparedStrategySchedule {
	schedulerMarket := strategySchedulerMarket(market)
	value := preparedStrategySchedule{market: market, desired: scheduler.DefaultDesiredState(),
		current: scheduler.CurrentBinding{SchedulerVersion: scheduler.SchedulerVersion, Market: schedulerMarket,
			Session: scheduler.SessionRegular, ConfigVersion: strategyRuntimeConfigDigest(), BuildDigest: strategyRuntimeBuildDigest()}}
	name := strategyDesiredFileName(market)
	if name == "" || !filepath.IsAbs(loader.configDir) {
		value.failed = true
		return value
	}
	desired, err := scheduler.NewDesiredStore(filepath.Join(loader.configDir, name)).LoadAt(ctx, observedAt)
	if err != nil {
		value.failed = true
		return value
	}
	value.desired = desired
	if !desired.Enabled || !desired.AutoStart {
		return value
	}
	config, pinErr := loader.activationConfig(market)
	if pinErr == nil {
		value.verifier, pinErr = scheduler.NewProductionActivationVerifier(config)
		value.manifestDigest = config.ManifestDigest
	}
	if loader.calendar == nil {
		value.failed = true
		return value
	}
	clockMarket := strategyClockMarket(market)
	location, err := clockMarket.Location()
	if err != nil {
		value.failed = true
		return value
	}
	payload, err := loader.calendar.TypedMarketCalendar(ctx, string(market), observedAt.In(location).Format("2006-01-02"))
	if err != nil {
		value.failed = true
		return value
	}
	calendar, err := scheduler.AdaptOfficialCalendar(clockMarket, payload, observedAt)
	if err != nil {
		value.failed = true
		return value
	}
	value.calendar = calendar
	value.current.CalendarVersion = calendar.Version
	return value
}

func (loader *strategyScheduleAuthorityLoader) activationConfig(market StrategyMarket) (scheduler.ProductionActivationConfig, error) {
	digest := strings.TrimSpace(loader.getenv(strategyActivationManifestDigestEnv(market)))
	keyID := strings.TrimSpace(loader.getenv(strategyActivationKeyIDEnv(market)))
	encoded := strings.TrimSpace(loader.getenv(strategyActivationPublicKeyEnv(market)))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		return scheduler.ProductionActivationConfig{}, errors.New("engine: strategy activation public key pin is unavailable")
	}
	return scheduler.ProductionActivationConfig{ConfigDir: loader.configDir, Market: strategySchedulerMarket(market),
		ManifestDigest: digest, TrustedKeyID: keyID, TrustedKey: ed25519.PublicKey(key)}, nil
}

func strategyActivationManifestDigestEnv(market StrategyMarket) string {
	return strategyActivationEnvPrefix + string(market) + "_MANIFEST_SHA256"
}

func strategyActivationKeyIDEnv(market StrategyMarket) string {
	return strategyActivationEnvPrefix + string(market) + "_KEY_ID"
}

func strategyActivationPublicKeyEnv(market StrategyMarket) string {
	return strategyActivationEnvPrefix + string(market) + "_PUBLIC_KEY_BASE64"
}

func strategyDesiredFileName(market StrategyMarket) string {
	return scheduler.ProductionDesiredFileName(strategySchedulerMarket(market))
}

func strategySchedulerMarket(market StrategyMarket) scheduler.MarketScope {
	if market == StrategyMarketKR {
		return scheduler.MarketScopeKR
	}
	if market == StrategyMarketUS {
		return scheduler.MarketScopeUS
	}
	return scheduler.MarketScopeNone
}

func strategyClockMarket(market StrategyMarket) clock.Market {
	if market == StrategyMarketKR {
		return clock.MarketKR
	}
	if market == StrategyMarketUS {
		return clock.MarketUS
	}
	return ""
}

func strategyRuntimeBuildDigest() string {
	build, _ := productionProtectionDigests()
	return "sha256:" + strings.TrimPrefix(build, "sha256:")
}

func strategyRuntimeConfigDigest() string {
	return strategyRuntimeConfigDigestFor(strategyrouter.RouterID, strategyrouter.RouterRelease, strategyflow.Descriptors())
}

func strategyRuntimeConfigDigestFor(routerID, routerRelease string, descriptors []strategyflow.Descriptor) string {
	type identity struct {
		Market, Horizon, LaneID, LaneVersion, Release string
	}
	values := make([]identity, 0, len(descriptors))
	for _, descriptor := range descriptors {
		values = append(values, identity{Market: string(descriptor.Market), Horizon: string(descriptor.Horizon),
			LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, Release: descriptor.Release})
	}
	sort.Slice(values, func(i, j int) bool {
		left, _ := json.Marshal(values[i])
		right, _ := json.Marshal(values[j])
		return string(left) < string(right)
	})
	payload := struct {
		Domain, RouterID, RouterRelease string
		Lanes                           []identity
	}{"TossOS/strategy-runtime-config/v1", routerID, routerRelease, values}
	data, _ := json.Marshal(payload)
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func failedStrategySchedulePair(observedAt time.Time) strategyScheduleAuthorityPair {
	market := func(value StrategyMarket) strategyScheduleMarketAuthority {
		return strategyScheduleMarketAuthority{market: value, snapshot: StrategyScheduleMarketSnapshot{Market: value,
			Reason: scheduler.ResumeVerificationFailed, ConfigDigest: strategyRuntimeConfigDigest()}}
	}
	return strategyScheduleAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
