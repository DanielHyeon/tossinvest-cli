// Package strategycoordinator 는 한 시장의 제안을 소유자 범위별로 모아
// 범위마다 최대 하나만 고르는 조정자다.
//
// 왜 따로 있나: 지금까지 이 코드는 "시장 하나에 제안 하나"를 전제하고
// 종목 목록을 그 자리에서 훑었다. 전략군이 넷이 되면 같은 종목에 여러
// 제안이 동시에 들어오고, 서로 다른 종목이 서로 다른 속도로 들어온다.
// 그때 필요한 것은 훑는 반복문이 아니라 받는 곳(큐), 접는 규칙(coalescing),
// 정해진 순서(deterministic ordering), 정해진 상한(bounded)이다.
//
// 이 패키지는 아무것도 바꾸지 않는다. 주문도 원장도 토글도 건드리지 않으며,
// 그 사실은 import 목록으로 증명된다(dependency_closure_test.go).
package strategycoordinator

import (
	"sort"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const (
	// LanesPerMarket 는 한 시장이 가지는 레인 수다. 동결 골든의 descriptors 가
	// 시장마다 넷을 못박고 있으며 golden_contract_test.go 가 파일에서 세어 대조한다.
	LanesPerMarket = 4
	// Capacity 는 큐가 담는 레인 칸의 상한이다. 서버가 정하며 호출자가 못 바꾼다.
	//
	// **이 값은 고른 것이 아니라 영수증에서 읽은 것이다.** 큐에 도착할 수 있는
	// 봉투는 그 시장 매니페스트가 실은 (종목, 레인) 범위뿐이고, 그 수의 상한은
	// `strategyproposal.MaxManifestScopes` 가 정한다(validScopes 가 그보다 큰
	// 매니페스트를 거절한다). 그래서 큐는 정확히 그만큼을 담는다.
	//
	// 이보다 작게 잡으면 **검증이 통과시킨 정상 매니페스트가 큐에서 넘쳐** 그
	// 시장을 매 주기 닫는다. 앞선 판본은 이 자리에 64 를 두고 "상류에 상한이
	// 없어서 유도할 영수증이 없다"고 적었지만 그 문장은 사실이 아니었다 —
	// 영수증은 internal/strategyproposal/production.go 에 있었다.
	//
	// 여기서 그 상수를 직접 import 하지 않는 이유: strategyproposal 은 journal 과
	// position 을 끌고 오므로, import 하는 순간 dependency_closure_test 가 지키는
	// "조정자는 아무것도 바꿀 수 없다"는 약속이 깨진다. 대신 두 값이 같은지는
	// receipt_contract_test.go 가 두 패키지를 함께 읽어 확인한다.
	Capacity = 10_000
)

// 아래 상수는 Admission.Detail/Outcome.Detail 에 들어가는 진단이다. 계약이 아니다.
// 중재 쪽 원인은 중재자가 이미 이름을 갖고 있으므로 그 이름을 그대로 쓴다.
const (
	DetailMarketScope      = "the submitted scope is not this coordinator's market"
	DetailNoSnapshot       = "the envelope carries no snapshot digest"
	DetailSnapshotConflict = "the same lane arrived with a different snapshot"
	DetailOverflow         = "the coordinator queue is full"
)

// Key 는 동결 골든 `queue.dedup_key_fields` 가 정한 정확히 여덟 필드의 열쇠다.
// Scope 가 앞 네 개(account, market, symbol, position_generation)를 들고 있다.
type Key struct {
	Scope          strategyrouter.OwnerKey
	Family         strategyrouter.Family
	LaneID         string
	LaneVersion    string
	SnapshotDigest string
}

// DedupKeyFields 는 골든이 적은 여덟 필드 이름을 그 순서대로 돌려준다.
func DedupKeyFields() []string {
	return []string{"account", "market", "symbol", "position_generation",
		"family", "lane_id", "lane_version", "snapshot_digest"}
}

// Parts 는 이 열쇠의 여덟 값이다. DedupKeyFields 와 길이·순서가 언제나 같다.
func (key Key) Parts() []any {
	return []any{key.Scope.AccountRef, key.Scope.Market, key.Scope.Symbol, key.Scope.PositionGeneration,
		key.Family, key.LaneID, key.LaneVersion, key.SnapshotDigest}
}

// laneSlot 은 스냅샷을 뺀 나머지 일곱 필드다. 큐의 한 칸을 가리킨다.
// 같은 레인이 스냅샷만 바꿔 다시 오면 같은 칸을 두고 다투게 되며, 그 다툼은
// 조용히 덮어쓰지 않고 닫는다 — 어느 쪽이 새것인지 정할 봉인된 근거가 없다.
type laneSlot struct {
	Scope       strategyrouter.OwnerKey
	Family      strategyrouter.Family
	LaneID      string
	LaneVersion string
}

func (key Key) laneSlot() laneSlot {
	return laneSlot{Scope: key.Scope, Family: key.Family, LaneID: key.LaneID, LaneVersion: key.LaneVersion}
}

type entry struct {
	key      Key
	identity string
	proposal strategyarbiter.Proposal
}

// Envelope 는 큐에 들어가는 봉투 하나다.
//
// SnapshotDigest 는 이 제안이 재생한 불변 증거 스냅샷의 다이제스트이며
// `strategyproposal.ProductionAuthority.SnapshotDigest()` 가 그 출처다.
// 그 값은 봉인된 계보 안에 들어 있지 않아서 여기로 따로 건네야 한다.
// 대신 조정자는 봉투를 접을 때 계보 신원까지 같은지 본다 — 다이제스트만
// 믿으면 잘못 적힌 다이제스트 하나가 서로 다른 두 제안을 한 칸에 겹쳐 놓는다.
type Envelope struct {
	Scope          strategyrouter.OwnerKey
	SnapshotDigest string
	Proposal       strategyarbiter.Proposal
}

// Selection 은 한 소유자 범위가 고른 제안 하나다. 제안 객체 자체가 아니라
// 봉인된 계보 신원을 건넨다 — 받는 쪽이 자기 목록에서 그 신원으로 찾게 해서,
// 조정자가 남의 자료구조를 들고 다니지 않게 한다.
type Selection struct {
	Scope           strategyrouter.OwnerKey
	Family          strategyrouter.Family
	ScorePPM        uint32
	LineageIdentity string
	ExistingOwner   bool
}

// Admission 은 제안 하나를 큐에 넣으려 한 결과다.
type Admission struct {
	Key      Key
	Admitted bool
	Overflow bool
	Refusal  strategyarbiter.Refusal
	Detail   string
}

// Outcome 은 한 시장의 조정 결과다.
//
// Overflow 를 중재 거절 코드와 따로 두는 이유: 동결 골든의 refusal_enums 에는
// 큐 코드가 없다. 여섯 중재 코드 중 하나를 빌려 쓰면 큐가 넘친 일이 봉인이
// 깨진 일로 보고된다. 이름이 없으면 지어내지 않고 형만 나눈다.
type Outcome struct {
	Selections []Selection
	Overflow   bool
	Refusal    strategyarbiter.Refusal
	Detail     string
	Scope      strategyrouter.OwnerKey
	Drops      uint64
}

// Closed 는 이 시장이 닫혔는지다. 닫혔으면 선택은 하나도 나가지 않는다.
func (outcome Outcome) Closed() bool {
	return outcome.Overflow || outcome.Refusal != strategyarbiter.RefusalNone
}

// MarketCoordinator 는 시장 하나의 조정자다. 골든이 시장마다 하나씩 정확히
// 둘을 요구한다(coordinator_key_fields=[market], coordinator_count=2).
type MarketCoordinator struct {
	market     strategyrouter.Market
	observedAt time.Time
	capacity   int

	mu    sync.Mutex
	slots map[laneSlot]entry
	drops uint64
	// 첫 결함만 남긴다. 뒤에 무엇이 더 오든 이 시장은 이미 닫혔다.
	faulted  bool
	overflow bool
	refusal  strategyarbiter.Refusal
	detail   string
	scope    strategyrouter.OwnerKey
}

// NewMarketCoordinator 는 서버가 정한 상한을 쓰는 생산 조정자를 만든다.
// 그 상한은 이 파일이 선언한 Capacity 하나뿐이다 — 생성자가 다른 수를 들고
// 있으면 Capacity 를 읽고 짠 넘침 시험이 영원히 넘치지 않는다.
func NewMarketCoordinator(market strategyrouter.Market, observedAt time.Time) *MarketCoordinator {
	return newMarketCoordinator(market, observedAt, Capacity)
}

// 미리 잡는 칸 수는 상한이 아니라 실제로 담길 만한 수여야 한다. 상한을 그대로
// 넘기면 아직 한 칸도 안 쓴 조정자가 상한만큼의 버킷을 들고 태어난다.
// 맵은 필요하면 알아서 자란다.
func newMarketCoordinator(market strategyrouter.Market, observedAt time.Time, capacity int) *MarketCoordinator {
	return &MarketCoordinator{market: market, observedAt: observedAt, capacity: capacity,
		slots: make(map[laneSlot]entry, LanesPerMarket)}
}

// Capacity 는 이 조정자가 담을 수 있는 레인 칸 수다.
func (coordinator *MarketCoordinator) Capacity() int { return coordinator.capacity }

// Depth 는 지금 담고 있는 레인 칸 수다. 유계 상태값이다.
func (coordinator *MarketCoordinator) Depth() int {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return len(coordinator.slots)
}

// Drops 는 접히거나 들어가지 못한 봉투의 수다. 유계 계수기다.
func (coordinator *MarketCoordinator) Drops() uint64 {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.drops
}

// Submit 은 봉인된 제안 하나를 큐에 넣는다.
//
// 넣지 못한 이유를 그 자리에서 돌려주고, 같은 이유를 조정자 안에도 남긴다.
// 반환값을 버린 호출자가 거절을 못 본 채 계속 가면 안 되기 때문이다.
func (coordinator *MarketCoordinator) Submit(envelope Envelope) Admission {
	scope, proposal := envelope.Scope, envelope.Proposal
	if scope.Market != coordinator.market {
		return coordinator.fault(scope, false, strategyarbiter.RefusalSealMismatch, DetailMarketScope)
	}
	if scope.AccountRef == "" || scope.Symbol == "" || scope.PositionGeneration == 0 {
		return coordinator.fault(scope, false, strategyarbiter.RefusalSealMismatch, strategyarbiter.DetailInvalidRequest)
	}
	// 생산 매니페스트는 빈 스냅샷 다이제스트를 이미 거절한다. 여기까지 빈 값이
	// 왔다면 배선이 그 값을 옮기지 않은 것이고, 그러면 모든 스냅샷이 한 열쇠로 뭉친다.
	if envelope.SnapshotDigest == "" {
		return coordinator.fault(scope, false, strategyarbiter.RefusalSealMismatch, DetailNoSnapshot)
	}
	// 계보를 열쇠로 쓰기 전에 봉인부터 본다. 봉인이 깨진 계보로 열쇠를 만들면
	// 서로 다른 제안이 한 칸에 겹쳐 하나가 조용히 사라질 수 있다.
	if !proposal.Result.ValidProposal() {
		return coordinator.fault(scope, false, strategyarbiter.RefusalSealMismatch, strategyarbiter.DetailProposalSeal)
	}
	lineage := proposal.Result.Lineage
	if lineage.AccountRef != scope.AccountRef || lineage.Market != scope.Market ||
		lineage.Symbol != scope.Symbol || lineage.PositionGeneration != scope.PositionGeneration {
		return coordinator.fault(scope, false, strategyarbiter.RefusalSealMismatch, strategyarbiter.DetailScope)
	}
	// 가족은 열쇠를 만들기 위해서만 읽는다. 붙지 않으면 빈 값으로 담고
	// 거절은 중재자가 한다 — 같은 판정을 두 곳에 두면 운영자가 보는 진단이 갈린다.
	family := strategyarbiter.ProposalFamily(proposal)
	key := Key{Scope: scope, Family: family, LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
		SnapshotDigest: envelope.SnapshotDigest}

	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	slot := key.laneSlot()
	if existing, ok := coordinator.slots[slot]; ok {
		// 열쇠와 봉인된 계보 신원이 **둘 다** 같아야 같은 봉투다. 다이제스트만
		// 보면 잘못 적힌 다이제스트가 서로 다른 제안을 조용히 덮어쓴다.
		if existing.key != key || existing.identity != lineage.Identity {
			return coordinator.faultLocked(scope, false, strategyarbiter.RefusalSealMismatch, DetailSnapshotConflict)
		}
		// 같은 봉투다. 마지막 하나로 접고 접힌 것을 센다.
		coordinator.slots[slot] = entry{key: key, identity: lineage.Identity, proposal: proposal}
		coordinator.drops++
		return Admission{Key: key, Admitted: true}
	}
	if len(coordinator.slots) >= coordinator.capacity {
		coordinator.drops++
		return coordinator.faultLocked(scope, true, strategyarbiter.RefusalNone, DetailOverflow)
	}
	coordinator.slots[slot] = entry{key: key, identity: lineage.Identity, proposal: proposal}
	return Admission{Key: key, Admitted: true}
}

// Arbitrate 는 담아 둔 제안을 소유자 범위별로 중재해 범위마다 최대 하나를 낸다.
//
// 한 범위가 닫히면 시장 전체를 닫는다. 닫힌 범위만 빼고 나머지를 내보내면
// 아래 파이프라인의 "시장에 하나" 관문이 오히려 만족되어, 막으려던 것과
// 상관없는 다른 범위가 대신 풀린다.
func (coordinator *MarketCoordinator) Arbitrate() Outcome {
	// 잠금을 끝까지 쥔다. 결함을 읽고 나서 놓았다가 다시 잡으면, 그 사이에
	// 결함을 낸 Submit 이 이 호출에 보이지 않는다. 그러면 **이미 닫힌**
	// 조정자가 비어 있지 않은 Selections 를 내놓는다 — fail-closed 로 지은
	// 것 한가운데의 fail-open 구멍이다. 지금 배선은 단일 goroutine 이지만
	// 이 타입은 sync.Mutex 를 걸고 동시 Submit 을 광고한다.
	//
	// 안에서 부르는 strategyarbiter.Arbitrate 는 순수 함수이고 이 패키지를
	// 되부르지 않는다(dependency_closure_test 가 그 방향을 지킨다). I/O 도 없다.
	// 그래서 잠금을 끝까지 쥐어도 막힐 곳이 없다.
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	drops := coordinator.drops
	if coordinator.faulted {
		return Outcome{Overflow: coordinator.overflow, Refusal: coordinator.refusal,
			Detail: coordinator.detail, Scope: coordinator.scope, Drops: drops}
	}
	grouped := make(map[strategyrouter.OwnerKey][]entry, len(coordinator.slots))
	for slot, value := range coordinator.slots {
		grouped[slot.Scope] = append(grouped[slot.Scope], value)
	}

	scopes := make([]strategyrouter.OwnerKey, 0, len(grouped))
	for scope := range grouped {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopeOrderLess(scopes[i], scopes[j]) })

	selections := make([]Selection, 0, len(scopes))
	for _, scope := range scopes {
		entries := grouped[scope]
		// 범위 안의 순서는 도착 순서가 아니라 봉인된 레인 신원으로 정한다.
		// 여러 worker 가 동시에 넣으면 도착 순서는 매번 달라지기 때문이다.
		sort.Slice(entries, func(i, j int) bool { return laneOrderLess(entries[i].key, entries[j].key) })
		proposals := make([]strategyarbiter.Proposal, 0, len(entries))
		for _, value := range entries {
			proposals = append(proposals, value.proposal)
		}
		outcome := strategyarbiter.Arbitrate(strategyarbiter.Request{AccountRef: scope.AccountRef,
			Market: scope.Market, Symbol: scope.Symbol, PositionGeneration: scope.PositionGeneration,
			ObservedAt: coordinator.observedAt, Proposals: proposals})
		if outcome.Refusal != strategyarbiter.RefusalNone {
			return Outcome{Refusal: outcome.Refusal, Detail: outcome.Detail, Scope: scope, Drops: drops}
		}
		selections = append(selections, Selection{Scope: scope, Family: outcome.Family, ScorePPM: outcome.ScorePPM,
			LineageIdentity: outcome.LineageIdentity, ExistingOwner: outcome.ExistingOwner})
	}
	return Outcome{Selections: selections, Drops: drops}
}

// scopeOrderLess 는 소유자 범위의 사전순이다.
func scopeOrderLess(left, right strategyrouter.OwnerKey) bool {
	if left.AccountRef != right.AccountRef {
		return left.AccountRef < right.AccountRef
	}
	if left.Market != right.Market {
		return left.Market < right.Market
	}
	if left.Symbol != right.Symbol {
		return left.Symbol < right.Symbol
	}
	return left.PositionGeneration < right.PositionGeneration
}

// laneOrderLess 는 한 범위 안 레인의 사전순이다.
func laneOrderLess(left, right Key) bool {
	if left.Family != right.Family {
		return left.Family < right.Family
	}
	if left.LaneID != right.LaneID {
		return left.LaneID < right.LaneID
	}
	return left.LaneVersion < right.LaneVersion
}

func (coordinator *MarketCoordinator) fault(scope strategyrouter.OwnerKey, overflow bool,
	code strategyarbiter.Refusal, detail string,
) Admission {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.faultLocked(scope, overflow, code, detail)
}

func (coordinator *MarketCoordinator) faultLocked(scope strategyrouter.OwnerKey, overflow bool,
	code strategyarbiter.Refusal, detail string,
) Admission {
	if !coordinator.faulted {
		coordinator.faulted, coordinator.overflow = true, overflow
		coordinator.refusal, coordinator.detail, coordinator.scope = code, detail, scope
	}
	return Admission{Overflow: overflow, Refusal: code, Detail: detail}
}
