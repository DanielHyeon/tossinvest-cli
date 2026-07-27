// Package position is the position projection state machine (change
// add-core-domain, task 6.1; position-ledger "Position 투영과 단일 권위",
// "포지션 상태기계").
//
// # What this package is, and what it deliberately is not
//
// A Position is a *projection*: the only things that move it are fill events
// and adjustment events, and it exposes no mutation API of its own (SHALL NOT).
// This package owns the rule that turns one such event into the next position
// state; it owns no storage, no clock, no broker client and no transaction. It
// is a pure function, so the transition table below can be tested row by row
// without a database, and so the same rule is applied identically by the fill
// apply hook (internal/journal, inside the fill transaction) and by the
// adjustment path.
//
// It is not the journal's state machine. The journal owns the atomic apply
// point and refuses to own the domain rule (internal/journal/apply_hook.go);
// this package owns the domain rule and refuses to own the storage. Neither
// half can be changed without the other's tests noticing.
//
// # Direction comes from the intent, never from the fill
//
// A broker fill carries an unsigned cumulative quantity and no side. The
// direction is re-derived from the intent behind the order (SHALL —
// position-ledger). Every local fill in this change's scope comes from an order
// that has an intent: the issuer records the intent before the Gateway is
// called, and there are no broker-resident protection orders yet. The direction
// source for a *triggered* order is defined by the change that introduces them
// (2c), not here, and this package has no code for it.
//
// Long-only, so ENTRY is a BUY and EXIT is a SELL (risk-management: long-only).
//
// # The four inputs of a transition
//
// A transition is `(현재 상태, 주문 역할, 누적 delta, lineage) → 다음 상태`,
// determinstically and with no arithmetic correction anywhere (SHALL). The four
// are spelled here as:
//
//	state        the instance's current State
//	role         ENTRY (buy) | EXIT (sell), from the intent's side
//	movement     what this observation's delta does to the instance quantity
//	disposition  what the *order* can still do, which is where lineage enters
//
// Movement is the cumulative delta's effect, not the delta itself, because the
// same delta means different things against different quantities:
//
//	NONE        Δ = 0 — a correction, or a terminal transition that filled
//	            nothing. Both reach the hook (issues.md, task 0.3) and neither
//	            moves quantity.
//	ADDS        Δ > 0 on an ENTRY
//	REDUCES     0 < Δ < quantity on an EXIT
//	FLATTENS    Δ = quantity on an EXIT
//	OVERSHOOTS  Δ > quantity on an EXIT — the sell cannot be attributed to what
//	            is held
//
// Disposition is what the order will still do, and it is the only place lineage
// is consulted:
//
//	WORKING    the order can still fill: not terminal, and its cumulative fill
//	           has not reached the quantity it was placed for
//	DONE       the order will not fill again and nothing carries its remainder:
//	           either it reached 원주문 수량, or it went terminal with no
//	           replace edge leaving it
//	SUCCEEDED  a replace edge leaves the order (lineage 승계): the remainder is
//	           carried by the child, so the *job* the order was doing is still
//	           in flight even though the order is over
//
// SUCCEEDED is what stops an amendment reading as "the entry ended early, then
// a stranger bought the same symbol". The official API answers a modify with a
// new order number (internal/journal/lineage.go), so without the edge the
// child's fills would look like a scale-in and the parent's cancellation would
// look like a completed entry.
//
// # 원주문 수량 is judged per order, and lineage composes it
//
// "OPENING 종료 판단(원주문 수량)" is evaluated against the *filling order's own*
// ordered quantity. For an order that was never amended that is literally the
// original request. For an amended one, the parent goes SUCCEEDED at whatever it
// filled and the child's ordered quantity is the remainder, so the chain
// completes exactly when the original request has been filled. Judging against
// a remembered "original" instead would need the projection to store an order
// identity the schema does not carry (design D7's positions table has no
// tracked-order column), and would be wrong the moment an amend changed the
// quantity.
//
// # The transition table
//
// Every row is enumerated in table.go as data, and this comment is the same
// table in prose order. 96 rows: 6 states × 2 roles × the movements each role
// can produce × 3 dispositions. FLAT and CLOSED hold quantity 0, so REDUCES and
// FLATTENS cannot be produced against them (both need 0 < Δ ≤ quantity) — those
// 12 combinations are structurally unreachable rather than omitted, and
// TestUnreachableMovementsCannotBeConstructed proves it.
//
// A row whose next state is RECONCILE is a *refusal*: the instance's quantity
// and cost basis are left exactly as they were (산식 보정 금지 — no arithmetic
// correction) and the caller enters the durable RECONCILE state. The refusal is
// a value, never a panic: a live account's fill loop must not be stopped by a
// state it did not expect, and the fill itself is still recorded, so nothing the
// broker told us is lost.
//
// ## ENTRY, Δ = 0 (NONE) — a correction, or an order ending with nothing filled
//
//	row  state    disposition  next     note
//	E01  FLAT     WORKING      FLAT     no instance exists to move
//	E02  FLAT     DONE         FLAT     an order that ended having filled nothing
//	E03  FLAT     SUCCEEDED    FLAT
//	E04  OPENING  WORKING      OPENING  cost basis may move, quantity may not
//	E05  OPENING  DONE         OPEN     the entry order is over: 진입 완료
//	E06  OPENING  SUCCEEDED    OPENING  lineage 승계 — the child carries the rest
//	E07  OPEN     WORKING      SCALING  an entry order is in flight again
//	E08  OPEN     DONE         OPEN
//	E09  OPEN     SUCCEEDED    SCALING
//	E10  SCALING  WORKING      SCALING
//	E11  SCALING  DONE         OPEN     SCALING 종료
//	E12  SCALING  SUCCEEDED    SCALING
//	E13  CLOSING  WORKING      CLOSING  nothing moved, so nothing is refused
//	E14  CLOSING  DONE         CLOSING
//	E15  CLOSING  SUCCEEDED    CLOSING
//	E16  CLOSED   WORKING      CLOSED   CLOSED 종결성
//	E17  CLOSED   DONE         CLOSED
//	E18  CLOSED   SUCCEEDED    CLOSED
//
// ## ENTRY, Δ > 0 (ADDS)
//
//	row  state    disposition  next     note
//	E19  FLAT     WORKING      OPENING  new instance
//	E20  FLAT     DONE         OPEN     즉시 전량체결 — FLAT→OPEN 직행, new instance
//	E21  FLAT     SUCCEEDED    OPENING  new instance, amended before it completed
//	E22  OPENING  WORKING      OPENING
//	E23  OPENING  DONE         OPEN     원주문 수량 도달
//	E24  OPENING  SUCCEEDED    OPENING
//	E25  OPEN     WORKING      SCALING  SCALING 진입
//	E26  OPEN     DONE         OPEN     a scale-in that completed in one observation
//	E27  OPEN     SUCCEEDED    SCALING
//	E28  SCALING  WORKING      SCALING
//	E29  SCALING  DONE         OPEN     SCALING 종료
//	E30  SCALING  SUCCEEDED    SCALING
//	E31  CLOSING  WORKING      RECONCILE  ENTRY_WHILE_CLOSING
//	E32  CLOSING  DONE         RECONCILE  ENTRY_WHILE_CLOSING
//	E33  CLOSING  SUCCEEDED    RECONCILE  ENTRY_WHILE_CLOSING
//	E34  CLOSED   WORKING      OPENING  재진입 = new instance
//	E35  CLOSED   DONE         OPEN     재진입, 즉시 전량체결, new instance
//	E36  CLOSED   SUCCEEDED    OPENING  new instance
//
// ## EXIT, Δ = 0 (NONE)
//
//	row  state    disposition  next     note
//	X01  FLAT     WORKING      FLAT
//	X02  FLAT     DONE         FLAT
//	X03  FLAT     SUCCEEDED    FLAT
//	X04  OPENING  WORKING      CLOSING  an exit order is in flight
//	X05  OPENING  DONE         OPENING  it ended without filling: entry still working
//	X06  OPENING  SUCCEEDED    CLOSING
//	X07  OPEN     WORKING      CLOSING
//	X08  OPEN     DONE         OPEN
//	X09  OPEN     SUCCEEDED    CLOSING
//	X10  SCALING  WORKING      CLOSING
//	X11  SCALING  DONE         SCALING  the entry order is still working
//	X12  SCALING  SUCCEEDED    CLOSING
//	X13  CLOSING  WORKING      CLOSING
//	X14  CLOSING  DONE         OPEN     no exit order left in flight
//	X15  CLOSING  SUCCEEDED    CLOSING
//	X16  CLOSED   WORKING      CLOSED   CLOSED 종결성
//	X17  CLOSED   DONE         CLOSED
//	X18  CLOSED   SUCCEEDED    CLOSED
//
// ## EXIT, 0 < Δ < quantity (REDUCES) — 매도 체결 귀속
//
//	row  state    disposition  next     note
//	X19  OPENING  WORKING      CLOSING
//	X20  OPENING  DONE         OPENING  a completed partial take; the entry works on
//	X21  OPENING  SUCCEEDED    CLOSING
//	X22  OPEN     WORKING      CLOSING  부분 청산 진행
//	X23  OPEN     DONE         OPEN     부분 청산 완료, 잔여 보유
//	X24  OPEN     SUCCEEDED    CLOSING
//	X25  SCALING  WORKING      CLOSING
//	X26  SCALING  DONE         SCALING
//	X27  SCALING  SUCCEEDED    CLOSING
//	X28  CLOSING  WORKING      CLOSING  청산 주문의 부분체결
//	X29  CLOSING  DONE         OPEN
//	X30  CLOSING  SUCCEEDED    CLOSING
//
// ## EXIT, Δ = quantity (FLATTENS) → CLOSED
//
//	row  state    disposition  next     note
//	X31  OPENING  WORKING      CLOSED   flat is flat even mid-entry
//	X32  OPENING  DONE         CLOSED
//	X33  OPENING  SUCCEEDED    CLOSED
//	X34  OPEN     WORKING      CLOSED
//	X35  OPEN     DONE         CLOSED
//	X36  OPEN     SUCCEEDED    CLOSED
//	X37  SCALING  WORKING      CLOSED
//	X38  SCALING  DONE         CLOSED
//	X39  SCALING  SUCCEEDED    CLOSED
//	X40  CLOSING  WORKING      CLOSED   전량 체결 시 CLOSED
//	X41  CLOSING  DONE         CLOSED
//	X42  CLOSING  SUCCEEDED    CLOSED
//
// ## EXIT, Δ > quantity (OVERSHOOTS) → RECONCILE
//
//	row  state    disposition  next       refusal
//	X43  FLAT     WORKING      RECONCILE  UNATTRIBUTED_SELL
//	X44  FLAT     DONE         RECONCILE  UNATTRIBUTED_SELL
//	X45  FLAT     SUCCEEDED    RECONCILE  UNATTRIBUTED_SELL
//	X46  OPENING  WORKING      RECONCILE  OVERSELL
//	X47  OPENING  DONE         RECONCILE  OVERSELL
//	X48  OPENING  SUCCEEDED    RECONCILE  OVERSELL
//	X49  OPEN     WORKING      RECONCILE  OVERSELL
//	X50  OPEN     DONE         RECONCILE  OVERSELL
//	X51  OPEN     SUCCEEDED    RECONCILE  OVERSELL
//	X52  SCALING  WORKING      RECONCILE  OVERSELL
//	X53  SCALING  DONE         RECONCILE  OVERSELL
//	X54  SCALING  SUCCEEDED    RECONCILE  OVERSELL
//	X55  CLOSING  WORKING      RECONCILE  OVERSELL
//	X56  CLOSING  DONE         RECONCILE  OVERSELL
//	X57  CLOSING  SUCCEEDED    RECONCILE  OVERSELL
//	X58  CLOSED   WORKING      RECONCILE  SELL_ON_CLOSED
//	X59  CLOSED   DONE         RECONCILE  SELL_ON_CLOSED
//	X60  CLOSED   SUCCEEDED    RECONCILE  SELL_ON_CLOSED
//
// # Why ENTRY_WHILE_CLOSING is a refusal and not a scale-in
//
// E31–E33 are the one refusal that is a judgement rather than an arithmetic
// impossibility. A buy filling while a sell is working is two contradictory
// instructions against one position, and the projection cannot tell which one
// the engine still believes. Calling it a scale-in would let the position grow
// while the protection path thinks it is shrinking; calling it RECONCILE blocks
// new entries, alerts, and leaves the adjustment path (task 6.2) to converge the
// quantity to the account's own value. That is the conservative direction
// (§0.9), and it costs nothing that cannot be recovered: the fill is still
// recorded, and the account is still the authority on the quantity.
//
// # Cost basis
//
// Quantity arithmetic is exact decimal-string addition and subtraction
// (internal/riskcalc). The cost basis needs multiplication and one division, so
// avg_price is the only rounded value in the projection; the rule and its bound
// are at DivideDecimal in decimal.go.
//
// Average-cost accounting: a BUY re-averages, a SELL does not move the unit
// cost (it realises P&L, which is trade_outcomes' business in task 8.1). The
// contribution of one order is `cumulative filled × average price`, so an
// observation's contribution to the position is the difference between that
// product now and the same product before — one formula that covers a fill, an
// EXECUTION_CORRECTION at unchanged quantity, and a terminal observation that
// moved nothing.
//
// An average price the broker did not report is "" and not 0. Treating it as 0
// would understate the cost basis, which understates the break-even price,
// which is the direction that sells at a loss believing it is flat. So an
// unknown price makes the instance's cost basis unknown for the rest of its
// life, and a reader that needs a number has to fail closed rather than be
// handed a wrong one.
//
// # Broker-behaviour claims
//
// None. The only broker facts this package leans on are the ones already
// established in internal/journal: the API reports a cumulative filled quantity
// with no per-fill identifier, and a modify answers with a new order number
// (docs/migration/openapi.latest.json, cited at fills.go and lineage.go).
package position
