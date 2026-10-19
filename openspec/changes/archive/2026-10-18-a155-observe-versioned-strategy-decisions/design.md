## Context

우선순위 8/9: 버전 전략 관측 전용 연결. 선행 change: a149, a150, a153, a154.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 오프라인 계산이 런타임 입력에서도 재현되는지 관측할 필요가 있다. 관측 결과를 주문·예약·운영 제어에서 분리한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

관측은 order-capable strategy entry path의 paper/shadow mode가 아니다. 주문 권한 없는 별도 consumer로 typed diagnostic records만 만든다. broker client, reservation writer, position mutation을 dependency로 받지 않으며 runtime coordinator에 EntryDecision을 제출하지 않는다. 과거 정본의 LIVE-only 실행 경로를 변경하지 않는다.
연구 record는 parent evidence ID, policy/calculation version, source/as_of, ready/refusal, entry/stop/target geometry, planned risk, optional outcome linkage를 가진다. append-only output을 사용한다. 기본 queue=256 records, encoded record max=64KiB, queue-full/oversize는 drop-newest와 이유별 counter로 보고한다. 생산자 enqueue는 nonblocking이며 observation disk I/O·retry를 trading/exit thread에서 수행하지 않는다. record ID는 입력 digest와 policy version으로 결정한다. 단일 writer가 full newline record append와 fsync 완료 후 checkpoint를 atomic replace한다. crash 후 마지막 불완전 record는 신규 파일/오류 상태로 격리하고 기존 evidence를 truncate하지 않는다. checkpoint보다 앞서 durable append된 동일 ID는 재생 시 duplicate skip하며, checkpoint가 durable record를 앞서면 복구를 중단해 명시적 오류를 반환한다. 기록 위치는 기존 증거 경로와 분리한다. 기록 실패/부하는 기존 protective exit를 지연시키지 않는다.
이번 범위에서 구현·검증은 fixtures와 offline replay까지다. 실제 관측 시작은 관측 market/session/window, 수집 API budget, 저장 위치·보존 기간을 적은 별도 사용자 승인 및 frozen manifest가 필요하다. 현재 승인으로 collector를 시작하거나 새 scheduler/컨테이너를 재시작하지 않는다. a128/r1/a147 경로에 쓰지 않는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Separate observation adapter consuming immutable market inputs; no gateway/reservation writer capability. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High isolation boundary; observation must have no order authority. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.

## Implementation contract — Manager refinement, 2026-10-18

Scope remains offline fixtures/replay only. a149/a150/a153/a154 are archived; a154 delivers a hypothetical calculator, not an approved policy. No runtime producer, collector, engine hook, scheduler, container or operating binding is added. a157 is excluded and owned by Claude in another checkout. UI is unaffected: impeccable: N/A — offline Go component and replay utility only. Existing source/evidence is reused read-only.

### Pure diagnostic input

Use a new independent `internal/strategyobserve` package and a small offline replay utility under `tools/`. Reuse `strategymarket.CalculateSessionFeatures` unchanged for session features. Its KR-only semantics remain: US inputs produce explicit unsupported diagnostics, never invented US Parker support. The writer itself is market-neutral. Reuse the pure a148 `ParkerRRFromDecision` only through a narrow allowlisted projection for its existing lane/version; do not construct a sealed order intent or call an evaluator/coordinator with execution authority. Do not import account journal, broker, reservation or operating stores. Realized return calculation is out of scope because no fills are supplied.

Input contains parent evidence ID, research policy version, session feature input and optional planned geometry (entry/stop/target, whole-share quantity, existing Parker lane ID/version). Public identifiers only; no credentials/account fields or unrestricted metadata map. Versions, source and as-of are retained with caller-supplied/unverified provenance. Planned price risk, when geometry and quantity are valid, is exact (entry-stop)*quantity in quote currency; costs and actual fill risk remain unknown. Missing geometry must not hide computable features. Readiness is diagnostic completeness, not entry eligibility or performance approval. Future outcome linkage is a separate new record and never an edit of an existing diagnostic.

Canonical JSON of the typed, bounded input plus schema/calculation/policy versions defines a SHA-256 stable ID. Include market/session/as-of and all geometry in the digest; canonical UTC times and explicitly fixed serialization prevent retry identity drift. Do not hash current wall-clock time, queue counters or output location. Records contain the input digest, not full raw market bars. Same ID with different canonical payload is an error, never silent deduplication. Raw input bytes remain untouched.

### Bounded producer and writer

Construction/calculation is outside the enqueue operation; no trading-thread integration exists in this release. Enqueue accepts an owned immutable encoded copy, capped at 65536 bytes including its final newline. Queue capacity is exactly 256. Nonblocking select drops newest when full; oversize, invalid, queue-full, closed and writer-failed outcomes have observable counters. No filesystem work or retry in enqueue; it must not share a mutex held across writer I/O. Test a deliberately blocked writer while producer and an independent synthetic exit task continue. This establishes architectural isolation, not measured production exit latency.

Close stops accepting, drains already accepted records and returns writer error. Use explicit synchronization to prevent send-on-closed and races; do not close a channel while producers can send. On any append/flush/checkpoint failure, latch failure, stop further append, expose failed/pending counts and return an error. Do not retry uncertain writes within the same writer. Cancellation does not claim queued records durable. Only one consumer owns file I/O.

### Durable replay and filesystem ownership

Use a dedicated explicitly supplied directory, never default to operating or existing evidence paths. First creation claims a new empty destination; reopen requires the component's own format marker. Reject symlinked destination/ledger/checkpoint/lock and foreign files, and explicitly refuse existing r1/a128/a147 evidence locations. Do not modify any protected evidence while validating rejection. Use exclusive nonblocking kernel locking with supported-platform implementation; unsupported platforms fail explicitly. Do not reuse an engine's own lock namespace or alter its locking policy.

Append a complete JSON line, check short writes, fsync the ledger, then replace a temporary checkpoint after temp-file fsync and sync its parent directory. Sync newly created ledger/marker directory entries before acknowledging durability. Checkpoint binds schema, sequence count, last record ID, byte offset and ledger prefix digest. Recovery validates the complete ledger and IDs against the checkpoint, refuses missing/foreign/ahead/mismatched checkpoint references, and skips identical durable records after a lagging checkpoint. An absent checkpoint is valid only for an empty owned ledger or an explicitly recoverable append-before-first-checkpoint crash; the ledger must independently validate. Identical records replayed in-process also skip.

Any unterminated final line (even valid JSON without newline), invalid record, conflicting ID or corrupt checkpoint yields an explicit failed recovery with the original bytes preserved. Choose the existing design's error-state quarantine option, not automatic truncation or new-segment rotation. Do not append after ambiguous recovery. Existing checkpoint temp residue may be validated/ignored only within the owned directory; do not delete arbitrary files. Recovery memory is bounded by a documented offline input/ledger limit with a refusal at the limit; no unbounded scan/map is acceptable. This is a local filesystem durability contract, not a guarantee for arbitrary network filesystems.

### Replay and acceptance

The CLI accepts local bounded JSON/JSONL fixtures and explicit output directory only; no networking or deployment dependencies. Reject unknown fields, malformed/oversized input and protected output paths before writing. Record accepted/durable/duplicate/dropped/failed separately; nonzero exit for invalid input, recovery/storage failure or dropped records, never a misleading successful complete replay. Printed results identify offline/unverified status and output location. Tests use disposable paths exclusively.

Required tests: independent numerical oracle; feature semantics/US unsupported; immutable caller payload; version/identity replay; duplicate same/different payload; queue-full and oversize exact boundaries; blocked/failing writer; concurrent enqueue/close race; short write; append-fsync/checkpoint/rename/dir-sync failure and restart at each boundary; lagging/ahead/corrupt checkpoint; torn tail; multi-writer and symlink/foreign directory rejection; protected-path contents unchanged; no broker/reservation/journal/coordinator authority or network/runtime wiring. Actual CLI output must be inspected. Scope can complete with fixture evidence; actual observations remain separately approved and unstarted.

No new dependency is expected: standard library exact arithmetic/serialization and existing platform lock patterns suffice. Source research and AST evidence are recorded before implementation. Manager reviews this refinement; independent and gstack design checks precede Terra code. Final lifecycle order is actual full gate → formal archive/spec synchronization → PM/strict checks → scoped commit → normal origin/main push. a156 and operationally blocked changes remain active until their own acceptance criteria are met.

Bounds selected for v1: JSONL input lines at most 1 MiB, at most 65536 input records per invocation; ledger at most 64 MiB and 65536 durable records. Record limit includes newline. Over-limit input/recovery is refused without truncation; prospective append limit failure is visible and non-success. Quantity is an optional unsigned positive whole number <=1000000; geometry is bounded decimal (32 characters, <=8 fractional digits) and risk uses exact rational strings, never a rounded display as input. Metadata strings are length-bounded and reject control characters; no unrestricted user paths or account fields are embedded in diagnostic rows. Use local Linux/Unix filesystem semantics only where actual exclusive locking is supported, with explicit unsupported errors elsewhere. Existing evidence is not migrated.

Geometry identity: derive quote currency from market (KR/KRW, US/USD) and reject any mismatching input currency. Require an explicit `unadjusted` geometry basis when geometry is supplied, and reject geometry if any supplied bar asserts another price basis. Empty or not-ready features do not alone block independently valid hypothetical geometry; all geometry remains caller-supplied/unverified. The Parker RR adapter is allowed only for KR/KRW with the existing Parker lane/version and unadjusted basis. For US, session features and Parker RR are explicitly unsupported; independently valid USD price-risk arithmetic may still be reported as hypothetical and does not imply a US Parker strategy.

Recovery integrity refinement: each line is an envelope containing stable input-based record ID, canonical diagnostic payload and its SHA-256 digest. Recovery recomputes the payload digest even for records beyond a lagging checkpoint; malformed envelopes, digest mismatch and conflicting ID/payload fail closed. This detects accidental corruption, not authenticated provenance or malicious rewriting by the filesystem owner.

Implementation-review clarification: when bars are supplied alongside geometry, every bar must explicitly declare `unadjusted`; an empty basis is not evidence of compatibility. No supplied bars remains permissible for independent hypothetical geometry. Quantity omission leaves price risk unknown while preserving computable features and RR; an explicitly supplied zero or out-of-range quantity is invalid. Validate both feature and geometry currencies against market. A durable record is exactly one canonical physical JSON line followed by one newline, so append and recovery accept the same format. Opening owned storage entries must not block on a FIFO or other non-regular file before rejecting its type.

Aggregate-gate compatibility refinement: the existing a047 import guard also rejects this pure RR reuse. Preserve its production isolation purpose: allow only the exact observation leaf and only `ParkerRRFromDecision`, `ParkerRRDiagnostics`, `DecisionRecord`, `LaneID`, and `LaneVersion` references, resolving import aliases and rejecting bypass forms. Add the inverse guard so production imports of the observation package remain limited to the exact offline CLI. Negative fixtures must demonstrate that forbidden engine APIs and runtime observation imports still fail. This is a test-contract refinement, not runtime wiring or permission to expand the engine API surface. Preserve the original failed gate receipt and repeat independent/gstack review and the aggregate gate after this change.

Filesystem refinement: v1 supports local Linux filesystems; other platforms return unsupported rather than silently omit locking/durability. Use existing x/sys/unix primitives, no new dependency. Traverse absolute ancestors with directory-relative no-follow opens, lock the destination directory inode nonblocking (avoid replaceable lock-file identity), and operate on owned children relative to that descriptor. Require regular non-symlink single-link ledger/marker/checkpoint files; hardlinked files are refused before write. Refuse output under any `openspec` evidence tree or reserved r1/a128/a147 evidence component, including resolved aliases. Missing destinations may be created only beneath an existing validated parent. Do not create intermediate parent trees. Reopen verifies marker and directory contents under the lock before append. A hostile privileged filesystem owner replacing ancestors is outside this offline tool's threat model; never claim hashes/locking provide authenticated evidence against such an attacker. No new platform abstraction framework is warranted.
