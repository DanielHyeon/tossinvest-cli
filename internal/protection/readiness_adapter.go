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
	contracts map[protectionreadiness.Market]supervisorContract
	seal      [32]byte
}

type supervisorContract struct {
	market                 protectionreadiness.Market
	sessionScope           string
	triggerSource          string
	replaceSemantics       string
	brokerCapabilityDigest string
	toolDigest             string
	seal                   [32]byte
}

type ReadinessRequest struct {
	Market    string
	OrderType string
	Quantity  uint64
}

func NewReadinessAdapter(provider protectionreadiness.SnapshotProvider, accountID, profileID string) (*ReadinessAdapter, error) {
	accountID, profileID = strings.TrimSpace(accountID), strings.TrimSpace(profileID)
	if provider == nil || accountID == "" || profileID == "" {
		return nil, errors.New("protection: readiness provider and exact account/profile are required")
	}
	adapter := &ReadinessAdapter{provider: provider, accountID: accountID, profileID: profileID, contracts: map[protectionreadiness.Market]supervisorContract{}}
	adapter.seal = adapterSeal(adapter)
	return adapter, nil
}

// NewPairedReadinessAdapter seals the read-only KR/US contracts published by
// the production provider. It creates no controller or broker mutation path.
func NewPairedReadinessAdapter(provider protectionreadiness.SnapshotProvider, accountID, profileID string, contracts []protectionreadiness.RuntimeContract) (*ReadinessAdapter, error) {
	if len(contracts) == 0 {
		return NewReadinessAdapter(provider, accountID, profileID)
	}
	sealed := make([]supervisorContract, 0, len(contracts))
	for _, contract := range contracts {
		value, err := newSupervisorContract(contract.Market, contract.SessionScope, contract.TriggerSource, contract.ReplaceSemantics, contract.BrokerCapabilityDigest, contract.ToolDigest)
		if err != nil {
			return nil, err
		}
		sealed = append(sealed, value)
	}
	return newSupervisorReadinessAdapter(provider, accountID, profileID, sealed)
}

func newSupervisorReadinessAdapter(provider protectionreadiness.SnapshotProvider, accountID, profileID string, contracts []supervisorContract) (*ReadinessAdapter, error) {
	adapter, err := NewReadinessAdapter(provider, accountID, profileID)
	if err != nil {
		return nil, err
	}
	for _, contract := range contracts {
		if !validSupervisorContract(contract) || contract.market == "" || adapter.contracts[contract.market].market != "" {
			return nil, errors.New("protection: exact unique supervisor contracts are required")
		}
		adapter.contracts[contract.market] = contract
	}
	if len(adapter.contracts) != 2 || adapter.contracts[protectionreadiness.MarketKR].market == "" || adapter.contracts[protectionreadiness.MarketUS].market == "" {
		return nil, errors.New("protection: paired KR/US supervisor contracts are required")
	}
	adapter.seal = adapterSeal(adapter)
	return adapter, nil
}

func DefaultReadinessAdapter(accountID, profileID string) (*ReadinessAdapter, error) {
	return NewReadinessAdapter(protectionreadiness.DefaultProvider(), accountID, profileID)
}

func (adapter *ReadinessAdapter) Check(ctx context.Context, request ReadinessRequest, now time.Time, previous Checkpoint) (Checkpoint, *ReadinessRefusal) {
	if adapter == nil || adapter.provider == nil || adapter.seal != adapterSeal(adapter) {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalStateCorrupt, Detail: "protection readiness adapter is missing or corrupt"}
	}
	var scopedMarket protectionreadiness.Market
	switch request.Market {
	case "kr":
		scopedMarket = protectionreadiness.MarketKR
	case "us":
		scopedMarket = protectionreadiness.MarketUS
	default:
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalInvalid, Detail: "unsupported protection readiness market"}
	}
	if request.OrderType == "" || request.Quantity == 0 {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalInvalid, Detail: "protection readiness requires an exact order type and integral quantity"}
	}
	snapshot, err := adapter.provider.Current(ctx)
	if err != nil {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalProviderUnavailable, Detail: "protection readiness provider unavailable: " + err.Error()}
	}
	contract := adapter.contracts[scopedMarket]
	decision := snapshot.Dispatch(protectionreadiness.DispatchScope{
		AccountID: adapter.accountID, ProfileID: adapter.profileID, Market: scopedMarket,
		OrderType: request.OrderType, Quantity: request.Quantity,
		SessionScope: contract.sessionScope, TriggerSource: contract.triggerSource,
		ReplaceSemantics: contract.replaceSemantics, BrokerCapabilityDigest: contract.brokerCapabilityDigest,
		ToolDigest: contract.toolDigest,
	}, now)
	if !decision.Allowed {
		return Checkpoint{}, &ReadinessRefusal{Code: decision.Code, Detail: "protection readiness refused for " + request.Market + ": " + string(decision.Code)}
	}
	checkpoint := Checkpoint{market: scopedMarket, generation: decision.Generation, identity: decision.SnapshotID}
	checkpoint.seal = checkpointSeal(checkpoint)
	if previous.Valid() && (previous.market != checkpoint.market || previous.generation != checkpoint.generation || previous.identity != checkpoint.identity) {
		return Checkpoint{}, &ReadinessRefusal{Code: protectionreadiness.RefusalSnapshotDrift, Detail: "protection readiness changed before dispatch"}
	}
	return checkpoint, nil
}

func newSupervisorContract(market protectionreadiness.Market, sessionScope, triggerSource, replaceSemantics, brokerCapabilityDigest, toolDigest string) (supervisorContract, error) {
	contract := supervisorContract{market: market, sessionScope: strings.TrimSpace(sessionScope), triggerSource: strings.TrimSpace(triggerSource), replaceSemantics: strings.TrimSpace(replaceSemantics), brokerCapabilityDigest: strings.TrimSpace(brokerCapabilityDigest), toolDigest: strings.TrimSpace(toolDigest)}
	if market != protectionreadiness.MarketKR && market != protectionreadiness.MarketUS || contract.sessionScope == "" || contract.triggerSource == "" || contract.replaceSemantics == "" || contract.brokerCapabilityDigest == "" || contract.toolDigest == "" {
		return supervisorContract{}, errors.New("protection: invalid supervisor contract")
	}
	contract.seal = supervisorContractSeal(contract)
	return contract, nil
}

func validSupervisorContract(contract supervisorContract) bool {
	return contract.seal != ([32]byte{}) && contract.seal == supervisorContractSeal(contract)
}

func supervisorContractSeal(contract supervisorContract) [32]byte {
	return readinessAdapterHash(string(contract.market), contract.sessionScope, contract.triggerSource, contract.replaceSemantics, contract.brokerCapabilityDigest, contract.toolDigest)
}

func adapterSeal(adapter *ReadinessAdapter) [32]byte {
	if adapter == nil {
		return [32]byte{}
	}
	values := []string{adapter.accountID, adapter.profileID, protectionreadiness.ReadinessRelease}
	for _, market := range []protectionreadiness.Market{protectionreadiness.MarketKR, protectionreadiness.MarketUS} {
		contract := adapter.contracts[market]
		values = append(values, string(market), string(contract.seal[:]))
	}
	return readinessAdapterHash(values...)
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
