package protectionlifecycle

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type OperationKind string

const (
	OperationSubmit  OperationKind = "SUBMIT"
	OperationReplace OperationKind = "REPLACE"
	OperationCancel  OperationKind = "CANCEL"
)

type Phase string

const (
	Unprotected       Phase = "UNPROTECTED"
	SubmitPending     Phase = "SUBMIT_PENDING"
	SubmitUnknown     Phase = "SUBMIT_UNKNOWN"
	Active            Phase = "ACTIVE"
	ReplacePending    Phase = "REPLACE_PENDING"
	ReplaceUnknown    Phase = "REPLACE_UNKNOWN"
	CancelPending     Phase = "CANCEL_PENDING"
	CancelUnknown     Phase = "CANCEL_UNKNOWN"
	ReconcileRequired Phase = "RECONCILE_REQUIRED"
	Terminal          Phase = "TERMINAL"
)

type BrokerStatus string

const (
	BrokerNotFound BrokerStatus = "NOT_FOUND"
	BrokerActive   BrokerStatus = "ACTIVE"
	BrokerCanceled BrokerStatus = "CANCELED"
	BrokerFilled   BrokerStatus = "FILLED"
	BrokerUnknown  BrokerStatus = "UNKNOWN"
)

type RefusalCode string

const (
	RefusalInvalidState       RefusalCode = "invalid_state"
	RefusalInvalidIdentity    RefusalCode = "invalid_identity"
	RefusalInvalidObservation RefusalCode = "invalid_exact_observation"
	RefusalEntryLatched       RefusalCode = "entry_latched"
	RefusalOperationPending   RefusalCode = "operation_pending"
	RefusalNotActive          RefusalCode = "protection_not_active"
	RefusalIdempotencyAbsent  RefusalCode = "broker_idempotency_unattested"
	RefusalContinuousCoverage RefusalCode = "continuous_coverage_unattested"
	RefusalTriggerRetreat     RefusalCode = "trigger_retreat"
	RefusalSellClaimExceeded  RefusalCode = "sell_claim_exceeded"
	RefusalConflictingFill    RefusalCode = "conflicting_duplicate_fill"
	RefusalFillExceeded       RefusalCode = "fill_exceeded"
	RefusalUnownedOrphan      RefusalCode = "unowned_orphan"
	RefusalOrphanConflict     RefusalCode = "orphan_observation_conflict"
)

type lifecycleError struct {
	code RefusalCode
	text string
}

func (err *lifecycleError) Error() string { return fmt.Sprintf("%s: %s", err.code, err.text) }

func refuse(code RefusalCode, format string, args ...any) error {
	return &lifecycleError{code: code, text: fmt.Sprintf(format, args...)}
}

func errorCode(err error) RefusalCode {
	if typed, ok := err.(*lifecycleError); ok {
		return typed.code
	}
	return ""
}

type PositionKey struct {
	AccountID  string
	PositionID string
	Market     Market
}

type PositionSeed struct {
	AccountID       string
	PositionID      string
	Market          Market
	Generation      uint64
	Holdings        uint64
	OtherSellClaims uint64
}

type ProtectionOrder struct {
	Generation    uint64
	Revision      uint64
	OperationKey  string
	BrokerOrderID string
	Status        BrokerStatus
	Quantity      uint64
	Trigger       uint64
}

type BrokerCommand struct {
	Kind          OperationKind
	Position      PositionKey
	Generation    uint64
	Revision      uint64
	OperationKey  string
	BrokerOrderID string
	Quantity      uint64
	Trigger       uint64
}

type BrokerObservation struct {
	AccountID     string
	PositionID    string
	Market        Market
	Generation    uint64
	Revision      uint64
	OperationKey  string
	BrokerOrderID string
	Status        BrokerStatus
	Quantity      uint64
	Trigger       uint64
}

type Fill struct {
	FillID        string
	BrokerOrderID string
	Quantity      uint64
	Fingerprint   string
}

type FillResult struct {
	Applied      bool
	Duplicate    bool
	PreserveExit bool
}

type PositionView struct {
	Key                PositionKey
	Generation         uint64
	ProtectionRevision uint64
	LifecycleRevision  uint64
	Holdings           uint64
	OtherSellClaims    uint64
	Phase              Phase
	EntryOpen          bool
	Desired            ProtectionOrder
	Observed           ProtectionOrder
}

type brokerCapability struct {
	exactOperationLookup bool
	exactBrokerIDLookup  bool
	cancelResultQuery    bool
	continuousReplace    bool
	idempotentSubmit     bool
	seal                 [32]byte
}

func newBrokerCapability(exactOperation, exactBrokerID, cancelQuery, continuousReplace, idempotent bool) brokerCapability {
	capability := brokerCapability{
		exactOperationLookup: exactOperation,
		exactBrokerIDLookup:  exactBrokerID,
		cancelResultQuery:    cancelQuery,
		continuousReplace:    continuousReplace,
		idempotentSubmit:     idempotent,
	}
	capability.seal = capabilitySeal(capability)
	return capability
}

func capabilitySeal(capability brokerCapability) [32]byte {
	return hashParts(
		fmt.Sprint(capability.exactOperationLookup), fmt.Sprint(capability.exactBrokerIDLookup),
		fmt.Sprint(capability.cancelResultQuery), fmt.Sprint(capability.continuousReplace),
		fmt.Sprint(capability.idempotentSubmit),
	)
}

func validCapability(capability brokerCapability) bool {
	return capability.seal == capabilitySeal(capability)
}

func validIdentity(value string) bool {
	if value == "" || len(value) > 256 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validMarket(market Market) bool { return market == MarketKR || market == MarketUS }

func operationIdentity(key PositionKey, generation, revision uint64, kind OperationKind) string {
	digest := hashParts(key.AccountID, key.PositionID, string(key.Market), fmt.Sprint(generation), fmt.Sprint(revision), string(kind))
	return hex.EncodeToString(digest[:])
}

func hashParts(parts ...string) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(part))
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return digest
}
