package protectionlifecycle

// a100 T1(tasks 1.1~1.3) — applyFill·prepareRegister 미실행 거부 분기 시험.
//
// 분기 ID 는 a100 번들(analysis/function-logic/internal-protectionlifecycle--applyfill ·
// --prepareregister 의 ast.json, source_sha256 de50441b…)이 정본임.
// 거절 시험은 오류 발생 여부가 아니라 **어느 가드의 어떤 문구인지**를 단언함 —
// 같은 RefusalCode 를 내는 가드가 여럿이라 코드만 보면 앞 가드에 걸린 가짜 GREEN 을 못 거름.
//
// 도달 불가 분기 셋(applyFill B2 · prepareRegister B3 · B4)은 본문을 실행시킬 입력이 없음.
// 그 시나리오는 실제로 막는 앞 가드의 문구로 고정하고, 앞 가드가 그것을 막는 이유
// (validState 진리표)를 구조 시험으로 못 박음. 결함 기록은
// openspec/changes/a100-wire-fill-to-broker-protection/analysis/t1-unreachable-branches.md.

import (
	"testing"
)

// 거절 결과의 실패 지점 단언 — 코드와 전체 문구(=가드 식별)를 함께 비교함.
func requireRefusal(t *testing.T, err error, code RefusalCode, text string) {
	t.Helper()
	if err == nil {
		t.Fatalf("거절 기대 %s: %s, 실제 nil", code, text)
	}
	if errorCode(err) != code {
		t.Fatalf("거절 코드 불일치 want=%s got=%v", code, err)
	}
	if want := string(code) + ": " + text; err.Error() != want {
		t.Fatalf("거절 지점 불일치 want=%q got=%q", want, err.Error())
	}
}

// 호출 전 상태 스냅숏 — 저장 봉인, 내용 해시(stateSeal 재계산), 포지션 수를 값으로 떠 둠.
// State 값 복사는 맵을 공유하므로 호출 뒤의 입력·반환 상태끼리 비교하면 제자리 변경을 못 잡음.
// 반드시 거절 호출 **전에** 떠서 비교해야 "필드 불변" 증명이 됨.
type stateSnapshot struct {
	seal      [32]byte
	content   [32]byte
	positions int
}

func snapshotState(state State) stateSnapshot {
	return stateSnapshot{seal: state.seal, content: stateSeal(state), positions: len(state.positions)}
}

// 거절 후 상태 불변 단언 — 입력 상태(호출자 소유, 맵 공유)와 반환 상태가 모두 호출 전 스냅숏과 같아야 함.
// 내용 해시가 같다는 것은 어떤 필드도 바뀌지 않았다는 뜻이고(stateSeal 은 전 필드 해시),
// 저장 봉인이 같다는 것은 재봉인되지 않았다는 뜻임.
func requireSameState(t *testing.T, before stateSnapshot, input, returned State) {
	t.Helper()
	for _, side := range []struct {
		name  string
		state State
	}{{"입력", input}, {"반환", returned}} {
		if side.state.seal != before.seal {
			t.Fatalf("거절 경로가 %s 상태 봉인을 바꿈 before=%x after=%x", side.name, before.seal, side.state.seal)
		}
		if stateSeal(side.state) != before.content {
			t.Fatalf("거절 경로가 %s 상태 내용을 바꿈(제자리 변경 포함)", side.name)
		}
		if len(side.state.positions) != before.positions {
			t.Fatalf("거절 경로가 %s 상태 포지션 수를 바꿈 before=%d after=%d", side.name, before.positions, len(side.state.positions))
		}
	}
}

// KR 보호가 ACTIVE 인 상태와 그 포지션 키 — registeredKR 의 키 전용 축약.
func registeredKRKey(t *testing.T) (State, PositionKey) {
	t.Helper()
	state, command, _ := registeredKR(t)
	return state, command.Position
}

// KR·US 두 포지션 모두 보호가 ACTIVE 인 상태 구성 — 교차 포지션 귀속 시험용.
func registeredKRAndUS(t *testing.T) (State, PositionKey, PositionKey) {
	t.Helper()
	state, krCommand, _ := registeredKR(t)
	pending, usCommand, err := prepareRegister(state, PositionKey{"acct", "us-pos", MarketUS}, 19, 200, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	active, err := applySubmitResult(pending, usCommand, acceptedObservation(usCommand, "us-broker"))
	if err != nil {
		t.Fatal(err)
	}
	return active, krCommand.Position, usCommand.Position
}

// ---------------------------------------------------------------------------
// 1.1 applyFill
// ---------------------------------------------------------------------------

// 1.1.1 B1 — 알 수 없는 포지션 키의 체결은 상태를 만들지 않고 typed refusal 로 끝남.
// B1 이 없으면 zero position 으로 B3 까지 내려가 "fill identity mismatch" 가 나므로,
// invalid_identity 문구 단언이 B1 을 다른 가드와 가름.
func TestA100T1ApplyFillB1UnknownPositionRefusedWithoutCreatingState(t *testing.T) {
	state, krKey := registeredKRKey(t)
	brokerID := state.view(krKey).Observed.BrokerOrderID
	for _, key := range []PositionKey{
		{"acct", "ghost-pos", MarketKR},    // 존재하지 않는 포지션
		{"other-acct", "kr-pos", MarketKR}, // 같은 포지션 ID, 다른 계좌
		{"acct", "kr-pos", MarketUS},       // 같은 포지션 ID, 다른 시장
	} {
		before := snapshotState(state)
		next, result, err := applyFill(state, key, Fill{FillID: "fill-ghost", BrokerOrderID: brokerID, Quantity: 1, Fingerprint: "trade-ghost"})
		requireRefusal(t, err, RefusalInvalidIdentity, "position missing")
		if result != (FillResult{PreserveExit: true}) {
			t.Fatalf("key=%+v 거절 결과가 청산 보존 외 플래그를 가짐 result=%+v", key, result)
		}
		requireSameState(t, before, state, next)
		if _, exists := next.positions[key]; exists {
			t.Fatalf("key=%+v 알 수 없는 포지션이 생성됨", key)
		}
	}
}

// 1.1.2 B2 — 봉인이 깨진 상태의 체결은 거부되고 어떤 필드도 복구·재봉인되지 않음.
// 실제로 막는 지점은 B1(mutablePosition 의 validState)이며 B2 본문은 도달 불가 —
// 두 지점의 반환값이 바이트 단위로 같아 관측으로 가를 수 없음(결함 기록 참조).
func TestA100T1ApplyFillB2BrokenSealRefusedWithoutReseal(t *testing.T) {
	state, krKey := registeredKRKey(t)
	brokerID := state.view(krKey).Observed.BrokerOrderID
	tampered := cloneState(state)
	position := tampered.positions[krKey]
	position.Holdings++ // 봉인 없이 보유 수량 위조
	tampered.positions[krKey] = position
	if validState(tampered) {
		t.Fatal("시험 전제 붕괴: 위조 상태가 유효로 판정됨")
	}

	before := snapshotState(tampered)
	next, result, err := applyFill(tampered, krKey, Fill{FillID: "fill-1", BrokerOrderID: brokerID, Quantity: 1, Fingerprint: "trade-1"})
	requireRefusal(t, err, RefusalInvalidState, "state seal invalid")
	if result != (FillResult{PreserveExit: true}) {
		t.Fatalf("거절 결과 result=%+v", result)
	}
	requireSameState(t, before, tampered, next) // 재봉인·복구 모두 봉인/내용 해시 변화로 드러남
	if validState(next) {
		t.Fatal("깨진 봉인이 재봉인됨")
	}
	if got := next.positions[krKey].Holdings; got != position.Holdings {
		t.Fatalf("위조 필드가 복구·변경됨 got=%d want=%d", got, position.Holdings)
	}
	if len(next.positions[krKey].Fills) != 0 {
		t.Fatal("깨진 상태에 체결이 기록됨")
	}
}

// 1.1.3 B3 — 다른 포지션의 broker order id 를 실은 체결은 거부됨(남의 포지션 귀속 방어선).
// 식별자 무효·빈 broker id·보호 미등록 포지션도 같은 가드에서 막힘.
// B3 이 없으면 정상 체결로 반영(Applied)되므로 문구 단언이 곧 실패 지점 단언임.
func TestA100T1ApplyFillB3ForeignOrInvalidFillIdentityRefused(t *testing.T) {
	state, krKey, usKey := registeredKRAndUS(t)
	krBroker := state.view(krKey).Observed.BrokerOrderID
	usBroker := state.view(usKey).Observed.BrokerOrderID
	if krBroker == "" || usBroker == "" || krBroker == usBroker {
		t.Fatalf("시험 전제 붕괴 kr=%q us=%q", krBroker, usBroker)
	}
	cases := []struct {
		name string
		key  PositionKey
		fill Fill
	}{
		{"US 주문 체결을 KR 포지션에", krKey, Fill{FillID: "fill-x", BrokerOrderID: usBroker, Quantity: 1, Fingerprint: "trade-x"}},
		{"KR 주문 체결을 US 포지션에", usKey, Fill{FillID: "fill-y", BrokerOrderID: krBroker, Quantity: 1, Fingerprint: "trade-y"}},
		{"빈 broker id", krKey, Fill{FillID: "fill-z", BrokerOrderID: "", Quantity: 1, Fingerprint: "trade-z"}},
		{"빈 FillID", krKey, Fill{FillID: "", BrokerOrderID: krBroker, Quantity: 1, Fingerprint: "trade-z"}},
		{"공백 FillID", krKey, Fill{FillID: " fill ", BrokerOrderID: krBroker, Quantity: 1, Fingerprint: "trade-z"}},
		{"제어문자 fingerprint", krKey, Fill{FillID: "fill-c", BrokerOrderID: krBroker, Quantity: 1, Fingerprint: "trade\n"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotState(state)
			next, result, err := applyFill(state, tc.key, tc.fill)
			requireRefusal(t, err, RefusalInvalidObservation, "fill identity mismatch")
			if result != (FillResult{PreserveExit: true}) {
				t.Fatalf("거절 결과 result=%+v", result)
			}
			requireSameState(t, before, state, next)
			if next.view(krKey).Holdings != 10 || next.view(usKey).Holdings != 20 {
				t.Fatalf("보유 수량 변경 kr=%d us=%d", next.view(krKey).Holdings, next.view(usKey).Holdings)
			}
		})
	}

	// 보호 미등록(UNPROTECTED) 포지션은 broker id 가 비어 있어 어떤 체결도 귀속되지 않음.
	t.Run("보호 미등록 포지션", func(t *testing.T) {
		fresh := baseState(t)
		key := PositionKey{"acct", "kr-pos", MarketKR}
		before := snapshotState(fresh)
		next, _, err := applyFill(fresh, key, Fill{FillID: "fill-u", BrokerOrderID: "", Quantity: 1, Fingerprint: "trade-u"})
		requireRefusal(t, err, RefusalInvalidObservation, "fill identity mismatch")
		requireSameState(t, before, fresh, next)
	})
}

// 1.1.4 B6 — 수량 0 또는 보호 claim(Observed.Quantity) 초과 체결은 거부됨.
// B6 이 없으면 0 은 빈 체결로 반영되고 초과분은 uint64 언더플로로 반영되므로 문구 단언이 실패 지점을 가름.
func TestA100T1ApplyFillB6ZeroOrExcessQuantityRefused(t *testing.T) {
	state, krKey := registeredKRKey(t)
	brokerID := state.view(krKey).Observed.BrokerOrderID
	if view := state.view(krKey); view.Observed.Quantity != 8 || view.Holdings != 10 {
		t.Fatalf("시험 전제 붕괴 view=%+v", view)
	}
	for _, quantity := range []uint64{0, 9, 10, 11} { // 0 · claim 8 초과(보유 이내) · 보유 동일 · 보유 초과
		before := snapshotState(state)
		next, result, err := applyFill(state, krKey, Fill{FillID: "fill-q", BrokerOrderID: brokerID, Quantity: quantity, Fingerprint: "trade-q"})
		requireRefusal(t, err, RefusalFillExceeded, "fill quantity exceeds claim")
		if result != (FillResult{PreserveExit: true}) {
			t.Fatalf("quantity=%d 거절 결과 result=%+v", quantity, result)
		}
		requireSameState(t, before, state, next)
		if _, recorded := next.positions[krKey].Fills["fill-q"]; recorded {
			t.Fatalf("quantity=%d 거절된 체결이 기록됨", quantity)
		}
	}
}

// 1.1.5 B7 — 잔량 0 이 되면 Terminal 로 닫히고, 종료 후 추가 전이는 멱등하게 거부됨.
func TestA100T1ApplyFillB7FullFillClosesTerminalThenRefusesFurtherFills(t *testing.T) {
	for _, steps := range [][]uint64{{8}, {3, 5}} { // 한 번에 전량 · 부분 후 잔량
		state, krKey := registeredKRKey(t)
		brokerID := state.view(krKey).Observed.BrokerOrderID
		var last Fill
		for index, quantity := range steps {
			last = Fill{FillID: "fill-" + string(rune('a'+index)), BrokerOrderID: brokerID, Quantity: quantity, Fingerprint: "trade-" + string(rune('a'+index))}
			next, result, err := applyFill(state, krKey, last)
			if err != nil || !result.Applied || !result.PreserveExit {
				t.Fatalf("steps=%v 체결 반영 실패 result=%+v err=%v", steps, result, err)
			}
			state = next
		}

		view := state.view(krKey)
		position := state.positions[krKey]
		if view.Phase != Terminal || view.Observed.Status != BrokerFilled || view.Desired.Status != BrokerFilled {
			t.Fatalf("steps=%v 전량 체결이 Terminal 로 닫히지 않음 view=%+v", steps, view)
		}
		if view.Observed.Quantity != 0 || view.Desired.Quantity != 0 || view.Holdings != 2 || position.HasPending || view.EntryOpen {
			t.Fatalf("steps=%v 종료 상태 수치 오류 view=%+v pending=%v", steps, view, position.HasPending)
		}
		if !validState(state) {
			t.Fatalf("steps=%v 종료 상태가 진리표를 통과하지 못함", steps)
		}

		// 같은 체결 재전달 → 멱등 Duplicate, 상태 불변.
		before := snapshotState(state)
		again, duplicate, err := applyFill(state, krKey, last)
		if err != nil || !duplicate.Duplicate || duplicate.Applied || !duplicate.PreserveExit {
			t.Fatalf("steps=%v 종료 후 중복 체결 result=%+v err=%v", steps, duplicate, err)
		}
		requireSameState(t, before, state, again)

		// 새 체결 → 잔량 0 이라 B6 에서 거부, 상태 불변.
		before = snapshotState(state)
		extra, result, err := applyFill(state, krKey, Fill{FillID: "fill-after", BrokerOrderID: brokerID, Quantity: 1, Fingerprint: "trade-after"})
		requireRefusal(t, err, RefusalFillExceeded, "fill quantity exceeds claim")
		if result != (FillResult{PreserveExit: true}) {
			t.Fatalf("steps=%v 종료 후 체결 결과 result=%+v", steps, result)
		}
		requireSameState(t, before, state, extra)
	}
}

// ---------------------------------------------------------------------------
// 1.2 prepareRegister
// ---------------------------------------------------------------------------

// 1.2.1 B2 — 진입이 닫혔거나(시장 latch·포지션 latch) phase 가 부적합하면 등록 거부.
// 세 절(시장 latch · 포지션 latch · phase)을 각각 **단독으로** 참이 되는 입력으로 잼 —
// 시장 latch 단독·포지션 latch 단독 사례는 해당 절이 없으면 등록이 성공하므로 절 고유 지점임.
func TestA100T1PrepareRegisterB2ClosedEntryOrWrongPhaseRefused(t *testing.T) {
	krKey := PositionKey{"acct", "kr-pos", MarketKR}

	t.Run("시장 latch(소유자 없는 고아 주문)", func(t *testing.T) {
		state := baseState(t)
		latched, _, err := discoverOrphan(state, BrokerObservation{AccountID: "acct", PositionID: "kr-other", Market: MarketKR, BrokerOrderID: "orphan-1", Status: BrokerActive, Quantity: 1, Trigger: 1})
		requireRefusal(t, err, RefusalUnownedOrphan, "broker order has no exact durable owner")
		if latched.marketEntryOpen(MarketKR) || latched.positions[krKey].EntryLatch != "" || latched.positions[krKey].Phase != Unprotected {
			t.Fatal("시험 전제 붕괴: 시장 latch 단독 사례가 아님")
		}
		before := snapshotState(latched)
		next, command, err := prepareRegister(latched, krKey, 8, 100, fullCapability())
		requireRefusal(t, err, RefusalEntryLatched, "entry is closed")
		if command != (BrokerCommand{}) {
			t.Fatalf("거절 경로가 명령을 만듦 command=%+v", command)
		}
		requireSameState(t, before, latched, next)
	})

	t.Run("포지션 latch 단독(UNPROTECTED + EntryLatch, 재봉인)", func(t *testing.T) {
		// 공개 전이로는 phase 가 UNPROTECTED·TERMINAL 이면서 EntryLatch 만 남는 상태를 B6 이 막는 형태로만
		// 만들 수 있어, 진리표가 허용하는 조합을 직접 심고 재봉인함(validState 통과를 전제로 단언).
		state := baseState(t)
		position := state.positions[krKey]
		position.EntryLatch = RefusalIdempotencyAbsent
		state.positions[krKey] = position
		state.reseal()
		if !validState(state) || !state.marketEntryOpen(MarketKR) || state.positions[krKey].Phase != Unprotected {
			t.Fatal("시험 전제 붕괴: 포지션 latch 단독 사례가 아님")
		}
		before := snapshotState(state)
		next, command, err := prepareRegister(state, krKey, 8, 100, fullCapability())
		requireRefusal(t, err, RefusalEntryLatched, "entry is closed")
		if command != (BrokerCommand{}) {
			t.Fatalf("거절 경로가 명령을 만듦 command=%+v", command)
		}
		requireSameState(t, before, state, next)
	})

	t.Run("포지션 latch + phase(충돌 체결 → RECONCILE_REQUIRED)", func(t *testing.T) {
		state, _, _ := registeredKR(t)
		brokerID := state.view(krKey).Observed.BrokerOrderID
		state, _, _ = applyFill(state, krKey, Fill{FillID: "fill-1", BrokerOrderID: brokerID, Quantity: 1, Fingerprint: "trade-1"})
		latched, _, err := applyFill(state, krKey, Fill{FillID: "fill-1", BrokerOrderID: brokerID, Quantity: 2, Fingerprint: "trade-1"})
		requireRefusal(t, err, RefusalConflictingFill, "fill ID reused with different content")
		if latched.positions[krKey].EntryLatch != RefusalConflictingFill || latched.positions[krKey].Phase != ReconcileRequired {
			t.Fatal("시험 전제 붕괴: 포지션 latch 가 아님")
		}
		before := snapshotState(latched)
		next, command, err := prepareRegister(latched, krKey, 7, 100, fullCapability())
		requireRefusal(t, err, RefusalEntryLatched, "entry is closed")
		if command != (BrokerCommand{}) {
			t.Fatalf("거절 경로가 명령을 만듦 command=%+v", command)
		}
		requireSameState(t, before, latched, next)
	})
}

// 1.2.2 B3 시나리오 — 이미 pending 인 포지션의 재등록은 두 번째 제출을 만들지 않음.
// 실제로 막는 지점은 B2("entry is closed")이며 B3 본문은 도달 불가(결함 기록 참조).
// SUBMIT_PENDING·SUBMIT_UNKNOWN·REPLACE_PENDING·CANCEL_PENDING 전부를 확인함.
func TestA100T1PrepareRegisterB3PendingOperationNeverSubmitsTwice(t *testing.T) {
	krKey := PositionKey{"acct", "kr-pos", MarketKR}
	submitPending, first, err := prepareRegister(baseState(t), krKey, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	submitUnknown, err := applySubmitResult(submitPending, first, unknownObservation(first))
	if err != nil {
		t.Fatal(err)
	}
	active, _, _ := registeredKR(t)
	replacePending, _, err := prepareReplace(active, krKey, 8, 101, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	cancelPending, _, err := prepareCancel(active, krKey, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	for name, state := range map[string]State{"SUBMIT_PENDING": submitPending, "SUBMIT_UNKNOWN": submitUnknown, "REPLACE_PENDING": replacePending, "CANCEL_PENDING": cancelPending} {
		t.Run(name, func(t *testing.T) {
			if !state.positions[krKey].HasPending {
				t.Fatal("시험 전제 붕괴: pending 이 아님")
			}
			before := snapshotState(state)
			next, command, err := prepareRegister(state, krKey, 8, 100, fullCapability())
			requireRefusal(t, err, RefusalEntryLatched, "entry is closed")
			if command != (BrokerCommand{}) {
				t.Fatalf("두 번째 제출 명령이 만들어짐 command=%+v", command)
			}
			requireSameState(t, before, state, next)
			if next.positions[krKey].Pending != state.positions[krKey].Pending {
				t.Fatal("기존 pending 명령이 바뀜")
			}
		})
	}
}

// 1.2.3 B4 시나리오 — 보호가 이미 ACTIVE 인 포지션에 두 번째 보호주문이 나가지 않음.
// 실제로 막는 지점은 B2(phase ACTIVE)이며 B4 본문은 도달 불가(결함 기록 참조).
func TestA100T1PrepareRegisterB4ActiveProtectionNeverSubmitsSecond(t *testing.T) {
	state, command, broker := registeredKR(t)
	if view := state.view(command.Position); view.Observed.Status != BrokerActive || view.Phase != Active || !view.EntryOpen {
		t.Fatalf("시험 전제 붕괴 view=%+v", view)
	}
	before := snapshotState(state)
	next, second, err := prepareRegister(state, command.Position, 8, 100, fullCapability())
	requireRefusal(t, err, RefusalEntryLatched, "entry is closed")
	if second != (BrokerCommand{}) || broker.submitCount != 1 {
		t.Fatalf("두 번째 보호주문 command=%+v submitCount=%d", second, broker.submitCount)
	}
	requireSameState(t, before, state, next)
}

// 1.2.4 B5 — 브로커가 정확한 operation 조회를 못 하면(capability 부재) 보호를 시도하지 않음.
// 같은 RefusalInvalidObservation 을 내는 B1(capability 봉인 위조)과 문구로 가름.
func TestA100T1PrepareRegisterB5MissingExactOperationLookupRefused(t *testing.T) {
	krKey := PositionKey{"acct", "kr-pos", MarketKR}
	state := baseState(t)

	attestedAbsent := newBrokerCapability(false, true, true, true, true) // 봉인은 유효, 능력만 부재
	before := snapshotState(state)
	next, command, err := prepareRegister(state, krKey, 8, 100, attestedAbsent)
	requireRefusal(t, err, RefusalInvalidObservation, "exact operation lookup unavailable")
	if command != (BrokerCommand{}) {
		t.Fatalf("능력 부재인데 명령이 만들어짐 command=%+v", command)
	}
	requireSameState(t, before, state, next)

	// 대조: 봉인 뒤에 능력을 꺼 위조하면 B1 의 봉인 검사에서 먼저 막힘(코드는 같고 문구가 다름).
	forged := fullCapability()
	forged.exactOperationLookup = false
	_, _, err = prepareRegister(state, krKey, 8, 100, forged)
	requireRefusal(t, err, RefusalInvalidObservation, "capability seal invalid")
}

// ---------------------------------------------------------------------------
// 도달 불가 구조 고정
// ---------------------------------------------------------------------------

// prepareRegister B3·B4 가 도달 불가인 이유를 못 박음: B2 를 통과하는 phase(UNPROTECTED·TERMINAL)는
// validState 진리표가 HasPending=true 나 Observed=ACTIVE 와의 조합을 거부하므로, 재봉인해도
// B1(mutablePosition)에서 invalid_state 로 끝남. 이 시험이 깨지면 B3·B4 가 다시 도달 가능해진 것이니
// branch-test-map 과 결함 기록을 재측정해야 함.
func TestA100T1UnreachablePrepareRegisterB3B4AreShadowedByStateTruthTable(t *testing.T) {
	// 두 시장 모두 잼 — 진리표가 한 시장만 건너뛰는 완화도 잡기 위함.
	active, krKey, usKey := registeredKRAndUS(t)
	terminal := active
	for _, key := range []PositionKey{krKey, usKey} {
		view := terminal.view(key)
		next, _, err := applyFill(terminal, key, Fill{FillID: "fill-all-" + string(key.Market), BrokerOrderID: view.Observed.BrokerOrderID, Quantity: view.Observed.Quantity, Fingerprint: "trade-all"})
		if err != nil || next.view(key).Phase != Terminal {
			t.Fatalf("시험 전제 붕괴 key=%+v phase=%s err=%v", key, next.view(key).Phase, err)
		}
		terminal = next
	}
	quantities := map[PositionKey]uint64{krKey: 8, usKey: 19} // 각 포지션의 보유-other 와 정확히 같은 보호 수량
	for _, base := range []struct {
		name  string
		state State
	}{{"UNPROTECTED", baseState(t)}, {"TERMINAL", terminal}} {
		for _, key := range []PositionKey{krKey, usKey} {
			for _, forge := range []struct {
				name  string
				apply func(*positionState)
			}{
				{"HasPending=true(B3 조건)", func(p *positionState) { p.HasPending = true }},
				{"Observed=ACTIVE(B4 조건)", func(p *positionState) { p.Observed.Status = BrokerActive }},
			} {
				t.Run(base.name+"/"+string(key.Market)+"/"+forge.name, func(t *testing.T) {
					if base.state.positions[key].Phase != Phase(base.name) {
						t.Fatalf("시험 전제 붕괴 phase=%s", base.state.positions[key].Phase)
					}
					forged := cloneState(base.state)
					position := forged.positions[key]
					forge.apply(&position)
					forged.positions[key] = position
					forged.reseal() // 봉인까지 맞춰도 진리표가 거부해야 함
					if validState(forged) {
						t.Fatal("진리표가 B3/B4 도달 조합을 허용함 — B3·B4 재측정 필요")
					}
					_, command, err := prepareRegister(forged, key, quantities[key], 100, fullCapability())
					requireRefusal(t, err, RefusalInvalidState, "state seal invalid")
					if command != (BrokerCommand{}) {
						t.Fatalf("command=%+v", command)
					}
				})
			}
		}
	}
}

// applyFill B2 가 도달 불가인 이유를 못 박음: B1 의 mutablePosition 은 첫 줄에서 같은 validState 를
// 부르고, applyFill 이 넘기는 zero capability(봉인 0)는 봉인 검사를 건너뛰므로 mutablePosition 이
// RefusalInvalidObservation 을 낼 길이 없음 — 즉 B1 의 예외절(!= InvalidObservation)이 거짓이 되는
// 오류가 없어 무효 봉인은 언제나 B1 에서 끝남.
func TestA100T1UnreachableApplyFillB2IsShadowedByB1(t *testing.T) {
	state, krKey := registeredKRKey(t)
	tampered := cloneState(state)
	position := tampered.positions[krKey]
	position.Observed.Quantity++
	tampered.positions[krKey] = position
	if _, err := mutablePosition(tampered, krKey, brokerCapability{}); errorCode(err) != RefusalInvalidState {
		t.Fatalf("무효 봉인이 mutablePosition 에서 invalid_state 로 끝나지 않음 err=%v", err)
	}
	for _, key := range []PositionKey{krKey, {"acct", "ghost", MarketKR}, {"", "", "XX"}} {
		if _, err := mutablePosition(state, key, brokerCapability{}); err != nil && errorCode(err) == RefusalInvalidObservation {
			t.Fatalf("zero capability 로 InvalidObservation 이 나옴 — applyFill B1 예외절·B2 재측정 필요 err=%v", err)
		}
	}
}
