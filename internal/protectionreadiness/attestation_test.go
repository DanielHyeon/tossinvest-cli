package protectionreadiness

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

var readinessNow = time.Date(2026, 8, 4, 5, 0, 0, 0, time.UTC)

func TestValidExactAttestationRequiresSealedSupervisorBinding(t *testing.T) {
	fixture := newReadinessFixture(t)
	result := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)}))
	if got := result.Snapshot.Verdict(MarketKR); got.State != Wired || got.Code != RefusalNone || got.Provenance.Serial != 1 {
		t.Fatalf("valid KR verdict=%+v", got)
	}
	if got := result.Snapshot.Verdict(MarketUS); got.State != Unwired || got.Code != RefusalMissingEvidence {
		t.Fatalf("missing US verdict=%+v", got)
	}
	if result.Mutations != 1 || result.ExternalMutations != 0 || !result.StateCommitAllowed || !result.NoLaneAuthority || !result.NoLiveAuthority || !result.PreserveExistingProtection || !result.PreserveReduceOnlyExit {
		t.Fatalf("readiness created authority or removed safety=%+v", result)
	}

	input := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)})
	input.Markets[MarketKR] = withForgedSupervisor(input.Markets[MarketKR])
	result = Assess(input)
	if got := result.Snapshot.Verdict(MarketKR); got.State != Unwired || got.Code != RefusalSupervisorUnwired {
		t.Fatalf("forged supervisor binding accepted=%+v", got)
	}
	input = fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)})
	marketInput := input.Markets[MarketKR]
	marketInput.Supervisor, _ = newSupervisorBinding(supervisorBindingInput{AccountID: "acct", ProfileID: "production", Market: MarketUS, BuildDigest: digestOf("build"), ComponentDigest: digestOf("supervisor-US"), Wired: true})
	input.Markets[MarketKR] = marketInput
	if got := Assess(input).Snapshot.Verdict(MarketKR); got.State != Unwired || got.Code != RefusalSupervisorUnwired {
		t.Fatalf("cross-market sealed supervisor accepted=%+v", got)
	}
}

func TestExactScopeMatrixFailsClosed(t *testing.T) {
	fixture := newReadinessFixture(t)
	tests := []struct {
		name string
		edit func(*runtimeScope)
	}{
		{"account", func(scope *runtimeScope) { scope.AccountID = "other" }},
		{"profile", func(scope *runtimeScope) { scope.ProfileID = "other" }},
		{"market", func(scope *runtimeScope) { scope.Market = MarketUS }},
		{"order type", func(scope *runtimeScope) { scope.OrderType = "TRAILING_STOP" }},
		{"session", func(scope *runtimeScope) { scope.SessionScope = "EXTENDED" }},
		{"quantity min", func(scope *runtimeScope) { scope.QuantityMin = 2 }},
		{"quantity max", func(scope *runtimeScope) { scope.QuantityMax = 101 }},
		{"trigger", func(scope *runtimeScope) { scope.TriggerSource = "MID" }},
		{"replace", func(scope *runtimeScope) { scope.ReplaceSemantics = ReplaceContinuousCoverage }},
		{"tool digest", func(scope *runtimeScope) { scope.ToolDigest = digestOf("other-tool") }},
		{"build digest", func(scope *runtimeScope) { scope.BuildDigest = digestOf("other-build") }},
		{"evidence digest", func(scope *runtimeScope) { scope.EvidenceDigest = digestOf("other-evidence") }},
		{"lookup field", func(scope *runtimeScope) { scope.Broker.ExactLookupField = LookupBrokerOrderID }},
		{"identity scope", func(scope *runtimeScope) { scope.Broker.IdentityUniquenessScope = "OTHER" }},
		{"dedup behavior", func(scope *runtimeScope) { scope.Broker.DuplicateSubmitBehavior = "REJECT" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marketInput := fixture.validMarketInput(t, MarketKR, fixture.krPrivate)
			test.edit(&marketInput.Scope)
			got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: marketInput})).Snapshot.Verdict(MarketKR)
			if got.State != Unwired || got.Code != RefusalScopeMismatch {
				t.Fatalf("scope mismatch accepted=%+v", got)
			}
		})
	}
}

func TestBrokerIdentityCapabilityIsAllOrNothing(t *testing.T) {
	fixture := newReadinessFixture(t)
	tests := []struct {
		name string
		edit func(*brokerCapability)
	}{
		{"client key forward", func(b *brokerCapability) { b.ClientOperationKeyForwarded = false }},
		{"client key echo", func(b *brokerCapability) { b.ClientOperationKeyEchoed = false }},
		{"lookup", func(b *brokerCapability) { b.ExactLookupField = "" }},
		{"identity", func(b *brokerCapability) { b.IdentityUniquenessScope = "" }},
		{"pending query", func(b *brokerCapability) { b.PendingStatusQuery = false }},
		{"terminal query", func(b *brokerCapability) { b.TerminalStatusQuery = false }},
		{"cancel query", func(b *brokerCapability) { b.CancelResultQuery = false }},
		{"dedup", func(b *brokerCapability) { b.DuplicateSubmitBehavior = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := fixture.body(MarketKR)
			test.edit(&body.Broker)
			input := fixture.marketInputForBody(t, body, fixture.krPrivate)
			input.Scope.Broker = body.Broker
			got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input})).Snapshot.Verdict(MarketKR)
			if got.State != Unwired || got.Code != RefusalBrokerCapabilityUnattested {
				t.Fatalf("missing broker capability accepted=%+v", got)
			}
		})
	}
}

func TestStrictJSONAndFileMetadataRefusals(t *testing.T) {
	fixture := newReadinessFixture(t)
	valid := fixture.validMarketInput(t, MarketKR, fixture.krPrivate)
	tests := []struct {
		name string
		edit func(*marketAssessmentInput)
		want RefusalCode
	}{
		{"unknown field", func(input *marketAssessmentInput) {
			input.File.bytes = append(input.File.bytes[:len(input.File.bytes)-1], []byte(`,"legacy":true}`)...)
			resealObservedFile(&input.File)
		}, RefusalUnknownField},
		{"duplicate field", func(input *marketAssessmentInput) {
			input.File.bytes = append([]byte(`{"schema_version":"protection-readiness/v1",`), input.File.bytes[1:]...)
			resealObservedFile(&input.File)
		}, RefusalDuplicateField},
		{"non canonical whitespace", func(input *marketAssessmentInput) {
			input.File.bytes = append([]byte(" "), input.File.bytes...)
			resealObservedFile(&input.File)
		}, RefusalNonCanonical},
		{"owner", func(input *marketAssessmentInput) {
			input.File.OwnerUID++
			input.File.seal = observedFileSeal(input.File)
		}, RefusalFileMetadata},
		{"mode", func(input *marketAssessmentInput) {
			input.File.Mode = 0o644
			input.File.seal = observedFileSeal(input.File)
		}, RefusalFileMetadata},
		{"symlink", func(input *marketAssessmentInput) {
			input.File.Symlink = true
			input.File.seal = observedFileSeal(input.File)
		}, RefusalFileMetadata},
		{"path", func(input *marketAssessmentInput) {
			input.File.ResolvedPath = "/tmp/other.json"
			input.File.seal = observedFileSeal(input.File)
		}, RefusalFileMetadata},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := cloneMarketInput(valid)
			test.edit(&input)
			got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input})).Snapshot.Verdict(MarketKR)
			if got.State != Unwired || got.Code != test.want {
				t.Fatalf("got=%+v want=%s", got, test.want)
			}
		})
	}
}

func TestKeyAlgorithmSignatureRevocationAndRotationPolicy(t *testing.T) {
	fixture := newReadinessFixture(t)
	tests := []struct {
		name string
		edit func(*attestationBody, *marketAssessmentInput)
		key  ed25519.PrivateKey
		want RefusalCode
	}{
		{"schema", func(body *attestationBody, _ *marketAssessmentInput) { body.SchemaVersion = "protection-readiness/v0" }, fixture.krPrivate, RefusalSchema},
		{"algorithm", func(body *attestationBody, _ *marketAssessmentInput) { body.SignatureAlgorithm = "ECDSA" }, fixture.krPrivate, RefusalAlgorithm},
		{"unknown key", func(body *attestationBody, _ *marketAssessmentInput) { body.KeyID = "unknown" }, fixture.krPrivate, RefusalUnknownKey},
		{"revoked", func(body *attestationBody, _ *marketAssessmentInput) { body.KeyID = "revoked" }, fixture.revokedPrivate, RefusalRevokedKey},
		{"overlap ended", func(body *attestationBody, _ *marketAssessmentInput) { body.KeyID = "old" }, fixture.oldPrivate, RefusalRotationWindow},
		{"signature", func(_ *attestationBody, input *marketAssessmentInput) { corruptSignature(&input.File) }, fixture.krPrivate, RefusalSignature},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := fixture.body(MarketKR)
			input := fixture.marketInputForBody(t, body, test.key)
			test.edit(&body, &input)
			if test.name != "signature" {
				input = fixture.marketInputForBody(t, body, test.key)
			}
			got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input})).Snapshot.Verdict(MarketKR)
			if got.State != Unwired || got.Code != test.want {
				t.Fatalf("got=%+v want=%s", got, test.want)
			}
		})
	}
}

func TestRotationOverlapPreservesSerialAcrossKeyIDs(t *testing.T) {
	fixture := newReadinessFixture(t)
	firstBody := fixture.body(MarketKR)
	firstBody.Serial = 10
	first := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.marketInputForBody(t, firstBody, fixture.krPrivate)}))
	if first.Snapshot.Verdict(MarketKR).State != Wired {
		t.Fatalf("initial active key refused=%+v", first.Snapshot.Verdict(MarketKR))
	}

	overlapBody := fixture.body(MarketKR)
	overlapBody.KeyID = "overlap"
	overlapBody.Serial = 9
	rollbackInput := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.marketInputForBody(t, overlapBody, fixture.overlapPrivate)})
	rollbackInput.State = first.NextState
	if got := Assess(rollbackInput).Snapshot.Verdict(MarketKR); got.Code != RefusalSerialRollback || got.State != Unwired {
		t.Fatalf("rotation reset serial accepted=%+v", got)
	}

	overlapBody.Serial = 11
	nextInput := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.marketInputForBody(t, overlapBody, fixture.overlapPrivate)})
	nextInput.State = first.NextState
	if got := Assess(nextInput).Snapshot.Verdict(MarketKR); got.Code != RefusalNone || got.State != Wired || got.Provenance.KeyID != "overlap" {
		t.Fatalf("bounded overlap with increasing serial refused=%+v", got)
	}
}

type readinessFixture struct {
	policy                                                           pinnedTrustPolicy
	state                                                            durableState
	time                                                             trustedTime
	krPrivate, usPrivate, oldPrivate, overlapPrivate, revokedPrivate ed25519.PrivateKey
}

func newReadinessFixture(t *testing.T) readinessFixture {
	t.Helper()
	krPublic, krPrivate := deterministicKey("kr")
	usPublic, usPrivate := deterministicKey("us")
	oldPublic, oldPrivate := deterministicKey("old")
	overlapPublic, overlapPrivate := deterministicKey("overlap")
	revokedPublic, revokedPrivate := deterministicKey("revoked")
	policy, err := newPinnedTrustPolicy(pinnedTrustPolicyInput{
		Release: ReadinessRelease, AllowedAlgorithms: []string{AlgorithmEd25519}, MaximumLifetime: 2 * time.Hour, MaximumRotationOverlap: 2 * time.Hour,
		RequiredOwnerUID: 1000, RequiredMode: 0o600, MaximumFileBytes: 32 * 1024,
		ExpectedPaths: map[Market]string{MarketKR: "/etc/tossos/protection-kr.json", MarketUS: "/etc/tossos/protection-us.json"},
		Keys: []pinnedKeyInput{
			{ID: "kr-active", PublicKey: krPublic, AcceptFrom: readinessNow.Add(-time.Hour), PrimaryUntil: readinessNow.Add(time.Hour), OverlapUntil: readinessNow.Add(2 * time.Hour)},
			{ID: "us-active", PublicKey: usPublic, AcceptFrom: readinessNow.Add(-time.Hour), PrimaryUntil: readinessNow.Add(time.Hour), OverlapUntil: readinessNow.Add(2 * time.Hour)},
			{ID: "old", PublicKey: oldPublic, AcceptFrom: readinessNow.Add(-3 * time.Hour), PrimaryUntil: readinessNow.Add(-2 * time.Hour), OverlapUntil: readinessNow.Add(-time.Minute)},
			{ID: "overlap", PublicKey: overlapPublic, AcceptFrom: readinessNow.Add(-time.Hour), PrimaryUntil: readinessNow.Add(-2 * time.Minute), OverlapUntil: readinessNow.Add(10 * time.Minute)},
			{ID: "revoked", PublicKey: revokedPublic, AcceptFrom: readinessNow.Add(-time.Hour), PrimaryUntil: readinessNow.Add(time.Hour), OverlapUntil: readinessNow.Add(2 * time.Hour), RevokedAt: readinessNow.Add(-time.Minute)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return readinessFixture{policy: policy, state: newDurableState(), time: newTrustedTime(readinessNow, "fixture-clock"), krPrivate: krPrivate, usPrivate: usPrivate, oldPrivate: oldPrivate, overlapPrivate: overlapPrivate, revokedPrivate: revokedPrivate}
}

func deterministicKey(label string) (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := sha256.Sum256([]byte("a071-" + label))
	private := ed25519.NewKeyFromSeed(seed[:])
	return private.Public().(ed25519.PublicKey), private
}

func (fixture readinessFixture) body(market Market) attestationBody {
	keyID := "kr-active"
	if market == MarketUS {
		keyID = "us-active"
	}
	return attestationBody{
		SchemaVersion: SchemaVersionV1, Serial: 1, KeyID: keyID, SignatureAlgorithm: AlgorithmEd25519,
		AccountID: "acct", ProfileID: "production", Market: market, OrderType: "STOP_MARKET", SessionScope: "REGULAR",
		QuantityMin: 1, QuantityMax: 100, TriggerSource: "LAST_TRADE", ReplaceSemantics: ReplaceAtomic,
		Broker: fixtureBrokerCapability(), ToolDigest: digestOf("tool"), BuildDigest: digestOf("build"), EvidenceDigest: digestOf("evidence-" + string(market)),
		IssuedAt: readinessNow.Add(-time.Minute).Format(time.RFC3339Nano), ExpiresAt: readinessNow.Add(time.Hour).Format(time.RFC3339Nano),
	}
}

func fixtureBrokerCapability() brokerCapability {
	return brokerCapability{ClientOperationKeyForwarded: true, ClientOperationKeyEchoed: true, ExactLookupField: LookupClientOperationKey,
		IdentityUniquenessScope: "ACCOUNT_MARKET_OPERATION_KEY", PendingStatusQuery: true, TerminalStatusQuery: true, CancelResultQuery: true,
		DuplicateSubmitBehavior: DuplicateIdempotentSameOrder}
}

func (fixture readinessFixture) validMarketInput(t *testing.T, market Market, private ed25519.PrivateKey) marketAssessmentInput {
	return fixture.marketInputForBody(t, fixture.body(market), private)
}

func (fixture readinessFixture) marketInputForBody(t *testing.T, body attestationBody, private ed25519.PrivateKey) marketAssessmentInput {
	t.Helper()
	canonical, err := canonicalAttestationBody(body)
	if err != nil {
		t.Fatal(err)
	}
	envelope := attestationEnvelope{attestationBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	file, err := newObservedFile(observedFileInput{Bytes: data, ResolvedPath: fixture.policy.expectedPaths[body.Market], OwnerUID: 1000, Mode: 0o600, Regular: true})
	if err != nil {
		t.Fatal(err)
	}
	scope := runtimeScope{AccountID: body.AccountID, ProfileID: body.ProfileID, Market: body.Market, OrderType: body.OrderType, SessionScope: body.SessionScope,
		QuantityMin: body.QuantityMin, QuantityMax: body.QuantityMax, TriggerSource: body.TriggerSource, ReplaceSemantics: body.ReplaceSemantics, Broker: body.Broker,
		ToolDigest: body.ToolDigest, BuildDigest: body.BuildDigest, EvidenceDigest: body.EvidenceDigest}
	supervisor, err := newSupervisorBinding(supervisorBindingInput{AccountID: body.AccountID, ProfileID: body.ProfileID, Market: body.Market, BuildDigest: body.BuildDigest, ComponentDigest: digestOf("supervisor-" + string(body.Market)), Wired: true})
	if err != nil {
		t.Fatal(err)
	}
	return marketAssessmentInput{Scope: scope, File: file, Supervisor: supervisor}
}

func (fixture readinessFixture) input(markets map[Market]marketAssessmentInput) assessmentInput {
	return assessmentInput{Policy: fixture.policy, State: fixture.state, Time: fixture.time, Markets: markets}
}

func withForgedSupervisor(input marketAssessmentInput) marketAssessmentInput {
	input.Supervisor.Wired = true
	input.Supervisor.seal = [32]byte{}
	return input
}

func cloneMarketInput(input marketAssessmentInput) marketAssessmentInput {
	input.File.bytes = append([]byte(nil), input.File.bytes...)
	return input
}

func resealObservedFile(file *observedFile) {
	file.Size = int64(len(file.bytes))
	file.seal = observedFileSeal(*file)
}

func corruptSignature(file *observedFile) {
	var envelope attestationEnvelope
	if err := json.Unmarshal(file.bytes, &envelope); err != nil {
		panic(err)
	}
	signature, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		panic(err)
	}
	signature[0] ^= 1
	envelope.Signature = base64.StdEncoding.EncodeToString(signature)
	file.bytes, err = json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	resealObservedFile(file)
}

func digestOf(value string) string {
	digest := sha256.Sum256([]byte(value))
	return fmtHex(digest[:])
}

func fmtHex(data []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(data)*2)
	for i, value := range data {
		out[i*2], out[i*2+1] = digits[value>>4], digits[value&15]
	}
	return string(out)
}
