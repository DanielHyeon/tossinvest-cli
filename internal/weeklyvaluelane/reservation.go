package weeklyvaluelane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MarketWeekEvidence struct {
	Market                                          Market
	Provider, TimeZone, SessionDate, StableIdentity string
	Official                                        bool
	CalendarGeneration, CalendarDigest              string
	ObservedAt, FreshUntil                          time.Time
}

func ValidateMarketWeek(week MarketWeekEvidence, evaluatedAt time.Time) RefusalCode {
	wantProvider, wantZone := "", ""
	switch week.Market {
	case MarketKR:
		wantProvider, wantZone = "XKRX_OFFICIAL", "Asia/Seoul"
	case MarketUS:
		wantProvider, wantZone = "XNYS_OFFICIAL", "America/New_York"
	default:
		return RefusalCalendarInvalid
	}
	for _, identity := range []string{week.Provider, week.TimeZone, week.SessionDate, week.StableIdentity, week.CalendarGeneration, week.CalendarDigest} {
		if !validBoundedIdentity(identity) {
			return RefusalCalendarInvalid
		}
	}
	if !week.Official || week.Provider != wantProvider || week.TimeZone != wantZone || week.ObservedAt.IsZero() || week.FreshUntil.IsZero() || evaluatedAt.IsZero() ||
		evaluatedAt.Before(week.ObservedAt) || evaluatedAt.After(week.FreshUntil) {
		return RefusalCalendarInvalid
	}
	location, err := time.LoadLocation(wantZone)
	if err != nil {
		return RefusalCalendarInvalid
	}
	session, err := time.ParseInLocation("2006-01-02", week.SessionDate, location)
	if err != nil || session.Weekday() != time.Monday {
		return RefusalCalendarInvalid
	}
	year, isoWeek := session.ISOWeek()
	if week.StableIdentity != stableMarketWeekIdentity(week.Market, year, isoWeek) {
		return RefusalCalendarInvalid
	}
	return ""
}

func stableMarketWeekIdentity(market Market, year, isoWeek int) string {
	exchange := map[Market]string{MarketKR: "KR-XKRX", MarketUS: "US-XNYS"}[market]
	if exchange == "" || year < 1 || year > 9999 || isoWeek < 1 || isoWeek > 53 {
		return ""
	}
	return fmt.Sprintf("%s-%04d-W%02d", exchange, year, isoWeek)
}

func CanonicalReservationKey(campaignID string, week MarketWeekEvidence) string {
	return strings.Join([]string{campaignID, string(week.Market), week.StableIdentity}, "|")
}

func reservationScopeKey(campaignID string, market Market) string {
	return strings.Join([]string{campaignID, string(market)}, "|")
}

type ReservationStatus string

const (
	ReservationActive   ReservationStatus = "ACTIVE"
	ReservationConsumed ReservationStatus = "CONSUMED"
	ReservationReleased ReservationStatus = "RELEASED"
)

type ReservationEntry struct {
	ReservationID, CampaignID, Key string
	MarketWeek                     MarketWeekEvidence
	PlannedOrdinal                 int
	Status                         ReservationStatus
}

type reservationReceipt struct {
	Fingerprint string
}

type reservationScopeState struct {
	Version          uint64
	PositiveLegCount int
	ConsumedOrdinals [7]bool
}

type ReservationState struct {
	entries  map[string]ReservationEntry
	receipts map[string]reservationReceipt
	scopes   map[string]reservationScopeState
	seal     [32]byte
}

func NewReservationState() ReservationState {
	state := ReservationState{entries: map[string]ReservationEntry{}, receipts: map[string]reservationReceipt{}, scopes: map[string]reservationScopeState{}}
	state.seal = reservationStateSeal(state)
	return state
}

func (state ReservationState) Entry(key string) (ReservationEntry, bool) {
	entry, ok := state.entries[key]
	return entry, ok
}

func (state ReservationState) Len() int { return len(state.entries) }

func (state ReservationState) ScopeVersion(campaignID string, market Market) uint64 {
	return state.scopes[reservationScopeKey(campaignID, market)].Version
}

func (state ReservationState) PositiveLegCount(campaignID string, market Market) int {
	return state.scopes[reservationScopeKey(campaignID, market)].PositiveLegCount
}

func (state ReservationState) NextPlannedOrdinal(campaignID string, market Market) int {
	count := state.PositiveLegCount(campaignID, market)
	if count >= 7 {
		return 0
	}
	return count + 1
}

type ReservationAction string

const (
	ReservationReserve      ReservationAction = "RESERVE"
	ReservationPositiveFill ReservationAction = "POSITIVE_FILL"
	ReservationZeroRelease  ReservationAction = "ZERO_RELEASE"
)

type ReservationCommand struct {
	Action               ReservationAction
	ExpectedVersion      uint64
	CampaignID           string
	MarketWeek           MarketWeekEvidence
	ReservationID        string
	IdempotencyKey       string
	PlannedOrdinal       int
	PositiveFillQuantity uint64
	AuthoritativeZero    bool
	PendingAttempts      int
	EvaluatedAt          time.Time
	evaluationSeal       [32]byte
}

func authorizeReservationCommand(command ReservationCommand) ReservationCommand {
	if command.EvaluatedAt.IsZero() || !validReservationCommandIdentityFields(command) {
		return command
	}
	command.evaluationSeal = reservationEvaluationSeal(command)
	return command
}

type ReservationResult struct {
	Applied, Duplicate bool
	Code               RefusalCode
}

func ApplyReservation(state ReservationState, command ReservationCommand) (ReservationState, ReservationResult) {
	if command.Action == ReservationPositiveFill {
		return state, ReservationResult{Code: RefusalReservationConflict}
	}
	return applyReservationTransition(state, command, false)
}

func applyReservationTransition(state ReservationState, command ReservationCommand, allowPositiveFill bool) (ReservationState, ReservationResult) {
	if !validReservationState(state) || !validReservationCommandIdentity(command) {
		return state, ReservationResult{Code: RefusalReservationConflict}
	}
	scopeKey := reservationScopeKey(command.CampaignID, command.MarketWeek.Market)
	receiptKey := scopeKey + "|" + command.IdempotencyKey
	fingerprint := reservationCommandFingerprint(command)
	if receipt, exists := state.receipts[receiptKey]; exists {
		if receipt.Fingerprint == fingerprint {
			return state, ReservationResult{Duplicate: true}
		}
		return state, ReservationResult{Duplicate: true, Code: RefusalReservationConflict}
	}
	scope := state.scopes[scopeKey]
	if command.ExpectedVersion != scope.Version {
		return state, ReservationResult{Code: RefusalVersionConflict}
	}
	next := cloneReservationState(state)
	switch command.Action {
	case ReservationReserve:
		if code := ValidateMarketWeek(command.MarketWeek, command.EvaluatedAt); code != "" {
			return state, ReservationResult{Code: code}
		}
		if scope.PositiveLegCount >= 7 {
			return state, ReservationResult{Code: RefusalPlanExhausted}
		}
		if command.PlannedOrdinal != scope.PositiveLegCount+1 || activeReservationExists(state, scopeKey) {
			return state, ReservationResult{Code: RefusalReservationConflict}
		}
		key := CanonicalReservationKey(command.CampaignID, command.MarketWeek)
		if _, exists := state.entries[key]; exists || reservationIDExists(state, command.ReservationID) {
			return state, ReservationResult{Code: RefusalReservationConflict}
		}
		next.entries[key] = ReservationEntry{ReservationID: command.ReservationID, CampaignID: command.CampaignID, Key: key, MarketWeek: command.MarketWeek, PlannedOrdinal: command.PlannedOrdinal, Status: ReservationActive}
	case ReservationPositiveFill:
		if !allowPositiveFill {
			return state, ReservationResult{Code: RefusalReservationConflict}
		}
		key, entry, ok := findReservation(state, command.ReservationID)
		if !ok {
			return state, ReservationResult{Code: RefusalReservationMissing}
		}
		if entry.Status != ReservationActive {
			return state, ReservationResult{Code: RefusalReservationTerminal}
		}
		if !reservationCommandMatchesEntry(command, entry) || command.PositiveFillQuantity == 0 || scope.PositiveLegCount >= 7 ||
			command.PlannedOrdinal != scope.PositiveLegCount+1 || scope.ConsumedOrdinals[command.PlannedOrdinal-1] {
			return state, ReservationResult{Code: RefusalReservationConflict}
		}
		entry.Status = ReservationConsumed
		next.entries[key] = entry
		scope.PositiveLegCount++
		scope.ConsumedOrdinals[command.PlannedOrdinal-1] = true
	case ReservationZeroRelease:
		key, entry, ok := findReservation(state, command.ReservationID)
		if !ok {
			return state, ReservationResult{Code: RefusalReservationMissing}
		}
		if entry.Status != ReservationActive {
			return state, ReservationResult{Code: RefusalReservationTerminal}
		}
		if !reservationCommandMatchesEntry(command, entry) || !command.AuthoritativeZero || command.PendingAttempts != 0 || command.PositiveFillQuantity != 0 {
			return state, ReservationResult{Code: RefusalReservationConflict}
		}
		entry.Status = ReservationReleased
		next.entries[key] = entry
	default:
		return state, ReservationResult{Code: RefusalReservationConflict}
	}
	scope.Version++
	next.scopes[scopeKey] = scope
	next.receipts[receiptKey] = reservationReceipt{Fingerprint: fingerprint}
	next.seal = reservationStateSeal(next)
	return next, ReservationResult{Applied: true}
}

func validReservationState(state ReservationState) bool {
	return state.entries != nil && state.receipts != nil && state.scopes != nil && state.seal != ([32]byte{}) && state.seal == reservationStateSeal(state)
}

func validReservationCommandIdentity(command ReservationCommand) bool {
	return validReservationCommandIdentityFields(command) && command.evaluationSeal != ([32]byte{}) && command.evaluationSeal == reservationEvaluationSeal(command)
}

func validReservationCommandIdentityFields(command ReservationCommand) bool {
	if !validBoundedIdentity(command.CampaignID) || !validBoundedIdentity(command.ReservationID) || !validBoundedIdentity(command.IdempotencyKey) ||
		(command.MarketWeek.Market != MarketKR && command.MarketWeek.Market != MarketUS) || command.PlannedOrdinal < 1 || command.PlannedOrdinal > 7 {
		return false
	}
	return true
}

func reservationEvaluationSeal(command ReservationCommand) [32]byte {
	return sha256.Sum256([]byte("weekly-value-trusted-reservation-evaluation-v1\x00" + reservationCommandFingerprint(command)))
}

func reservationCommandMatchesEntry(command ReservationCommand, entry ReservationEntry) bool {
	return command.CampaignID == entry.CampaignID && command.MarketWeek.Market == entry.MarketWeek.Market &&
		command.MarketWeek.StableIdentity == entry.MarketWeek.StableIdentity && command.PlannedOrdinal == entry.PlannedOrdinal
}

func activeReservationExists(state ReservationState, scopeKey string) bool {
	for _, entry := range state.entries {
		if reservationScopeKey(entry.CampaignID, entry.MarketWeek.Market) == scopeKey && entry.Status == ReservationActive {
			return true
		}
	}
	return false
}

func cloneReservationState(state ReservationState) ReservationState {
	next := ReservationState{entries: make(map[string]ReservationEntry, len(state.entries)), receipts: make(map[string]reservationReceipt, len(state.receipts)), scopes: make(map[string]reservationScopeState, len(state.scopes))}
	for key, value := range state.entries {
		next.entries[key] = value
	}
	for key, value := range state.receipts {
		next.receipts[key] = value
	}
	for key, value := range state.scopes {
		next.scopes[key] = value
	}
	return next
}

func findReservation(state ReservationState, reservationID string) (string, ReservationEntry, bool) {
	for key, entry := range state.entries {
		if entry.ReservationID == reservationID {
			return key, entry, true
		}
	}
	return "", ReservationEntry{}, false
}

func reservationStateSeal(state ReservationState) [32]byte {
	parts := []string{"weekly-value-reservation-state-v1"}
	entryKeys := make([]string, 0, len(state.entries))
	for key := range state.entries {
		entryKeys = append(entryKeys, key)
	}
	sort.Strings(entryKeys)
	for _, key := range entryKeys {
		entry := state.entries[key]
		parts = append(parts, key, entry.ReservationID, entry.CampaignID, entry.Key, string(entry.MarketWeek.Market), entry.MarketWeek.Provider,
			entry.MarketWeek.TimeZone, entry.MarketWeek.SessionDate, entry.MarketWeek.StableIdentity, strconv.FormatBool(entry.MarketWeek.Official),
			entry.MarketWeek.CalendarGeneration, entry.MarketWeek.CalendarDigest, canonicalTime(entry.MarketWeek.ObservedAt), canonicalTime(entry.MarketWeek.FreshUntil),
			strconv.Itoa(entry.PlannedOrdinal), string(entry.Status))
	}
	receiptKeys := make([]string, 0, len(state.receipts))
	for key := range state.receipts {
		receiptKeys = append(receiptKeys, key)
	}
	sort.Strings(receiptKeys)
	for _, key := range receiptKeys {
		parts = append(parts, key, state.receipts[key].Fingerprint)
	}
	scopeKeys := make([]string, 0, len(state.scopes))
	for key := range state.scopes {
		scopeKeys = append(scopeKeys, key)
	}
	sort.Strings(scopeKeys)
	for _, key := range scopeKeys {
		scope := state.scopes[key]
		parts = append(parts, key, strconv.FormatUint(scope.Version, 10), strconv.Itoa(scope.PositiveLegCount))
		for _, consumed := range scope.ConsumedOrdinals {
			parts = append(parts, strconv.FormatBool(consumed))
		}
	}
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}

func reservationIDExists(state ReservationState, reservationID string) bool {
	_, _, exists := findReservation(state, reservationID)
	return exists
}

func reservationCommandFingerprint(command ReservationCommand) string {
	parts := []string{string(command.Action), strconv.FormatUint(command.ExpectedVersion, 10), command.CampaignID, string(command.MarketWeek.Market), command.MarketWeek.Provider,
		command.MarketWeek.TimeZone, command.MarketWeek.SessionDate, command.MarketWeek.StableIdentity, strconv.FormatBool(command.MarketWeek.Official),
		command.MarketWeek.CalendarGeneration, command.MarketWeek.CalendarDigest, canonicalTime(command.MarketWeek.ObservedAt), canonicalTime(command.MarketWeek.FreshUntil),
		command.ReservationID, strconv.Itoa(command.PlannedOrdinal), strconv.FormatUint(command.PositiveFillQuantity, 10), strconv.FormatBool(command.AuthoritativeZero),
		strconv.Itoa(command.PendingAttempts), canonicalTime(command.EvaluatedAt)}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
