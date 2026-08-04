package protectionreadiness

import "time"

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

func validMarket(market Market) bool { return market == MarketKR || market == MarketUS }

type Readiness string

const (
	Wired   Readiness = "WIRED"
	Unwired Readiness = "UNWIRED"
)

type RefusalCode string

const (
	RefusalNone                       RefusalCode = ""
	RefusalMissingEvidence            RefusalCode = "missing_evidence"
	RefusalInvalid                    RefusalCode = "invalid_attestation"
	RefusalFileMetadata               RefusalCode = "attestation_file_metadata"
	RefusalDuplicateField             RefusalCode = "attestation_duplicate_field"
	RefusalUnknownField               RefusalCode = "attestation_unknown_field"
	RefusalNonCanonical               RefusalCode = "attestation_non_canonical"
	RefusalSchema                     RefusalCode = "attestation_schema"
	RefusalAlgorithm                  RefusalCode = "attestation_algorithm"
	RefusalUnknownKey                 RefusalCode = "attestation_unknown_key"
	RefusalRevokedKey                 RefusalCode = "attestation_key_revoked"
	RefusalRotationWindow             RefusalCode = "attestation_rotation_window"
	RefusalSignature                  RefusalCode = "attestation_signature"
	RefusalSerialRollback             RefusalCode = "attestation_serial_rollback"
	RefusalTrustedTimeUnavailable     RefusalCode = "trusted_time_unavailable"
	RefusalTrustedTimeRollback        RefusalCode = "trusted_time_rollback"
	RefusalMaximumLifetime            RefusalCode = "attestation_maximum_lifetime"
	RefusalIssuedInFuture             RefusalCode = "attestation_issued_in_future"
	RefusalExpired                    RefusalCode = "attestation_expired"
	RefusalScopeMismatch              RefusalCode = "attestation_scope_mismatch"
	RefusalBrokerCapabilityUnattested RefusalCode = "broker_identity_capability_unattested"
	RefusalSupervisorUnwired          RefusalCode = "protection_supervisor_unwired"
	RefusalStateCorrupt               RefusalCode = "attestation_state_corrupt"
	RefusalProviderUnavailable        RefusalCode = "readiness_provider_unavailable"
	RefusalSnapshotDrift              RefusalCode = "readiness_snapshot_drift"
)

const (
	ReadinessRelease             = "a071-kr-us-protection-readiness-v1"
	SchemaVersionV1              = "protection-readiness/v1"
	AlgorithmEd25519             = "Ed25519"
	ReplaceAtomic                = "ATOMIC"
	ReplaceContinuousCoverage    = "CONTINUOUS_COVERAGE"
	LookupClientOperationKey     = "CLIENT_OPERATION_KEY"
	LookupBrokerOrderID          = "BROKER_ORDER_ID"
	DuplicateIdempotentSameOrder = "IDEMPOTENT_SAME_ORDER"
)

type brokerCapability struct {
	ClientOperationKeyForwarded bool   `json:"client_operation_key_forwarded"`
	ClientOperationKeyEchoed    bool   `json:"client_operation_key_echoed"`
	ExactLookupField            string `json:"exact_lookup_field"`
	IdentityUniquenessScope     string `json:"identity_uniqueness_scope"`
	PendingStatusQuery          bool   `json:"pending_status_query"`
	TerminalStatusQuery         bool   `json:"terminal_status_query"`
	CancelResultQuery           bool   `json:"cancel_result_query"`
	DuplicateSubmitBehavior     string `json:"duplicate_submit_behavior"`
}

func validBrokerCapability(capability brokerCapability) bool {
	return capability.ClientOperationKeyForwarded && capability.ClientOperationKeyEchoed &&
		(capability.ExactLookupField == LookupClientOperationKey || capability.ExactLookupField == LookupBrokerOrderID) &&
		capability.IdentityUniquenessScope == "ACCOUNT_MARKET_OPERATION_KEY" && capability.PendingStatusQuery &&
		capability.TerminalStatusQuery && capability.CancelResultQuery && capability.DuplicateSubmitBehavior == DuplicateIdempotentSameOrder
}

type runtimeScope struct {
	AccountID        string
	ProfileID        string
	Market           Market
	OrderType        string
	SessionScope     string
	QuantityMin      uint64
	QuantityMax      uint64
	TriggerSource    string
	ReplaceSemantics string
	Broker           brokerCapability
	ToolDigest       string
	BuildDigest      string
	EvidenceDigest   string
}

type Provenance struct {
	AccountID              string
	ProfileID              string
	OrderType              string
	SessionScope           string
	QuantityMin            uint64
	QuantityMax            uint64
	TriggerSource          string
	ReplaceSemantics       string
	BrokerCapabilityDigest string
	ToolDigest             string
	KeyID                  string
	Serial                 uint64
	BodyDigest             string
	BuildDigest            string
	EvidenceDigest         string
	SupervisorDigest       string
	IssuedAt               time.Time
	ExpiresAt              time.Time
}

func brokerCapabilityDigest(capability brokerCapability) string {
	digest := hashStrings(
		boolString(capability.ClientOperationKeyForwarded),
		boolString(capability.ClientOperationKeyEchoed),
		capability.ExactLookupField,
		capability.IdentityUniquenessScope,
		boolString(capability.PendingStatusQuery),
		boolString(capability.TerminalStatusQuery),
		boolString(capability.CancelResultQuery),
		capability.DuplicateSubmitBehavior,
	)
	return hexBytes(digest[:])
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

type Verdict struct {
	Market     Market
	State      Readiness
	Code       RefusalCode
	Provenance Provenance
}
