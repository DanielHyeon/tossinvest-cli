# a121 design — read-backed artifact reconciliation

## Context

The current cleanup contract deliberately records a failed conditional DELETE
and leaves the record outstanding. That is correct for any 404 by itself: the
broker may have cancelled, expired, replaced, triggered, or otherwise removed
the object. A generic human observation of no active orders also lacks the
record's object-type and identity proof. The repair must add a different audit
fact, never reinterpret either input.

## Decision

Introduce a dedicated reconciliation operation that only reads official
conditional-order pages and appends one local record event when it can prove a
single record-owned conditional is absent. It does not call an order mutation.

The operation must:

1. select only an outstanding conditional artifact from the chosen verification
   record and its resolved profile/account context;
2. read every bounded page for both official `OPEN` and `CLOSED` groups using
   the recorded symbol and exact opaque identifier comparison;
3. reject repeated/missing cursors, incomplete pages, wrong profile/account,
   unsupported object type, stale snapshot, ambiguous identity, or any read
   failure without writing a terminal event;
4. append `reconciled-absent` only when the exact identifier is absent from the
   complete fresh authoritative result; and
5. make the event idempotent, retain a digest/basis for the read without raw
   account or broker identifiers, and leave the original failed cleanup entry
   intact.

`Outstanding` may stop presenting only an artifact with that explicit event.
The projection must label it reconciled absent, never cancelled or filled.
`PendingCleanup` and resume planning must not schedule a cancellation for that
artifact after reconciliation.

## Safety and attestation boundary

The operation is an account read plus local audit append. It must never be
counted in successful endpoint, verified capability, soak-attestation, or
engine-start evidence. It cannot change a failed `cleanup`, `conditional-*`,
or other measurement verdict. A DELETE 404 without the successful official read
remains outstanding.

The CLI must make the record/profile target explicit, redact operator output,
and refuse broad account scans or arbitrary identifiers. It must not offer an
order-mutation flag, hidden approval bypass, retry of cleanup, `--redo`, or a
new record path.

## Compatibility and rollback

Existing records without a reconciliation event retain their current behavior.
The event is append-only and ignored by older binaries as an unknown record
kind only if their parser already preserves forward-compatible entries; if that
is not true, schema/version handling must fail closed before write. Rollback
means keeping the event and using the prior binary only when it safely reads the
version; it never removes or rewrites evidence.

## Verification approach

Create RED tests for every fail-closed boundary before implementation. Use fake
official readers and isolated records only. Tests must prove no broker mutation
method is reachable, preserve the failed DELETE event, and prove the new event
does not influence endpoint or attestation success.

---

## Revision 1 (2026-09-27) — 적대 리뷰의 설계 갭 셋을 코드 영수증으로 닫는다

이 절은 위 Decision 1~5 를 **대체하지 않고 좁힌다.** 충돌하면 이 절이 이긴다. 근거 줄은 HEAD
`cb378a63` 기준이고, 함수 내부 분기를 근거로 쓰는 자리는 `analysis/ast-evidence/` 의 AST 를
먼저 만들어 그 열거로 인용한다. 코드에 근거가 없는 결정은 지어내지 않고 `[비움 — Qn]` 으로 남긴다
(끝의 「열린 질문」). 초안(3cf64d5d) 뒤 gstack 문서 리뷰 1회(P0 1 · P1 8 · P2 6)를 반영한 판이다 —
반영 대조는 review.md 「Revision 1」.

### 리뷰가 막은 갭 셋 (원문)

`analysis/proposal-freeze-adversarial-review.md`:

1. *"ConditionalOrdersRaw exposes only independently fetched OPEN and CLOSED pages with cursors; it
   carries no server snapshot revision, read-cut timestamp, or cross-group consistency token. …
   disappearance of the parent conditional is compatible with a triggered child order or successor
   artifact."*
2. *"The append-only event cannot currently be both exact and free of raw broker IDs. … neither
   document defines the event's identity encoding, version, collision/domain separation, or projection
   lookup rule."*
3. *"The proposal says 'chosen profile/account context' but does not require all reconciliation
   candidates to have a single matching record account/profile/market or specify how a masked record
   reference is compared with the current selected account."*

### G1 — 부재 증명: "id 가 없다" 가 아니라 "그 심볼에 이 기록이 모르는 살아 있는·발동한 것이 없다"

**코드가 주는 것.**

- 응답 페이지에는 스냅숏 필드가 없다 — `apiConditionalOrderPage` 는 `conditionalOrders` ·
  `nextCursor` · `hasNext` 셋뿐이다(`internal/official/conditional_reads.go:44-48`). 원시 행
  `RawConditionalOrder` 에도 없다(`:105-134`). **서버 컷 토큰은 없다 — 리뷰의 전제가 코드로 참이다.**
- 두 그룹의 어휘는 `ConditionalOrdersRaw` B1 의 오류 문구가 적는다: OPEN = `{WATCHING, PAUSED,
  ORDERING, ORDERED}`, CLOSED = `{COMPLETED, EXPIRED}` (`conditional_reads.go:158-164`). **취소된
  조건주문의 그룹은 어휘에 없다.** 발동한 조건주문은 child 의 id 를 `TriggeredOrderID` 로 싣는다
  (`:129-131`).
- 측정된 사실 둘이 "id 부재" 를 모호하게 만든다.
  - 정정은 새 id 를 발급하고 옛 id 를 즉시 404 로 만든다 — M19(US)·M40(양 시장), 기록은
    `Artifact.ChainID` 주석(`internal/verifylive/record.go:236-241`). **옛 id 가 없다는 것은 후속이 있다는
    것과 양립한다 — 살아 있든 이미 발동했든.**
  - 취소 뒤 by-id 읽기가 실패한다 — `conditional.cancel.gone_after=true`(M20). 다만 AST
    `verifylive--runner.stepconditionalcancel` B3(`steps.go:789`)는 **어떤 오류로든** 읽기가 실패하면
    `true` 를 적으므로, 그 측정은 "404" 가 아니라 "by-id 읽기 실패" 다. `steps_trigger.go:517-520` 이
    스스로 적는다: "a 404 says 'cancelled' and 'fired and gone' with the same words".
- 전 페이지·전 그룹 읽기의 선례는 이미 있다 — `Runner.m0RecoverPending`(AST B2 OPEN·CLOSED,
  B3·B14 페이지 상한 `maxFixturePages = 10` `steps.go:61`, B4 반복 커서, B13 빈 커서, B6 읽기 실패 →
  전부 HOLD). 그 함수는 M0 부작용(체크포인트·receipt)을 가지므로 **재사용하지 않고** 같은 검사 모양만
  새 읽기 전용 경로에 옮긴다(hard-evidence 4 와 같은 결론).
- 읽기 도구는 하나로 정한다 — **`ProtectionConditionalOrdersRaw`**(`internal/official/protection_reads.go:24-47`,
  M0 의 좁은 인터페이스 `m0RawConditionalPageReader` 와 같은 모양)와 일반 주문의 **`OrdersPageRaw`**
  (`orders_raw.go:62-92`). 앞의 것은 status 를 강제하지 않으므로 명령이 늘 `"OPEN"`/`"CLOSED"` 를 명시한다.
  `OrderSide`·`ClientOrderID` 를 보존하는 쪽이 이것이다(`ConditionalOrdersRaw` 는 둘을 채우지 않는다).

**결정.** 대사가 기록할 수 있는 것은 "이 id 가 목록에 없다" 가 아니라 다음이 **모두** 참인 상태다.
하나라도 확인할 수 없으면 아무것도 쓰지 않고 거절한다.

1. **그 심볼의 OPEN 조건주문이 0건이다** — id 와 무관하게. 옛 id 가 정정으로 사라졌다면 살아 있는 후속은
   같은 심볼의 OPEN 에 있다(정정은 심볼을 바꿀 수 없다 — 정정 바디 `ConditionalModifyBody` 의 필드는
   type·quantity·orderType·expireDate·first·second·confirmHighValueOrder 뿐이다,
   `internal/official/conditional_writes.go:52-60`). 예외는 두지 않는다(Q4 — Manager 결정 2026-09-27).
2. **그 심볼의 CLOSED 조건주문에 발동 흔적이 없다** — 대상 id 가 없고, **어떤 id 의 행이든**
   `TriggeredOrderID` 가 비어 있지 않거나 status 가 `COMPLETED` 이면 거절한다. 정정으로 id 가 바뀐 후속이
   발동하면 그것은 **다른 id** 로 CLOSED 에 남으므로(리뷰 P0), 대상 id 만 보면 놓친다. 대상 id 가 CLOSED 에
   `EXPIRED` 로 있으면 부재가 아니라 종결이고, 이 change 는 그것도 거절한다(Q6 — Manager 결정 2026-09-27). 거절 메시지는 "만료된 artifact 는 이 경로로 영구히 대사되지 않는다" 를 적는다.
   **(codex F6 — 2026-10-05 상향, P2-8 잔여 처리를 대체)**: 채택 리더는 첫 다리의 `TriggeredOrderID`
   만 싣고 `Second` 를 버린다(`protection_reads.go:49-57`, API 모델에는 `Second` 가 있다 —
   `conditional_reads.go:38-39`). 「둘째 다리 발동은 부모 COMPLETED 로 걸린다」는 미검증 가정이므로
   채택하지 않는다 — **대사 리더는 `Second` 존재를 그대로 노출하고, OPEN·CLOSED 어느 그룹이든
   `Second` 가 비어 있지 않은 행이 하나라도 있으면 거절한다**(OCO 는 이 경로로 대사하지 않는다,
   fail-closed — 치르는 값에 기록).
   **(codex F4)**: CLOSED 행 판정은 거부 목록이 아니라 **허용 목록**이다 — status 가 측정·골든으로
   확정한 종결 어휘(RED 로트가 `verify-execution-capability` 영수증/골든에서 전사해 allowlist 로
   커밋; 설계 산문이 지어내지 않는다) 안에 있고 필수 필드(id·status·symbol·market)가 전부 채워진
   행만 "흔적 없음" 판정의 표본이 된다. **status 결측·미지의 값·그룹과 모순된 값(CLOSED 에 OPEN
   계열)은 그 자체로 거절한다** — 두 번 읽기의 일치는 의미 유효성을 세우지 못한다.
3. **그 심볼의 OPEN 일반 주문이 0건이다** — 발동한 child 가 아직 체결되지 않았다면 그것은 일반 주문으로
   호가에 있다. `OrdersPageRaw(OrdersFilter{Status: "OPEN", Symbol: …})` 를 끝까지 읽는다.
4. **목록에서도 사라진 발동을 배제한다** (Q1 결정 (a) — 사용자 2026-09-28). CLOSED 발동 잔존·보존
   기간의 사람 실측(장중·조회 전용) **전에는 이 조건을 만족시킬 수 없으므로 거절한다**(잠정 (c));
   측정 뒤에는 보존 기간 안의 artifact 에 한해 2·3 이 이 조건을 덮고, 더 오래된 artifact 는 거절한다.
   **한도의 형태(freeze P0-1 수리, Manager 2026-10-05):** 한도는 **기간**(duration)이다 — 행 수 상한
   가능성은 페이지 상한 거절(G1-6)이 따로 덮는다. 나이의 기준점은 **대사할 outstanding 줄의
   `CreatedAt`** 이고, 그 값이 영(zero time)이면 거절한다(cleanup 줄은 영 시각을 실을 수 있다 —
   `cleanup.go:166-171`). 저장 형태: **비공개(unexported) 값이고 생산 빌드에서는 항상 nil**(측정
   부재 = 거절) — 시험만 seam 으로 주입하고, 실값은 측정 뒤 **별도 리뷰를 거친 상수 커밋**으로만
   들어온다(설정·플래그·환경 변수 경로 금지 — 사람이 숫자를 넣어 안전 게이트를 우회하는 문을 열지
   않는다). 생산 nil 은 구조(AST) 시험으로 핀한다.
   **(codex F5 — Q1 측정 요구 강화)**: 기간 측정만으로는 **행 수 기반 축출**을 배제하지 못한다 —
   서버가 최신 N 행만 보존하면 `hasNext=false` 가 정당하게 돌아오고 클라이언트 페이지 상한에
   닿지 않은 채 발동 흔적이 밀려날 수 있다. Q1 측정은 **축출 모형(기간 기반인지 행 수 기반인지)**
   을 함께 밝혀야 하고, 행 수 기반이거나 판별 불능이면 G1-4 는 충족 불가로 남아 명령은 계속
   거절한다(치르는 값에 추가).
   원 근거: 발동한 조건주문이 CLOSED 에 얼마나 오래 남는지
   (보존 기간)도, 발동 뒤 목록에 남는지도 **측정되지 않았다**(발동 측정은 deferred —
   verify-execution-capability 2.5). 2·3 은 목록에 남아 있는 동안만 덮는다. 목록 밖의 발동을 배제할 근거는
   verify-observes-the-trigger 가 채택한 **보유 수량** 뿐인데, `Artifact` 에는 방향도 수량도 없고
   (`record.go:182-244`) `raceEvidence` 는 `held < heldBefore` 만 비교한다(`steps_trigger.go:553`).
   정해지기 전에는 이 조건을 만족시킬 수 없으므로 명령은 거절한다.
5. **행이 대상과 같은 심볼·시장이다** — 읽은 모든 행의 `Symbol` 이 대상과 같아야 한다(심볼 필터가 지켜지지
   않은 응답은 거절). `Market` 은 행이 있을 때만 대조된다 — 1·3 이 통과하는 경우 행이 0건일 수 있으므로,
   그때 시장 결속은 기록 파일 이름과 심볼뿐이다(G3-4).
6. **읽기 집합이 흔들리지 않았다** — 컷 토큰이 없으므로 같은 읽기 집합(1·2·3 의 전 페이지)을 **두 번** 연속
   수행하고, 두 결과를 (그룹, id, status, triggeredOrderId) 의 **multiset** 으로 비교해 같아야 한다
   (freeze P1-4 수리: 집합이 아니라 multiset — 페이지 경계 이동이 만든 중복을 집합이 접어 숨긴다).
   **한 읽기 안에서 (그룹, id) 중복이 나오면 그 자체로 거절**하고, 읽기 순서는 「조건주문 OPEN →
   일반 OPEN → CLOSED」 로 고정한다(전이 창을 한 방향으로). 첫 읽기 시작부터 둘째 읽기 끝까지의
   경과가 Q3 한도(처리 2026-09-27: RED 로트가 이름 있는 보수 상수로 **확정**하고 리뷰가 승인 — 확정
   전 거절)를 넘으면 거절한다. **(codex F7 상향)**: Q3 창의 끝은 둘째 읽기 완료가 아니라 **추가
   승인 직전**이다 — 기록 재읽기·종목 조회를 포함한 전 과정이 한도 안이어야 하고, 추가 직전에
   관측 신선도와 Q1 나이를 **다시** 검사한다(로컬 잠금은 브로커 상태를 얼리지 않는다).
   **승인 순서(freeze 재검 P2)**: 사람 승인(mutating 명령 승인·두-마스크 확인)은 **목록 읽기 전**에
   받는다(계좌 읽기 뒤는 가능) — 승인 대기 시간이 Q3 창을 소모해 정상 실행이 전부 거절되는 모양을
   피한다. 페이지 오류·반복
   커서·`hasNext` 인데 빈 커서·상한 도달(m0RecoverPending B4·B6·B13·B14 와 같은 모양)이면 거절한다.
   **(codex F2)**: 읽기 응답은 **스키마 존재 검증**을 통과해야 한다 — `unwrapAndDecode` 는 `result`
   키 **부재**만 거절하고 `{"result":null}`·`{"result":{}}` 는 무오류 빈 페이지가 된다
   (`client.go:213-227` 실측, 어댑터는 빈 목록·`HasNext=false` 를 돌려준다). 대사 리더는 컬렉션
   키·페이지 필드의 **존재와 형**을 따로 단언하고(codex R2-1 보강: 존재만으로는 부족 — 컬렉션은
   **non-null 배열**, `hasNext` 는 **non-null 불리언**, 커서와 `hasNext` 의 일관성(`hasNext` 인데
   빈 커서 등)까지; null 값은 Go 제로값으로 접혀 `HasNext=false`·빈 목록이 된다), 결측·null·형
   불일치면 "빈 목록" 이 아니라 거절로 처리한다. 같은 검증을 일반 주문 페이지에도 적용한다
   (`orders_raw.go:47-50` 동일 제로값 노출) — 두 번 같은 기형 응답은 multiset 비교를 통과하고
   종목 조회 양성 대조도 못 가르므로, 이 검증 없이는 기형 응답이 가짜 부재가 된다.

**치르는 값(숨기지 않는다).** 페이지 상한이 10 × 100 이므로(`steps.go:61`, limit 100) CLOSED 이력이 그보다
긴 심볼은 영구히 거절된다. 2 의 `COMPLETED` 거절은 그 심볼에 과거 발동 이력이 하나라도 목록에 남아 있으면
영구히 거절한다. 둘 다 fail-closed 의 대가다.
**셋째 대가(freeze P1-5):** Q1(a) 아래에서 **a063 의 artifact 는 사실상 영구 거절일 공산이 크다** —
artifact 는 2026-09-06 이전 생성이라 측정 시점 나이가 보존 한도를 넘기 쉽고(G1-4 거절), 그 심볼에서
한도 안의 발동이 있으면 G1-2 가 거절한다. 어느 쪽이든 닫힌다. 그 경우 a063 잔여물의 처분은 **사용자
재결정 항목**이 된다(proposal a063 절에 동일 기록). 따라서 **Q1 측정 표본은 a063 artifact 의 심볼로
만들지 않는다**(같은 심볼의 발동 이력을 CLOSED 에 남기는 순간 G1-2 영구 거절을 확정짓는다).
**빈-응답 양성 대조(freeze P1-3):** 수락 경로는 세 그룹이 모두 비어 있는 상태라 G1-5·G1-6 이 공허하게
참이 될 수 있다. 통제 둘을 세운다 — (i) 조회 심볼 문자열은 호출자 입력이 아니라 **artifact 줄이 기록한
심볼 바이트 그대로**를 쓴다(철자 오류로 인한 가짜 부재 제거), (ii) 읽기 전후로 그 심볼의 **종목 조회
GET(instrument/quote)** 이 성공하고 응답이 같은 심볼을 되돌려야 한다 — 세션·심볼이 산 채로 목록만
비어 있음을 가른다. 잔여 위험(목록 endpoint 만의 서버측 공벡터)은 이중 읽기와 함께도 0 이 아니며,
그것은 기록한다 — 증명이 아니라 창 좁히기라는 이 절의 성격과 같다.
**codex 라운드 추가 대가(2026-10-05):** ④ OCO(`Second` 비공란) 행이 하나라도 보이면 그 심볼은
이 경로로 대사되지 않는다(F6 fail-closed). ⑤ Q1 측정이 축출 모형을 행 수 기반·판별 불능으로
판정하면 G1-4 는 영구 충족 불가다(F5). ⑥ 끝 4자리 마스크는 신원이 아니다 — 자격 교체 + 접미
충돌 + 사람 오승인이 결합하면 다른 계좌의 빈 목록으로 대사될 수 있고, a121 은 이를 사람 승인
두-마스크 표시와 seq 명시 재결속으로 좁히되 없애지는 못한다(F1·R2-5 — 대사 줄 계좌 digest 는
추측 가능성 때문에 철회했고, 내구 결속은 기록 작성기 후속 change).

이 규칙은 G1 을 **증명**하지 않는다 — 서버가 일관된 컷을 주지 않는 한 증명은 불가능하다. 대신 "살아 있는
것도 발동한 흔적도 그 심볼에 하나도 없고, 두 번의 읽기가 같다" 로 창을 좁히고, 좁힐 수 없는 모양은 전부
거절 쪽으로 보낸다. 리뷰가 요구한 fail-closed 대안이 이것이다.

### G2 — 사건의 신원: 기록이 이미 쓰는 키를 그대로 쓴다

**코드가 주는 것.**

- 기록은 artifact 를 이미 **원문 id** 로 담는다 — `Artifact.ID` 주석 "Printed in full on purpose: an
  operator who has to go and cancel something by hand needs it"(`record.go:185-187`). M0 체크포인트
  줄도 "may retain exact broker identifiers"(`record.go:85-88`). 계좌는 반대로 마스킹된 참조만
  담는다(`Entry.AccountRef`, `record.go:265-266`).
- 투영의 키는 `Kind + "\x00" + ID` 다 — 키는 `record.go:547` 의 대입에서 만들어지고(분기 아님),
  AST `verifylive--outstandinglines` B3(`:548`)은 처음 본 키의 순서를 적는 분기다(로트 1 정정 D1).
  종결은 단조다 — terminal 줄 뒤의 비-terminal 줄은 되살리지 않는다(B4 `:552`). 투영은 비-terminal 만
  낸다(B6 `:562`). 종결 술어는 한 곳이다 — `Artifact.terminal()`(`record.go:575`)이고 그 주석이
  "a third ending added later is honoured by all of them at once" 라고 이 확장을 예정해 두었다(`:571-574`).
  (정확화 — freeze P2-5: `Cancelled`/`Filled` 를 직접 읽는 자리가 셋 있다 — `m0_manual.go:11`·
  `runner.go:1033`·`steps.go:1152`. 셋 다 원본·진행 중 줄만 훑어 현재는 영향 0 이지만, 「한 곳」은
  문자 그대로는 아니다.)
- 줄의 종류는 이미 여럿이다(`KindStep`·`KindApproval`·`KindCleanup`·`KindM0Checkpoint`,
  `record.go:65-88`). `decodeEntry` 는 `FormatVersion > 1` 만 거절한다(AST B2 `:442`) — 모르는 `Kind`
  나 필드는 거절하지 않는다(`json.Unmarshal`, B1).
- 성공 endpoint 증거는 `Kind` 와 무관하게 **모든 줄의 `Calls`** 에서 온다(`SucceededEndpoints` AST B1·B3,
  `internal/verifylive/endpoints.go:82-104`).
- 정리 대상 선택은 이미 한 함수가 한다 — `PendingCleanup`(`cleanup.go:119-121`)이 보유 게이트
  (`holdGate` AST B1·B2 `:155-163` — `HeldUntil` 이 비어도 조건주문은 `conditional-cancel` 까지 보유, 해제는
  `cleanupFrom` AST B3 `settled && heldAfter` `:131`)와 M0 수동 대사 제외(`withoutM0ManualReconcile`, `m0_manual.go:45`)를
  함께 적용한다.
- 비-원문 태그의 선례가 있다 — `CausalReceipt.tag(id) = sha256(runID + "\x00" + id)`
  (`internal/verifylive/receipt.go:353-361`). 기록 지문은 `Digest`(16바이트 sha256,
  `record.go:583-595`)이고 입력을 받은 순서대로 `json.Marshal` 한다.

**결정(골격).**

- 새 줄 종류 `KindReconcile = "reconcile"` 한 줄이 대상 artifact 하나를 담는다. 그 artifact 의 신원 필드
  (`Kind`·`ID`·`Symbol`·`ChainID`)는 **대사할 outstanding 줄의 값 그대로**이고, 새 종결 필드
  `ReconciledAbsent bool` · `ReconciledAt time.Time` 을 세운다. 나머지 필드(freeze P2-6, 재검 P2-c
  정정): `Verdict` 빈 값(판정 아님), `Mutating` false(브로커 변이 없음 — CLI 주석 S2 의 mutating=true
  와는 별개 축), `HeldUntil` 영, `CreatedAt` = **대사할 outstanding 줄의 값 그대로**(생성 시각 의미
  보존 — 추가 시각은 `ReconciledAt` 가 담고, 구 바이너리 되돌림에서도 생성 시각이 바뀌어 보이지
  않는다), `StepID` 는 카탈로그 밖 고유 값(아래 R1). `Artifact.terminal()` 이 그것을 셋째
  종결로 포함한다 — 그러면 `outstandingLines` 를 쓰는 소비자 전부(`PendingCleanup`, `liveCount`
  `mutate.go:679`, report·status·abort·redo)가 한 번에 따른다. report·status 는 그것을 **reconciled absent**
  로 표시하고 cancelled·filled 로 쓰지 않는다(`BuildReport`·`BuildProgress` 는 task 1.2 의 편집 대상).
- **대사 줄은 `Calls` 를 싣지 않는다** — 싣는 순간 그 GET 들이 성공 endpoint 가 되고(`endpoints.go:82-104`),
  그 효과는 구 바이너리로 되돌려도 남는다. 읽기 근거는 관측 `reconcile.basis` 하나에만 적는다.
- **근거 지문**: G1-6 의 정렬 집합을 버전·도메인 태그(`"a121/reconcile-basis/v1"`)와 함께 정규 순서로
  직렬화해 `Digest` 한다. 페이지 원문은 적지 않는다.
- **선택**: 호출자는 id 를 주지 않는다. 후보는 `PendingCleanup(entries)` 중 `Kind == "conditional-order"` 이고,
  `M0Unsettled`(`m0_recovery.go:23`)가 가리키는 artifact 는 제외한다. **정확히 하나가 아니면** 거절한다
  (리뷰: "reject zero or multiple eligible outstanding conditionals"). 보유가 풀리지 않은 조건주문은
  `PendingCleanup` 이 내지 않으므로 후보가 되지 않는다 — a063 의 artifact 는 실패한 `conditional-cancel` 이
  보유를 풀었으므로 후보다.
- **멱등**: 대상이 이미 종결이면 `outstandingLines` 가 그것을 내지 않으므로 선택 단계에서 "대사할 것
  없음" 으로 거절한다.
- **동시성**: `verify run` 과 같은 배제를 쥔다 — journal 실행 flock(`acquireVerifyExecutionLock`,
  `cmd/tossctl/verify.go:439`)과 rate-budget lease(`acquireVerifyRateBudget` `:796-807`). 추가 직전에 기록을
  다시 읽어 첫 읽기와 지문이 같은지 대조하고, 다르면 거절한다. 지문의 정의(freeze P2-7): **파일 원문
  바이트의 sha256**. **(codex F3 + R2-3 — 두 검사를 **병존**시킨다)**: `LoadEntries` 는 마지막
  비공백 줄이 해독 불능이면 **개행 여부와 무관하게** 조용히 버린다(`record.go:423-430` 실측 —
  꼬리 줄만 관용). 개행 검사**만으로는** 못 닫으므로 대사 경로는 선택 전과 추가 직전에 **기록의
  모든 비공백 줄을 엄격 해독**하고(관용 경로 사용 금지) 해독 불능 줄이 하나라도 있으면 거절한다.
  **그리고 개행 꼬리 검사도 유지한다**(R2-3): `Recorder.Append` 는 새 객체 **뒤**에만 개행을
  붙이므로(`record.go:326-345`) 개행 없는 완전한 JSON 꼬리에 이어 붙으면
  `{"a":…}{"b":…}` 로 붙어 둘 다 깨진다 — 엄격 해독은 이 모양을 추가 **전**에는 못 본다.
  추가는 「전 줄 엄격 해독 통과 ∧ 파일이 `\n` 으로 끝남」일 때만 허용한다.
- **브로커 획득(no-live-mutation 봉인 — freeze P1-1)**: 대사 경로는 `verifylive.Broker` 도
  `*official.Client` 도 받지 않는다. **읽기 전용 좁은 인터페이스**(`ProtectionConditionalOrdersRaw` ·
  `OrdersPageRaw` · `Accounts` · 종목 조회 GET — 뒤 둘은 G3-2 자체 판정(P1-2)과 양성 대조(P1-3)용)를
  새로 정의하고, `Broker` 를 절대 반환하지 않는 **전용 좁은 생성자**가 그것을
  만든다(기존 `verifyBrokerFactory`·type assertion 선례 `m0_recovery.go:129` 재사용 금지 — 단언이
  가능한 구체 객체가 스코프에 있으면 봉인이 아니다). 구조 시험: 대사 파일들(cmd·verifylive 양쪽)의
  AST census 가 쓰기 메서드 7개 이름(PlaceOrder·CancelOrder·ModifyOrder·CreateConditionalOrder·
  ModifyConditionalOrder·ModifyConditionalOrderRef·CancelConditionalOrder) 호출과 **type assertion
  자체**를 금지한다. 정확화(P2-9): 토큰 갱신 POST(`token.go:138`)는 auth 기반이라 이 경계 밖이며
  그 사실을 경계 서술에 명시한다 — 「주문·조건주문 변이 도달 0」 이 주장의 전부다.
- **재개 계획에 주는 효과(freeze P1-6)**: 대사 뒤 `subjectLost`(`redo.go:122`)가 참이 되어
  `conditional-register` 가 `RedoSet` 으로 돌아갈 수 있다 — 콘솔 재개가 **새 조건주문 설치를 제안**하게
  된다(여전히 사람 일괄 승인 뒤에만 실행). 이 change 는 그 동작을 바꾸지 않고, RED 가 대사 전/후의
  `RedoSet` 을 핀한다(tasks 2.3).
- **구 바이너리**: `FormatVersion` 은 1 그대로다(필드 추가). 구 바이너리는 `reconciled_absent` 를 모르는
  필드로 버리므로 그 줄을 비-terminal 줄로 읽는다 — `outstandingLines` B4 는 terminal 뒤의 비-terminal
  만 막으므로 **구 바이너리에서는 그 artifact 가 여전히 outstanding** 이고, `holdGate` 기본값과 그 줄의 위치
  때문에 **다시 보유 상태**가 된다(다음 `conditional-cancel` 판정 전까지 정리 대상도 아니다). 틀리는 방향이
  "다시 보인다" 이므로 안전 쪽이고, 되돌림은 증거를 지우지 않는다.
- **원문 id 경계** (Q2 결정 (a) — 사용자 2026-09-28: spec 경계 문구 수정 + 기존 키 재사용,
  `outstandingLines` 는 편집 대상 아님). 위 골격은 대사 줄에 **새 원문 id 를 하나도 더하지 않는다** — 같은
  기록이 이미 담은 id 를 같은 키로 반복할 뿐이다. 그러나 현재 delta spec 은 원문 브로커 식별자 경계를 두고
  리뷰는 "a raw ID violates the stated boundary" 라고 읽었다. 둘 중 하나를 골라야 한다 — (a) spec 문구를
  "대사 줄은 대상 artifact 줄이 이미 담은 식별자 외에 새 브로커 식별자를 더하지 않고, 읽기 근거는 지문으로만
  남긴다" 로 고친다, (b) 대사 줄은 `tag` 선례처럼 sha256(도메인·기록 run·id) 태그만 담고 `outstandingLines`
  가 각 artifact 의 태그를 계산해 대조한다. (b) 는 투영의 키 규칙(B3)을 바꾸므로 그 함수가 편집 대상이 된다.

### G3 — 계좌·프로필·시장의 같음: 실행 가능한 셋

**코드가 주는 것.**

- 현재 계좌 참조는 `resolveVerifyAccount` 가 정한다 — `DisplayName` 이 빈 값이 아닌 **첫** 계좌의
  `DisplayName` 이고(AST B5·B6 `verify.go:941-945`), 그 값은 공식 응답의 **원문 계좌번호**다
  (`internal/official/reads.go:93` `DisplayName: a.AccountNo`). 계좌 id 가 숫자가 아니면 seq 는 0 이다
  (B7 `:947`). seq 가 0 이면 `buildVerifyBroker` 는 헤더 계좌를 클라이언트의 지연 해석에 맡긴다(AST B5
  `:894`). 주석이 두 값을 한 항목에서 가져오는 이유를 적는다: "taking them from different accounts
  would produce a record that names one account and measured another"(`:911-914`).
  **결정(freeze P1-2, Manager 2026-10-05):** G3-2 의 「계좌가 정확히 하나·seq ≠ 0」 판정은
  `resolveVerifyAccount`·`buildVerifyBroker` 를 **편집하지 않고**, 대사 전용 경로가 **자체
  `Accounts()` 호출**로 수행한다 — 후보 전수를 받아 DisplayName 비공란 계좌가 정확히 하나인지 세고,
  seq 는 그 항목에서 직접 해석하며 0 이면 거절한다(P1-1 의 좁은 생성자가 이 읽기도 소유 — `Broker`
  경유 금지). 두 함수는 비편집이므로 FLM 대상에서 빠지고, 새 경로는 새 파일의 새 코드다.
  **(codex R2-2 보강)**: 검증한 seq 와 transport 가 실제 쓰는 seq 는 다를 수 있다 — `Accounts()`
  는 **첫 양수 seq 를 자동 캐시**하고(`reads.go:38-54`, `client.go:244-250`) 이후 범위 지정 읽기가
  그 캐시를 탄다. 좁은 생성자는 검증 통과한 seq 로 **명시적으로 재결속**해야 하며(`WithAccountSeq`
  선례 — `verify.go:899-905`), 기형 계좌 행(빈 accountNo + 양수 seq 등)은 계수에서 빼는 게 아니라
  **그 자체로 거절**한다 — 안 그러면 표시는 계좌 B 를 말하고 부재는 계좌 A 에서 읽힌다.
  **(codex F1 — 마스크 결속의 한계와 처분, Manager 2026-10-05):** 기록의 계좌 결속은 끝 4자리
  마스크뿐이라, 같은 프로필 디렉터리의 자격 증명을 **끝 4자리가 같은 다른 계좌**로 바꾸면 모든
  검사(단일 계좌·마스크 일치·seq≠0)를 통과한 채 B 계좌의 빈 목록으로 A 계좌의 artifact 를 대사할
  수 있다. 전면 수리(기록 작성 시점의 내구 신원 digest)는 verify 기록 **작성기** 편집이라 a121
  범위 밖이다. a121 의 처분(codex R2-5 로 개정 — 2026-10-05): ~~대사 줄 계좌 digest~~ **철회** —
  sha256(도메인 ‖ run id ‖ 계좌번호)는 키 없는 구성이라 기록을 읽는 쪽이 계좌번호(저엔트로피)를
  추측 대조할 수 있고, 그것은 사실상 계좌 식별자의 난독 저장이다(안전 불변식 8 위반 소지 +
  "keyed" 라는 spec 표기와 모순). 대사 줄은 **마스크 참조만** 싣는다. 남는 완화: ① 4.3 의 사람
  승인 출력에 기록 마스크·현재 계좌 마스크·계좌 수를 **반드시 표시**(명령은 mutating=true 라
  승인 없이 돌지 않는다), ② R2-2 의 seq 명시 재결속(위), ③ 잔여(자격 교체 + 접미 충돌 + 사람
  오승인의 결합)는 치르는 값에 적는다. 내구 결속(제대로 된 keyed 구성 + 키 수명 주기)은 기록
  **작성기** 후속 change 범위로 proposal 에 기록한다.
- 기록의 계좌는 줄마다 **마스킹된** `AccountRef` 다 — `maskedAccount(r.accountRef)`(`runner.go:278·688·890`)
  = `attest.Mask`, 끝 4자리만 남긴다(`internal/attest/attest.go:281-291`).
- 자격 증명은 환경 변수가 파일보다 우선한다 — 단, 실측(로트 1 D4): 우선은 **둘 다** 비어 있지 않을 때만이다
  (`internal/official/credentials.go:30-34`, `:32`). G3 의 거절 조건은 보수 쪽으로 고정한다(Manager
  2026-10-05): **둘 중 하나라도 설정돼 있으면 거절** — 하나만 설정된 반쪽 상태는 어느 자격으로 읽었는지가
  모호해지는 쪽이므로 fail-closed.
- 기록 경로는 `resolveVerifyRecordFor` 가 정한다 — `--record` override 가 이기고(AST B1 `:765`),
  아니면 `--config-dir` 프로필 아래(B2 `:769`), 아니면 `journal.DataDir()` 다(자격 증명의 기본 경로와 다른
  뿌리). 시장마다 파일이 다르다(`RecordFileName`, `record.go:55-60`). **`Entry` 에는 시장 필드가 없다**
  (`record.go:247-279`).
- 원시 행은 브로커가 준 `Market`("KR"/"US")을 그대로 담는다(`conditional_reads.go:111`).

**결정.** 주문·조건주문 **목록 읽기**와 로컬 추가 **전에** 아래가 모두 참이어야 한다(freeze 재검
P0-R1 정정: 이 검사 자신이 쓰는 계좌 목록 읽기와 종목 조회 GET 만이 그보다 앞설 수 있는 공식
읽기다). 하나라도 아니면 거절한다.

1. **계좌**: 대상 artifact 를 언급한 모든 기록 줄의 `AccountRef` 가 서로 같고,
   `attest.Mask(TrimSpace(현재 참조)) == TrimSpace(entry.AccountRef)` 다. 비어 있거나 섞였으면 거절.
   이 대조는 끝 4자리 일치이지 신원 일치가 아니다 — 끝 4자리가 같은 두 계좌를 가르지 못한다(Q5 와 묶인다).
2. **헤더 계좌**: seq 가 0 이면 거절한다 — 지연 해석은 참조를 준 그 항목을 고른다는 보장을 이 경로에서
   재확인할 수 없다(`buildVerifyBroker` B5). 계좌 목록에 `DisplayName` 이 빈 값이 아닌 항목이 둘 이상이면
   거절한다(Q5 — Manager 결정 2026-09-27). ~~끝 4자리 충돌은 계좌가 둘 이상일 때만 생긴다~~
   **(codex F1 로 반증된 문장 — 2026-10-05 정정)**: 자격 증명을 접미가 같은 **다른 단일 계좌**로
   교체하면 후보가 하나여도 충돌한다. 이 거절이 덮는 것은 「한 자격에 후보 둘 이상」뿐이고,
   교체 잔여는 G3 F1 처분(사람 승인 두-마스크 표시)과 치르는 값 ⑥이 담당한다.
3. **프로필**: `--config-dir` 를 필수로 받고, 환경 변수 자격 증명이 설정돼 있으면 거절하고, `--record` override
   를 받지 않는다. 그러면 자격 증명 파일과 기록이 같은 `--config-dir` 아래에서 유도된다
   (`resolveVerifyRecordFor` B2) — 프로필 결속이 구성으로 성립하는 것은 **이 세 조건을 모두 걸었을 때뿐**이다.
4. **시장**: `--market` 을 필수로 받는다. 기록 파일 이름은 그 값으로 정해지므로(`verify.go:768`) 이것은
   파일 선택이지 대조가 아니다. 실제 대조는 G1-5 — 행이 있을 때 행의 `Market == market` 이다.

### 열린 질문 — 코드에 근거가 없어 사람이 정할 것

- **Q1 — 목록 밖의 발동 배제.** **결정(사용자, 2026-09-28 "일괄 수용 예외 없음"): (a) — 실측 승인.**
  발동 목록 잔존·보존 기간의 측정이 큐에 올랐다. 측정은 발동한 조건주문의 자연 발생을 요구하므로
  사람 실측 항목이다(장중·조회 전용). 측정 결과가 나오기 전에는 G1-4 를 만족시킬 수 없으므로
  대사 명령은 거절만 한다 — (c) 의 잠정 상태가 측정 완료까지 유지된다. 원 질문은 아래에 남긴다.
  발동한 조건주문이 CLOSED 에 남는지·얼마나 오래 남는지 측정되지 않았다.
  (a) 발동 목록 잔존과 보존 기간을 **측정한 뒤** 그 기간 안의 artifact 에만 G1-2·3 으로 충분하다고 본다
  — 이 경우에도 G1-2 는 대상 id 가 아니라 심볼의 모든 행을 본다, (b) 기록에 artifact 이후의 보유 수량
  기준값과 방향·수량이 있을 때만 보유 대조로 배제한다 — 현재 `Artifact` 에 방향·수량이 없으므로 기록 형식
  확장이 선행된다, (c) 이 change 에서는 G1-4 를 만족시킬 수 없음을 받아들이고 명령은 늘 거절한다(a063 이
  계속 막힘). 어느 것인가.
- **Q2 — 대사 줄의 신원 표현.** **결정(사용자, 2026-09-28 "일괄 수용 예외 없음"): (a) — spec 경계 문구
  수정 + 기존 키 재사용(Manager 추천안 채택).** 원 질문: G2 의 (a) spec 경계 문구 수정 + 기존 키 재사용,
  (b) 도메인 태그.
- **Q3 — 두 번 읽기의 경과 한도.** 코드에 이 용도의 상수가 없다. 값(초)과 초과 시 거절을 정한다. **처리(Manager, 2026-09-27): 구현 로트로 미룬다 — 구현 시 보수 형태로 제안한다.**
- **Q4 — 같은 심볼의 다른 조건주문.** **결정(Manager, 2026-09-27): 거절. 후속 증명 예외는 만들지 않는다(범위 확장 금지).** 아래는 원 질문이다. 기본은 거절이다. 엔진·수동 보호 조건주문이 같은 심볼에 상주하는
  계좌에서는 대사가 영영 불가능해진다 — 그 비용을 받아들이는가, 아니면 기록이 소유를 증명할 수 있는
  예외(예: `ChainID` 로 연결된 후속)를 둘 것인가.
- **Q5 — 계좌가 여럿이거나 끝 4자리가 같은 자격 증명.** **결정(Manager, 2026-09-27): 거절.** `resolveVerifyAccount` 는 첫 항목을 고르고, 기록
  대조는 끝 4자리뿐이다. 후보가 둘 이상이면 거절할 것인가.
- **Q6 — 대상이 CLOSED 에 `EXPIRED` 로 있을 때.** **결정(Manager, 2026-09-27): 거절. 귀결 — a063 의 artifact 가 EXPIRED 면 이 경로로는 영구히 대사되지 않고, 그때 사용자 재결정 항목이 된다. 거절 메시지와 이 문서 양쪽에 적는다.** 기본은 거절이다(부재가 아니라 종결). a063 의 artifact 가
  만료로 끝났다면 이 change 로는 영영 대사되지 않는다. 만료를 별도의 종결 사건으로 받을 것인가(범위 확장),
  거절로 둘 것인가.

### 로트 1 처분 (Manager, 2026-10-05 — 근거 `analysis/code-context/evidence-reconciliation.md`)

- **S1 — 표시 범위.** `Report.WriteText`(`report.go:239-291`)·`Progress.WriteText`(`:346-379`)를
  편집 대상에 **추가**한다(reconciled absent 라벨이 텍스트 출력에 닿으려면 필요 — FLM 은 구현 로트
  시작 시 생성, tasks 1.2 의 "at least" 아래). 콘솔 템플릿(`templates.go:620`·`:700`)은 **비편집**:
  대사된 artifact 는 새 바이너리에서 outstanding 투영에서 빠지는 것으로 충분하고, UI 에 대사 사건을
  새로 그리는 것은 범위 밖이다.
- **S2 — 명령 주석.** 새 대사 명령은 `mutating: true` 로 등재한다(보수 — 기록에 영속 이벤트를 쓰고,
  tasks 4.3 이 사람 승인을 요구하는 것과 정합; 대화형 에이전트 자동 실행 금지가 따라온다).
  `TestMutatingAnnotationOnTradeCommands` 의 고정 집합 갱신은 구현 로트 몫이다.
- **R1 — 대사 줄의 StepID.** `StepID` 만 비교하고 Kind 를 보지 않는 소비자는 **넷**이다(freeze P2-4
  정정): `LastEntry`(`record.go:478`)·`heldAfter`(`cleanup.go:186`)·`m0ManualReconcileIDs`
  (`m0_manual.go:9`)·`baselineSellable`(`steps.go:1180`). 대사 줄의 StepID 는 **`Steps()` 카탈로그 ∪
  {cleanup, abort} 밖의 고유 값**이어야 하며 특히 `conditional-cancel` 재사용 금지 — 구조 시험이 그
  합집합 부재를 단언한다(tasks 2.2.2).
- **STORY acceptance 2 판독.** "Only … appends" 는 **제약**으로 읽는다(능력 아님) — Q1 측정 전
  거절-전용 상태에서도 충족된다. Story 본문은 수정하지 않는다.

### RED 로트 처분 (Manager, 2026-10-05 — 착지 ffa8eb0d, A-RED 리뷰 2.5라운드)

- **가드 순서 계약(GREEN 구속, A-RED P2-5 — Q6 도달이 이 순서에 의존하므로 load-bearing):**
  사전 검사(기록·프로필·계좌, 목록 읽기 전) → 승인(두-마스크, 목록 읽기 전) → 읽기 전 Q1/Q3 →
  종목 조회 → 두 번 읽기(페이지·중복 검사 포함, 순서 조건 OPEN→일반 OPEN→CLOSED) → 종목 조회 →
  신선도·Q1 재검 → multiset 비교 → 행 검사(OCO→필드 결측→심볼→시장) → 부재 검사(OPEN 조건 →
  일반 OPEN → CLOSED target(EXPIRED→Q6) → 발동 흔적) → **status allowlist 는 마지막** →
  엄격 해독 → 지문 → 개행 → 추가. mixed 가 mismatch 보다 먼저.
- **Q3 확정: `reconcileFreshnessBoundValue = 15s`, 429 재시도 없음**(한 번의 429 로 한도를 넘게
  보수 설정 — 거절 후 사람 재실행; 근거 `analysis/red-lot/q3-freshness-and-closed-allowlist.md`).
  이 절이 코드 주석의 「리뷰 승인」의 원장 기록이다(A-RED P2-c).
- **RED 구체화 승인(구현자 판정 요청 7건):** ① `Client.Reconcile*` 해독 메서드 신설(기존 어댑터가
  null 을 접으므로 필연 — 같은 endpoint·새 해독·새 파일) ② `Accounts` 는 cmd 좁은 생성자 소유,
  verifylive `ReconcileReader` 는 GET 3개 ③ 설계 밖 거절 코드(m0-unresolved·account-ref-unusable)
  + `RefuseRecordFormat` 추인 — 전부 시험 참조 ④ `second` **키 부재 = 비-OCO 로 규정**(null 과
  동일 decode; F6 은 비공란만 거절) ⑤ 도달 census allowlist 에 Recorder 3이름(OpenRecorder·
  Append·Close) 포함 — 설계가 명한 추가 1회의 유일 통로 ⑥ M0 제외가 현 `withoutM0ManualReconcile`
  와 중복(두 판정 — 한쪽이 가려짐): GREEN 로트에서 층 내려 변이로 가를 것(RED 후보) ⑦ 측정 한계
  3(암묵 캐시 vs 명시 재결속 동일 헤더·읽기 순서는 호출 열 핀·F7 은 종목 조회 지연 모의) 기록 수용.
- **F4 allowlist 전사 출처 확정:** 저장소의 어떤 영수증·골든에도 CLOSED 종결 어휘가 없다(RED 로트
  실측 — verify-execution-capability 에 COMPLETED/EXPIRED 관측 0). **Q1 사람 실측 세션이 CLOSED
  status 어휘 영수증을 함께 채집**하고 그 전사가 allowlist 가 된다. 전사 전 allowlist 공집합 =
  CLOSED 행이 하나라도 있으면 거절(치르는 값 — 타 id EXPIRED 하나로도 거절됨을 기록).
