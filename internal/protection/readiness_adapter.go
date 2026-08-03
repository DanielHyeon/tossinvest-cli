package protection

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
)

const RefusalProviderUnavailable = protectionreadiness.RefusalProviderUnavailable

// ReadinessRefusal is the typed, non-authoritative reason an entry cannot use a
// current snapshot.
type ReadinessRefusal struct {
	Code   protectionreadiness.RefusalCode
	Detail string
}

func (refusal *ReadinessRefusal) Error() string { return refusal.Detail }

// Checkpoint is the exact market snapshot accepted before dispatch. Its fields
// are private so callers cannot synthesize a generation or identity.
type Checkpoint struct {
	market     protectionreadiness.Market
	generation uint64
	identity   string
	seal       [32]byte
}

func (checkpoint Checkpoint) Valid() bool {
	return checkpoint.identity != "" && checkpoint.generation > 0 && checkpoint.seal == checkpointSeal(checkpoint)
}

func checkpointSeal(checkpoint Checkpoint) [32]byte {
	return readinessAdapterHash(string(checkpoint.market), stringUintAdapter(checkpoint.generation), checkpoint.identity)
}

// ReadinessAdapter binds an immutable snapshot provider to the account/profile
// the gateway actually owns. Its private fields prevent config or booleans from
// manufacturing a WIRED decision.
type ReadinessAdapter struct {
	provider  protectionreadiness.SnapshotProvider
	accountID string
	profileID string
	seal      [32]byte
}

func NewReadinessAdapter(provider protectionreadiness.SnapshotProvider, accountID, profileID string) (*ReadinessAdapter, error) {
	accountID, profileID = strings.TrimSpace(accountID), strings.TrimSpace(profileID)
	if provider == nil || accountID == "" || profileID == "" {
		return nil, errors.New("protection: readiness provider and exact account/profile are required")
	}
	adapter := &ReadinessAdapter{provider: provider, accountID: accountID, profileID: profileID}
	adapter.seal = readinessAdapterHash(accountID, profileID, protectionreadiness.ReadinessRelease)
	return adapter, nil
}

func DefaultReadinessAdapter(accountID, profileID string) (*ReadinessAdapter, error) {
	return NewReadinessAdapter(protectionreadiness.DefaultProvider(), accountID, profileID)
}

func (adapter *ReadinessAdapter) Check(ctx context.Context, market string, now time.Time, previous Checkpoint) (Checkpoint, *ReadinessRefusal) {
	if adapter == nil || adapter.provider == nil || adapter.seal != readinessAdapterHash(adapter.accountID, adapter.profileID, protectionreadiness.ReadinessRelease) {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalStateCorrupt, Detail: "protection readiness adapter is missing or corrupt"}
	}
	var scopedMarket protectionreadiness.Market
	switch market {
	case "kr":
		scopedMarket = protectionreadiness.MarketKR
	case "us":
		scopedMarket = protectionreadiness.MarketUS
	default:
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalInvalid, Detail: "unsupported protection readiness market"}
	}
	snapshot, err := adapter.provider.Current(ctx)
	if err != nil {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalProviderUnavailable, Detail: "protection readiness provider unavailable: " + err.Error()}
	}
	decision := snapshot.Dispatch(protectionreadiness.DispatchScope{AccountID: adapter.accountID, ProfileID: adapter.profileID, Market: scopedMarket}, now)
	if !decision.Allowed {
		return Checkpoint{}, &ReadinessRefusal{Code: decision.Code, Detail: "protection readiness refused for " + market + ": " + string(decision.Code)}
	}
	checkpoint := Checkpoint{market: scopedMarket, generation: decision.Generation, identity: decision.SnapshotID}
	checkpoint.seal = checkpointSeal(checkpoint)
	if previous.Valid() && (previous.market != checkpoint.market || previous.generation != checkpoint.generation || previous.identity != checkpoint.identity) {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalSnapshotDrift, Detail: "protection readiness changed before dispatch"}
	}
	return checkpoint, nil
}

func readinessAdapterHash(values ...string) [32]byte {
	hash := sha256.New()
	for _, value := range values {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return digest
}

func stringUintAdapter(value uint64) string {
	if value == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = digits[value%10]
		value /= 10
	}
	return string(buffer[index:])
}
