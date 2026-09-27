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
3. **그 심볼의 OPEN 일반 주문이 0건이다** — 발동한 child 가 아직 체결되지 않았다면 그것은 일반 주문으로
   호가에 있다. `OrdersPageRaw(OrdersFilter{Status: "OPEN", Symbol: …})` 를 끝까지 읽는다.
4. **목록에서도 사라진 발동을 배제한다** `[비움 — Q1]`. 발동한 조건주문이 CLOSED 에 얼마나 오래 남는지
   (보존 기간)도, 발동 뒤 목록에 남는지도 **측정되지 않았다**(발동 측정은 deferred —
   verify-execution-capability 2.5). 2·3 은 목록에 남아 있는 동안만 덮는다. 목록 밖의 발동을 배제할 근거는
   verify-observes-the-trigger 가 채택한 **보유 수량** 뿐인데, `Artifact` 에는 방향도 수량도 없고
   (`record.go:182-243`) `raceEvidence` 는 `held < heldBefore` 만 비교한다(`steps_trigger.go:553`).
   정해지기 전에는 이 조건을 만족시킬 수 없으므로 명령은 거절한다.
5. **행이 대상과 같은 심볼·시장이다** — 읽은 모든 행의 `Symbol` 이 대상과 같아야 한다(심볼 필터가 지켜지지
   않은 응답은 거절). `Market` 은 행이 있을 때만 대조된다 — 1·3 이 통과하는 경우 행이 0건일 수 있으므로,
   그때 시장 결속은 기록 파일 이름과 심볼뿐이다(G3-4).
6. **읽기 집합이 흔들리지 않았다** — 컷 토큰이 없으므로 같은 읽기 집합(1·2·3 의 전 페이지)을 **두 번** 연속
   수행하고, 두 결과를 (그룹, id, status, triggeredOrderId) 로 **정렬한 집합**으로 비교해 같아야 한다.
   첫 읽기 시작부터 둘째 읽기 끝까지의 경과가 한도 `[비움 — Q3]` 를 넘으면 거절한다. 페이지 오류·반복
   커서·`hasNext` 인데 빈 커서·상한 도달(m0RecoverPending B4·B6·B13·B14 와 같은 모양)이면 거절한다.

**치르는 값(숨기지 않는다).** 페이지 상한이 10 × 100 이므로(`steps.go:61`, limit 100) CLOSED 이력이 그보다
긴 심볼은 영구히 거절된다. 2 의 `COMPLETED` 거절은 그 심볼에 과거 발동 이력이 하나라도 목록에 남아 있으면
영구히 거절한다. 둘 다 fail-closed 의 대가다.

이 규칙은 G1 을 **증명**하지 않는다 — 서버가 일관된 컷을 주지 않는 한 증명은 불가능하다. 대신 "살아 있는
것도 발동한 흔적도 그 심볼에 하나도 없고, 두 번의 읽기가 같다" 로 창을 좁히고, 좁힐 수 없는 모양은 전부
거절 쪽으로 보낸다. 리뷰가 요구한 fail-closed 대안이 이것이다.

### G2 — 사건의 신원: 기록이 이미 쓰는 키를 그대로 쓴다

**코드가 주는 것.**

- 기록은 artifact 를 이미 **원문 id** 로 담는다 — `Artifact.ID` 주석 "Printed in full on purpose: an
  operator who has to go and cancel something by hand needs it"(`record.go:185-187`). M0 체크포인트
  줄도 "may retain exact broker identifiers"(`record.go:85-88`). 계좌는 반대로 마스킹된 참조만
  담는다(`Entry.AccountRef`, `record.go:265-266`).
- 투영의 키는 `Kind + "\x00" + ID` 다(AST `verifylive--outstandinglines` B3, `record.go:548`).
  종결은 단조다 — terminal 줄 뒤의 비-terminal 줄은 되살리지 않는다(B4 `:552`). 투영은 비-terminal 만
  낸다(B6 `:562`). 종결 술어는 한 곳이다 — `Artifact.terminal()`(`record.go:575`)이고 그 주석이
  "a third ending added later is honoured by all of them at once" 라고 이 확장을 예정해 두었다(`:571-574`).
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
  `ReconciledAbsent bool` · `ReconciledAt time.Time` 을 세운다. `Artifact.terminal()` 이 그것을 셋째
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
  다시 읽어 첫 읽기와 지문이 같은지 대조하고, 다르면 거절한다.
- **구 바이너리**: `FormatVersion` 은 1 그대로다(필드 추가). 구 바이너리는 `reconciled_absent` 를 모르는
  필드로 버리므로 그 줄을 비-terminal 줄로 읽는다 — `outstandingLines` B4 는 terminal 뒤의 비-terminal
  만 막으므로 **구 바이너리에서는 그 artifact 가 여전히 outstanding** 이고, `holdGate` 기본값과 그 줄의 위치
  때문에 **다시 보유 상태**가 된다(다음 `conditional-cancel` 판정 전까지 정리 대상도 아니다). 틀리는 방향이
  "다시 보인다" 이므로 안전 쪽이고, 되돌림은 증거를 지우지 않는다.
- **원문 id 경계** `[비움 — Q2]`. 위 골격은 대사 줄에 **새 원문 id 를 하나도 더하지 않는다** — 같은
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
- 기록의 계좌는 줄마다 **마스킹된** `AccountRef` 다 — `maskedAccount(r.accountRef)`(`runner.go:278·688·890`)
  = `attest.Mask`, 끝 4자리만 남긴다(`internal/attest/attest.go:281-291`).
- 자격 증명은 환경 변수가 파일보다 우선한다 — `TOSSCTL_OPENAPI_KEY`·`SECRET` 이 있으면 파일을 읽지 않는다
  (`internal/official/credentials.go:30-34`).
- 기록 경로는 `resolveVerifyRecordFor` 가 정한다 — `--record` override 가 이기고(AST B1 `:765`),
  아니면 `--config-dir` 프로필 아래(B2 `:769`), 아니면 `journal.DataDir()` 다(자격 증명의 기본 경로와 다른
  뿌리). 시장마다 파일이 다르다(`RecordFileName`, `record.go:55-60`). **`Entry` 에는 시장 필드가 없다**
  (`record.go:247-279`).
- 원시 행은 브로커가 준 `Market`("KR"/"US")을 그대로 담는다(`conditional_reads.go:111`).

**결정.** 공식 읽기와 로컬 추가 **전에** 아래가 모두 참이어야 한다. 하나라도 아니면 거절한다.

1. **계좌**: 대상 artifact 를 언급한 모든 기록 줄의 `AccountRef` 가 서로 같고,
   `attest.Mask(TrimSpace(현재 참조)) == TrimSpace(entry.AccountRef)` 다. 비어 있거나 섞였으면 거절.
   이 대조는 끝 4자리 일치이지 신원 일치가 아니다 — 끝 4자리가 같은 두 계좌를 가르지 못한다(Q5 와 묶인다).
2. **헤더 계좌**: seq 가 0 이면 거절한다 — 지연 해석은 참조를 준 그 항목을 고른다는 보장을 이 경로에서
   재확인할 수 없다(`buildVerifyBroker` B5). 계좌 목록에 `DisplayName` 이 빈 값이 아닌 항목이 둘 이상이면
   거절한다(Q5 — Manager 결정 2026-09-27). 끝 4자리 충돌은 계좌가 둘 이상일 때만 생기므로 이 거절이 덮는다.
3. **프로필**: `--config-dir` 를 필수로 받고, 환경 변수 자격 증명이 설정돼 있으면 거절하고, `--record` override
   를 받지 않는다. 그러면 자격 증명 파일과 기록이 같은 `--config-dir` 아래에서 유도된다
   (`resolveVerifyRecordFor` B2) — 프로필 결속이 구성으로 성립하는 것은 **이 세 조건을 모두 걸었을 때뿐**이다.
4. **시장**: `--market` 을 필수로 받는다. 기록 파일 이름은 그 값으로 정해지므로(`verify.go:768`) 이것은
   파일 선택이지 대조가 아니다. 실제 대조는 G1-5 — 행이 있을 때 행의 `Market == market` 이다.

### 열린 질문 — 코드에 근거가 없어 사람이 정할 것

- **Q1 — 목록 밖의 발동 배제.** **사용자행(Manager 가 큐에 올림).** 발동한 조건주문이 CLOSED 에 남는지·얼마나 오래 남는지 측정되지 않았다.
  (a) 발동 목록 잔존과 보존 기간을 **측정한 뒤** 그 기간 안의 artifact 에만 G1-2·3 으로 충분하다고 본다
  — 이 경우에도 G1-2 는 대상 id 가 아니라 심볼의 모든 행을 본다, (b) 기록에 artifact 이후의 보유 수량
  기준값과 방향·수량이 있을 때만 보유 대조로 배제한다 — 현재 `Artifact` 에 방향·수량이 없으므로 기록 형식
  확장이 선행된다, (c) 이 change 에서는 G1-4 를 만족시킬 수 없음을 받아들이고 명령은 늘 거절한다(a063 이
  계속 막힘). 어느 것인가.
- **Q2 — 대사 줄의 신원 표현.** G2 의 (a) spec 경계 문구 수정 + 기존 키 재사용, (b) 도메인 태그. 어느 것인가. **사용자행(Manager 추천: (a)).**
- **Q3 — 두 번 읽기의 경과 한도.** 코드에 이 용도의 상수가 없다. 값(초)과 초과 시 거절을 정한다. **처리(Manager, 2026-09-27): 구현 로트로 미룬다 — 구현 시 보수 형태로 제안한다.**
- **Q4 — 같은 심볼의 다른 조건주문.** **결정(Manager, 2026-09-27): 거절. 후속 증명 예외는 만들지 않는다(범위 확장 금지).** 아래는 원 질문이다. 기본은 거절이다. 엔진·수동 보호 조건주문이 같은 심볼에 상주하는
  계좌에서는 대사가 영영 불가능해진다 — 그 비용을 받아들이는가, 아니면 기록이 소유를 증명할 수 있는
  예외(예: `ChainID` 로 연결된 후속)를 둘 것인가.
- **Q5 — 계좌가 여럿이거나 끝 4자리가 같은 자격 증명.** **결정(Manager, 2026-09-27): 거절.** `resolveVerifyAccount` 는 첫 항목을 고르고, 기록
  대조는 끝 4자리뿐이다. 후보가 둘 이상이면 거절할 것인가.
- **Q6 — 대상이 CLOSED 에 `EXPIRED` 로 있을 때.** **결정(Manager, 2026-09-27): 거절. 귀결 — a063 의 artifact 가 EXPIRED 면 이 경로로는 영구히 대사되지 않고, 그때 사용자 재결정 항목이 된다. 거절 메시지와 이 문서 양쪽에 적는다.** 기본은 거절이다(부재가 아니라 종결). a063 의 artifact 가
  만료로 끝났다면 이 change 로는 영영 대사되지 않는다. 만료를 별도의 종결 사건으로 받을 것인가(범위 확장),
  거절로 둘 것인가.
