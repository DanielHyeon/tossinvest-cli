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
(끝의 「열린 질문」).

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

### G1 — 부재 증명: "id 가 없다" 가 아니라 "그 심볼에 이 기록이 모르는 살아 있는 것이 없다"

**코드가 주는 것.**

- 응답 페이지에는 스냅숏 필드가 없다 — `apiConditionalOrderPage` 는 `conditionalOrders` ·
  `nextCursor` · `hasNext` 셋뿐이다(`internal/official/conditional_reads.go:44-48`). 원시 행
  `RawConditionalOrder` 에도 없다(`:105-134`, AST `official--client.conditionalordersraw` B7 은
  행을 복사할 뿐). **서버 컷 토큰은 없다 — 리뷰의 전제가 코드로 참이다.**
- 두 그룹의 어휘는 `ConditionalOrdersRaw` B1 의 오류 문구가 적는다: OPEN = `{WATCHING, PAUSED,
  ORDERING, ORDERED}`, CLOSED = `{COMPLETED, EXPIRED}` (`conditional_reads.go:158-164`). **취소된
  조건주문의 그룹은 어휘에 없다.**
- 측정된 사실 둘이 "id 부재" 를 모호하게 만든다.
  - 정정은 새 id 를 발급하고 옛 id 를 즉시 404 로 만든다 — M19(US)·M40(양 시장), 기록은
    `Artifact.ChainID` 주석(`internal/verifylive/record.go:236-241`). **옛 id 가 없다는 것은 살아 있는
    후속이 있다는 것과 양립한다.**
  - 취소 뒤 by-id 읽기가 실패한다 — `conditional.cancel.gone_after=true`(M20). 다만 AST
    `verifylive--runner.stepconditionalcancel` B3(`steps.go:789`)는 **어떤 오류로든** 읽기가 실패하면
    `true` 를 적으므로, 그 측정은 "404" 가 아니라 "by-id 읽기 실패" 다. 그리고
    `steps_trigger.go:517-520` 이 스스로 적는다: "a 404 says 'cancelled' and 'fired and gone' with the
    same words".
- 전 페이지·전 그룹 읽기의 선례는 이미 있다 — `Runner.m0RecoverPending`(AST B2 OPEN·CLOSED,
  B3·B14 페이지 상한 `maxFixturePages = 10` `steps.go:61`, B4 반복 커서, B12·B13 빈 커서, B6 읽기 실패
  → 전부 HOLD). 그 함수는 M0 부작용(체크포인트·receipt)을 가지므로 **재사용하지 않고** 같은 검사
  모양만 새 읽기 전용 경로에 옮긴다(hard-evidence 4 와 같은 결론).

**결정.** 대사가 기록할 수 있는 것은 "이 id 가 목록에 없다" 가 아니라 다음 넷이 **모두** 참인 상태다.
하나라도 확인할 수 없으면 아무것도 쓰지 않고 거절한다.

1. **그 심볼의 OPEN 그룹이 비어 있다** — id 와 무관하게, 대상 artifact 의 `Symbol` 로 OPEN 을 끝까지
   읽어 조건주문이 **0건**이다. 옛 id 가 정정으로 사라졌다면 후속은 같은 심볼의 OPEN 에 있다(정정은
   심볼을 바꿀 수 없다 — 정정 바디 `ConditionalModifyBody` 의 필드는 type·quantity·orderType·expireDate·
   first·second·confirmHighValueOrder 뿐이고 심볼이 없다, `internal/official/conditional_writes.go:52-60`). 다른 조건주문이 하나라도 있으면 그것이 후속인지 무관한 것인지 가를 수 없으므로 거절한다
   `[정책 — Q4]`.
2. **그 심볼의 CLOSED 그룹에 대상 id 가 없다.** 대상 id 가 CLOSED 에 있으면 그것은 부재가 아니라
   종결 상태(`COMPLETED`/`EXPIRED`)다 — 이 change 는 그 행을 `reconciled-absent` 로 쓰지 않고
   거절한다(발동이면 child 가 있다).
3. **발동으로 사라진 경우를 배제한다** `[비움 — Q1]`. 발동한 조건주문이 CLOSED 목록에
   `COMPLETED`+`triggeredOrderId` 로 **남는지는 측정되지 않았다**(발동 측정은 deferred —
   verify-execution-capability 2.5). 남는다는 측정이 있으면 2 가 이것을 덮는다. 없으면 대안은
   verify-observes-the-trigger 가 이미 채택한 근거 — **보유 수량** — 뿐인데, 그러려면 기록에 그
   artifact 이후의 보유 수량 기준값이 있어야 한다. 어느 쪽인지 정해지기 전에는 이 조건을 만족시킬 수
   없으므로 명령은 거절한다.
4. **읽기 집합이 흔들리지 않았다** — 컷 토큰이 없으므로 같은 읽기 집합(OPEN 전 페이지 + CLOSED 전
   페이지)을 **두 번** 연속 수행하고, 두 결과의 (그룹, id, status) 집합이 같아야 한다. 첫 읽기 시작부터
   둘째 읽기 끝까지의 경과가 한도 `[비움 — Q3]` 를 넘으면 거절한다. 한 번이라도 페이지 오류·반복 커서·
   빈 커서·상한 도달(m0RecoverPending B4·B6·B13·B14 와 같은 모양)이면 거절한다.

이 규칙은 G1 을 **증명**하지 않는다 — 서버가 일관된 컷을 주지 않는 한 증명은 불가능하다. 대신 "읽기
사이에 무엇이 바뀌었으면 두 번의 읽기가 다르다" 와 "살아 있는 것이 그 심볼에 하나도 없다" 로 창을
좁히고, 좁힐 수 없는 모양은 전부 거절 쪽으로 보낸다. 리뷰가 요구한 fail-closed 대안이 이것이다.

### G2 — 사건의 신원: 기록이 이미 쓰는 키를 그대로 쓴다

**코드가 주는 것.**

- 기록은 artifact 를 이미 **원문 id** 로 담는다 — `Artifact.ID` 주석 "Printed in full on purpose: an
  operator who has to go and cancel something by hand needs it"(`record.go:185-187`). M0 체크포인트
  줄도 "may retain exact broker identifiers"(`record.go:85-88`). 계좌는 반대로 마스킹된 참조만
  담는다(`Entry.AccountRef`, `record.go:265-266`).
- 투영의 키는 `Kind + "\x00" + ID` 다(AST `verifylive--outstandinglines` B3, `record.go:548`).
  종결은 단조다 — terminal 줄 뒤의 비-terminal 줄은 되살리지 않는다(B4 `:552`). 투영은 비-terminal 만
  낸다(B6 `:562`). 그리고 종결 술어는 한 곳이다 — `Artifact.terminal()`(`record.go:575`)이고 그 주석이
  "a third ending added later is honoured by all of them at once" 라고 이 확장을 예정해 두었다.
- 줄의 종류는 이미 여럿이다(`KindStep`·`KindApproval`·`KindCleanup`·`KindM0Checkpoint`,
  `record.go:62-89`). `decodeEntry` 는 `FormatVersion > 1` 만 거절한다(AST B2 `:442`) — 모르는 `Kind`
  나 필드는 거절하지 않는다(`json.Unmarshal`, B1).
- 비-원문 태그의 선례가 있다 — `CausalReceipt.tag(id) = sha256(runID + "\x00" + id)`
  (`internal/verifylive/receipt.go:353-361`). 기록 지문은 `Digest`(16바이트 sha256,
  `record.go:583-595`).

**결정(골격).**

- 새 줄 종류 `KindReconcile = "reconcile"` 한 줄이 대상 artifact 하나를 담는다. 그 artifact 의 신원 필드
  (`Kind`·`ID`·`Symbol`·`ChainID`)는 **대사할 outstanding 줄의 값 그대로**이고, 새 종결 필드
  `ReconciledAbsent bool` · `ReconciledAt time.Time` 을 세운다. `Artifact.terminal()` 이 그것을 셋째
  종결로 포함한다 — 그러면 `outstandingLines` 를 쓰는 소비자 전부(`PendingCleanup`
  `cleanup.go:119-135`, `liveCount` `mutate.go:679`, report·status·abort·redo)가 한 번에 따른다.
- **멱등**: 대상이 이미 종결이면 `outstandingLines` 가 그것을 내지 않으므로 선택 단계에서 "대사할 것
  없음" 으로 거절한다 — 둘째 줄은 쓰이지 않는다.
- **선택**: 호출자는 id 를 주지 않는다. 명령은 기록의 `Outstanding` 중 `Kind == "conditional-order"` 인
  것을 세고, **정확히 하나가 아니면** 거절한다(리뷰: "reject zero or multiple eligible outstanding
  conditionals"). `HeldUntil` 이 걸린 artifact 는 후보가 아니다(보유 게이트는 cleanup 의 규칙이고
  대사가 그것을 우회해서는 안 된다).
- **근거 지문**: 줄의 관측 `reconcile.basis` 에 두 번의 읽기 집합을 버전·도메인 태그
  (`"a121/reconcile-basis/v1"`)와 함께 `Digest` 로 적는다. 페이지 원문은 적지 않는다.
- **구 바이너리**: `FormatVersion` 은 1 그대로다(필드 추가). 구 바이너리는 `reconciled_absent` 를 모르는
  필드로 버리므로 그 줄을 비-terminal 줄로 읽는다 — `outstandingLines` B4 는 terminal 뒤의 비-terminal
  만 막으므로 **구 바이너리에서는 그 artifact 가 여전히 outstanding** 이다. 틀리는 방향이 "다시 보인다"
  이므로 안전 쪽이고, 되돌림은 증거를 지우지 않는다.
- **원문 id 경계** `[비움 — Q2]`. 위 골격은 대사 줄에 **새 원문 id 를 하나도 더하지 않는다** — 같은
  기록이 이미 담은 id 를 같은 키로 반복할 뿐이다. 그러나 현재 delta spec 은 "without retaining raw …
  broker identifiers" 라고 쓰고 리뷰는 "a raw ID violates the stated boundary" 라고 읽었다. 둘 중 하나를
  골라야 한다 — (a) spec 문구를 "대사 줄은 대상 artifact 줄이 이미 담은 식별자 외에 새 브로커
  식별자를 더하지 않고, 읽기 근거는 지문으로만 남긴다" 로 고친다, (b) 대사 줄은 `tag` 선례처럼
  sha256(도메인·기록 run·id) 태그만 담고 `outstandingLines` 가 각 artifact 의 태그를 계산해 대조한다.
  (b) 는 투영의 키 규칙(B3)을 바꾸므로 그 함수가 편집 대상이 된다.

### G3 — 계좌·프로필·시장의 같음: 실행 가능한 셋

**코드가 주는 것.**

- 현재 계좌 참조는 `resolveVerifyAccount` 가 정한다 — `DisplayName` 이 빈 값이 아닌 **첫** 계좌의
  `DisplayName` 이고(AST B5·B6 `verify.go:941-945`), 그 계좌 id 가 숫자가 아니면 seq 는 0 이다(B7
  `:947`). seq 가 0 이면 `buildVerifyBroker` 는 헤더 계좌를 클라이언트의 지연 해석에 맡긴다(AST B5
  `:894`). 그 주석이 두 값을 한 항목에서 가져오는 이유를 적는다: "taking them from different accounts
  would produce a record that names one account and measured another"(`:911-914`).
- 기록의 계좌는 줄마다 `AccountRef`(마스킹)다(`record.go:265-266`).
- 기록 경로는 `resolveVerifyRecordFor` 가 정한다 — `--record` override 가 이기고(AST B1 `:765`),
  아니면 `--config-dir` 프로필 아래(B2 `:769`), 아니면 데이터 디렉터리다. 시장마다 파일이 다르다
  (`RecordFileName`, `record.go:55-60`).
- 원시 행은 브로커가 준 `Market`("KR"/"US")을 그대로 담는다(`conditional_reads.go:108-110`).

**결정.** 공식 읽기와 로컬 추가 **전에** 아래가 모두 참이어야 한다. 하나라도 아니면 거절한다.

1. **계좌**: 대상 artifact 를 언급한 모든 기록 줄의 `AccountRef` 가 서로 같고, 현재
   `resolveVerifyAccount` 의 참조와 문자열로 같다(양쪽 `TrimSpace`). 비어 있거나 섞였으면 거절.
2. **헤더 계좌**: seq 가 0 이면 거절한다 — 지연 해석은 참조를 준 그 항목을 고른다는 보장을 이 경로에서
   재확인할 수 없다(`buildVerifyBroker` B5). 계좌 목록에 `DisplayName` 이 빈 값이 아닌 항목이 둘 이상이면
   `[정책 — Q5]`.
3. **프로필**: `--record` override 를 받지 않는다(`resolveVerifyRecordFor` B1 을 닫는다). 기록은 자격
   증명과 **같은 root**(`--config-dir` 또는 기본)에서 유도된 경로여야 한다 — 그래서 프로필 결속은 구성으로
   성립한다.
4. **시장**: `--market` 을 필수로 받고 기록 파일은 `RecordFileName(market)` 이어야 한다. 읽은 행 중
   대상 심볼의 행은 모두 `Market == market` 이어야 하고(G1 의 1·2 가 읽는 행 전부), 아니면 거절한다.

### 열린 질문 — 코드에 근거가 없어 사람이 정할 것

- **Q1 — 발동으로 사라진 조건주문의 배제.** 발동한 조건주문이 `CLOSED` 목록에 `COMPLETED` 로 남는지는
  측정되지 않았다(발동 측정 deferred). (a) 그 측정을 선행 조건으로 삼는다, (b) 기록에 artifact 이후의
  보유 수량 기준값이 있을 때만 보유 수량 대조로 배제한다(없으면 거절), (c) 이 change 에서는 G1-3 을
  만족시킬 수 없음을 받아들이고 명령은 늘 거절한다(대사 불가 = a063 이 계속 막힘). 어느 것인가.
- **Q2 — 대사 줄의 신원 표현.** G2 의 (a) spec 경계 문구 수정 + 기존 키 재사용, (b) 도메인 태그. 어느 것인가.
- **Q3 — 두 번 읽기의 경과 한도.** 코드에 이 용도의 상수가 없다. 값(초)과 초과 시 거절을 정한다.
- **Q4 — 같은 심볼의 다른 조건주문.** 기본은 거절이다. 엔진·수동 보호 조건주문이 같은 심볼에 상주하는
  계좌에서는 대사가 영영 불가능해진다 — 그 비용을 받아들이는가, 아니면 기록이 소유를 증명할 수 있는
  예외(예: `ChainID` 로 연결된 후속)를 둘 것인가.
- **Q5 — 계좌가 여럿인 자격 증명.** `resolveVerifyAccount` 는 첫 항목을 고른다. 대사에서는 후보가 둘
  이상이면 거절할 것인가.
