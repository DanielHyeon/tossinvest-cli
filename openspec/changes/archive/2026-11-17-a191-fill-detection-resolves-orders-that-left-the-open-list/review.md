# Review

## 1단계 기록 (구현 팀메이트 Opus, 2026-11-16)

### 기억 회고

- 조회함, 해당 없음 — ai-memory·Hindsight 모두 이 결함(목록 이탈 + 단건 PENDING)에 대한 기존 결정·함정 없음(Manager 조회, 세션 기록만 걸림).
- 파일 기억 중 적용한 것(현재 HEAD 로 재확인): `fail-closed-must-name-what-it-rejects`(design Q3 의 N1~N10 표), `a-new-guard-unpins-the-guards-behind-it`
  (design §5 가드 위치), `two-judgements-cover-for-each-other`(D9 — a167 이 판정을 새로 만들지 말 것), `missing-tool-reports-clean`(`codegraph affected`
  기본값 불신), `flm-before-claiming-not-before-editing`(인용 함수 19개 AST 를 문서보다 먼저), `tossos-shared-openapi-credential-token-war`(Q9 의
  `--config-dir`).

### 번호 배정

- 확인 범위: `git branch -a --list '*a191*'`(이 브랜치만), `git worktree list` 59개 중 a191 은 이 worktree 하나, `git log --all --oneline -5000`
  제목의 `a191` 0건, `~/tossos-work`(이 worktree 만), `openspec/changes`·`archive`(a191 0), `docs/pm/portfolio`(Story a190 까지).
- a191 을 씀(Manager 배정과 같음). 이름: `a191-fill-detection-resolves-orders-that-left-the-open-list`.

### Story·Feature

- `STORY-TOS-a191` → `FEAT-TOS-002`(Broker execution verification and measurement).
- 근거: 결함의 원인이 브로커 실측(M31·M32)과 실행 기반(harden-execution-base 3.1 의 체결 감지)에 있고, 같은 Feature 에 a165·a166·a171(실행 경로)이
  있다. 대안 FEAT-TOS-003(Engine interlocks)은 대사 차단이 결과로 걸리지만 고치는 대상이 인터록이 아니라 고르지 않음. **Manager 확인 대상.**

### 증거

- CodeGraph 1.6.0: `analysis/code-context/`(worktree 색인, 원문 `codegraph-callers.txt`·`codegraph-affected.txt`).
- CodeGraphContext·GBrain: not-applicable(사유 `codegraphcontext-context.md` — 공유 그래프 DB 쓰기 회피, worktree GBrain 홈 없음).
- AST: 편집 대상 `analysis/function-logic/internal-filldetect--detector.collect/`(ast·FLM·BTM·risk), 인용 19개 `analysis/cited-ast/`.
  손으로 읽은 것은 `analysis/cited-ast/README.md` 표와 design 본문에 표시.
- 영수증: `verify-execution-capability/measurements.md` M7·M8·M10·M25·M28·M31·M32, `analysis/retry-matrix-slo-2026-10-22.md`, openapi.

### Pre-Edit 선언 (확정 2026-11-16, Manager 결정 반영)

```text
Pre-Edit Gate:
- change id / task id: a191-fill-detection-resolves-orders-that-left-the-open-list / 3.13
- 대상 심볼(패키지.함수): filldetect.Detector.collect(호출 2줄), filldetect.JournalTracked.TrackedOrders(필드 1줄),
  cmd/tossctl.engineFillDetector(필드 1줄). 새 파일 internal/filldetect/closed_lookup.go, cmd/tossctl/a191_fill_alerts.go(새 함수만).
  타입 필드: Detector.Alerts·leftOpen·leftOpenCycle, TrackedOrder.IntentID, Cycle.ClosedReads·Unresolved.
- CodeGraph definition/callers/callees/impact: collect 호출자 pollLocked 1; JournalTracked.TrackedOrders 는 TrackedSource 구현(호출자 collect
  call@370 경유); engineFillDetector 호출자 engineRuntime 1 + 시험 2; affected(*_test.go) 깊이 1: 72, 2: 563, 5: 1336 (codegraph 1.6.0).
- CodeGraphContext 후보와 evidence reconciliation: CGC·GBrain not-applicable, 불일치 0 (evidence-reconciliation.md).
- 기존 동작 파악 근거: collect AST(분기 19), cited AST 19개, 기존 시험 detect_test 18·ledger_test·a087 핀, measurements M31·M32, openapi.
- Function Logic Map / Branch Test Map: analysis/function-logic/ 의 collect 번들 + 편집 전 AST 를 남긴 뒤 ledger·engineFillDetector 번들을 만든다.
- upstream 상속 테스트 영향: no — filldetect 는 TossOS 전용(eb58ba489), 엔진 루프에서만 돈다.
- 실패 테스트 선행 작성: yes (tasks 3.x RED, 실패 원문은 analysis/red/)
- 설정·DB·journal 변경과 rollback: 없음 — 스키마·설정·토글 무변경. rollback = 이전 이미지.
- 안전 불변식 §0 위반 여부 검토:
  §0.3 손절 즉시성 — 조건부 통과: exit 루프 경로에 직접 호출 0, 간접 지연은 M8(조회 전체 2s·주기당 1건)로 묶음 → tasks 4.6 측정으로 확정.
  §0.4 예산 — D10 + tasks 4.4(retry matrix 정본 표, a165·a166 선례대로 아카이브 표를 같은 변경에서 고침).
  §0.2 토글 OFF — 해당 없음(D8). §0.9 — 손절·익절·사이징 로직 무변경. §0.1 — 실계좌 호출 0(가짜 pager·httptest 없음).
```

### 리뷰 보이스 계획

- 작성자: Claude(Opus). 바깥 보이스: Codex(다른 계열), 읽기 전용 샌드박스, 적대적 Eng 관점 필수(High-risk). 회차 상한 2.
- 리뷰가 제안한 새 가드·새 정책은 설계에 넣지 않고 아래 「Manager 결정 필요」에 적는다.

### 리뷰 1회차 (2026-11-16, Codex gpt-6-astra, `codex exec -s read-only`, 적대적 Eng)

원문 `analysis/review/review-1-codex.md`. 판정: 「이대로 구현하지 말 것」. 12건. 내가 코드로 확인한 뒤의 처분:

| # | 등급 | 요지 | 확인 | 처분 |
|---|---|---|---|---|
| 1 | BLOCKER | `execution` 없음/null → 체결 0 으로 읽혀 수량 일치처럼 보임 | 맞음(payload.go:87, `parseDecimal` 178~184) | 사실은 design D3-9 에 적음. 「행에 execution 존재 요구」는 새 가드 → **M6** |
| 2 | BLOCKER | `ScanOrders` 가 `HasNext=true`+빈 커서를 끝으로 침 | 맞음(B6 if@755) | 사실은 D3-8 에 적음. 거절은 새 가드 → **M7** |
| 3 | MAJOR | 「정확히 하나」와 중복 처리가 어긋남 | 맞음 — 내 설계 결함 | **수정**: raw 바이트가 전부 같을 때만 하나로 셈, 같은 키·다른 payload 는 ambiguous(D3-2). 시험 3.7b·변이 m20 |
| 4 | BLOCKER | delta 는 원장 직전값 대비 → 수용 시 CLOSED 가격으로 체결 기록 가능, Q5 「delta 0」 틀림 | 맞음(fills.go:379~382·412~415) — 내 설계 결함 | **수정**: 상태 필드만 교체, 수량·가격·금액·체결 시각은 단건 유지(D3-7). Q4·Q5 정정. 시험 3.4b·변이 m19 |
| 5 | BLOCKER | 조회가 적용 전에 돌아 다른 주문(보호 매도)의 반영을 늦춤, 요청 시한 15s | 맞음(`pollLocked` call@293 → B5, client.go:22·137) | D11 사실 정정. 막는 방법은 새 가드 → **M8**(2단계 착수 전 필수) |
| 6 | MAJOR | CLOSED 실패에도 freshness 도장 = 불완전한 그림 도장 | 부분 동의 — OPEN 은 완주, 그림은 오늘보다 덜 완전하지 않음. 「미해결 추적 주문이 있는 채로 신선」은 오늘도 같음 | 정책 → **M9** |
| 7 | MAJOR | 페이지를 안 세 예산 과소 계상, 총량 상한·질의 재사용·재시작 동시 조회 없음 | 맞음 | 수치 **수정**(D10: 주문당 ≈0.1 req/s). 상한·재사용 → **M8** |
| 8 | MAJOR | 「보유 대사가 잘못된 종결을 잡는다」는 틀림 | 맞음(`Compare` B11·B21·B22, `BlocksEntry`) — 내 주장 결함 | **수정**: Q3 통과 문단 정정. 추가 안전망 → **M10** |
| 9 | MAJOR | 정정 타이밍의 기존 즉시 래치는 30s 보고와 무관하게 남음 | 맞음 — 오늘 동작 | N4 행 정정(사실). 「일관된 미해결」 대 「모순 파생」 계약 분리는 범위 밖 후속 |
| 10 | MAJOR | legacy 빈 범위 행이면 복구 약속이 거짓, 수렴 크레딧 없음 | 맞음(converge.go:142~145) | proposal·Q5(d) **조건부**로 정정, Q5(c) 손 읽기 확정. legacy 픽스처·복구 절차는 Manager 판단(tasks 5.3) |
| 11 | MINOR | 기존 가짜 pager 가 status 를 안 봐 우발 CLOSED 호출을 못 잡음 | 맞음 | 새 시험이 질의 순서 단언(tasks 3.7c). 기존 가짜는 안 고침(시험 함수 편집 = gate ⑤) |
| 12 | NIT | `B45(if@475)` → `case` | 맞음 | **수정** |

리뷰가 제안한 새 가드·정책(#1·#2·#5·#6·#7·#8)은 설계에 넣지 않았다 — 아래 M6~M10.

### 리뷰 2회차·최종 (2026-11-16, Codex gpt-6-astra, 같은 방식)

원문 `analysis/review/review-2-codex.md`. 판정: 「문서 묶음 그대로는 착수 불가 — spec·착수 게이트·M8/M9 설명 정합화 필요」. 1회차 수정 중
#1·#2 사실, #4(D3-7 필드 분리 — downstream 이 전부 `Derived` 에서 읽음, 빠진 필드 없음, 단건 체결 시각으로 새 refusal 없음), #5 사실, #7·#8·#9·#12 는
**맞음**, #3·#10 은 **부분 완료**로 판정. 새 지적 6건 — 전부 내 문서의 정합성 문제라 반영함:

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| 1 | MAJOR | spec 이 옛 D3(「정확히 하나」·「행이 스냅숏을 대신」)를 그대로 명령 | **수정**: spec R1 에 raw 바이트 중복 규칙·상태 필드만 교체·수량 등 단건 유지, ambiguous 시나리오 확장, design 사유 표 정정 |
| 2 | MAJOR | spec 의 「래치 없음」·「열린 주문에서 빠진다」가 기존 래치·legacy 예외와 충돌 | **수정**: 「종결 목록 조회가 더하는」 기록·래치로 한정, 성공 시나리오·Q5(b) 에 legacy 전제 |
| 3 | MAJOR | M8 (나) 「Apply 뒤로」는 공유 wire 점유를 못 막음, (가) 시한 단위 불명 | **수정**: M8 선택지 설명 보강(시한은 조회 전체 단위, (나) 의 잔여 위험 명시, 수치 결정 필수) |
| 4 | MAJOR | M9 추천 근거「대사가 막는다」는 보장 아님(별도 루프·심볼 범위) | **수정**: 「잔여 위험 수용」으로 다시 씀 |
| 5 | MAJOR | M6~M10 이 착수 게이트에 없고 Pre-Edit 이 손절 즉시성 「통과」 선언 | **수정**: tasks 머리말·1.8 에 M1~M10(M8 필수), Pre-Edit §0 판정을 「미확정」, tasks 4.6(M8 검증) 추가 |
| 6 | MINOR | 「위험 버킷은 종결 해제만」은 delta>0 BUY 에서 틀림 | **수정**: delta=0/delta>0 로 나눠 적음(리뷰 인용 줄 `risk_bucket_fill.go:208~243` 은 손으로 확인 안 함) |

리뷰 회차 상한(2)에 닿아 3회차는 하지 않음. 2회차 수정분은 **재리뷰받지 않았다** — Manager 검증 대상.

## Manager 결정 (2026-11-16, 1단계 검증 통과 뒤)

| # | 결정 | 반영 |
|---|---|---|
| M1 | A — `Detector.Alerts` ← `ectx.Notifier`, `fill.order_unresolved` normal. 브로커 주문번호 안 실음, 내부 intent id 로 식별 | design D6. intent id 는 a129 화이트리스트에 있는 필드 종류(`intent_id`). `TrackedOrder` 가 안 들고 있어 `JournalTracked.TrackedOrders` 에 1줄(번들 1 추가). spec R4 문장 고침 |
| M2 | 제외 — 발동은 `!Terminal && !FailClosed` | design D1·N4, spec R1, 시험 `TestA191UnknownSingleReadIsNeverLookedUp`·변이 m21. `ReplacedWithLineage` 시험 삭제 |
| M3 | 30s/30s | D5·D6 |
| M4 | (가) | D3-6 그대로 |
| M5 | FEAT-TOS-002 | Story 그대로 |
| M6 | (가) 요구 — `closed_list_execution_missing` | D3-9, spec R2, 한 함수 `closedRowCarriesFill`·시험·변이 m22 로 격리. **Q9 실측(2026-11-17) 뒤 (가) 유지** — 체결 없이 취소된 실제 행이 `execution.filledQuantity` 를 문자열로 실어 보냄(design Q9 결과) |
| M7 | 불채택 | D12 잔여 위험 |
| M8 | (가) 축소판 — 조회 전체 2s(context 는 끊기 용도만, 시각 판정은 주입 Clock), 주기당 1건(안 물은 것 → lastLookup 오래된 순, deferred 는 lastLookup 갱신·삭제 안 함), 질의 재사용 없음, `pollLocked` 무편집 | D5·D10·D11, spec R1, 시험·변이 m23·m24·m26, tasks 4.6 |
| M9 | 도장 유지, 잔여 위험 | D12 |
| M10 | 잔여 위험, 재등장 경보 후속 | D12 |
| 장부 | Manager 읽기 전용 확인: scoped·legacy 둘 다 범위 채움 → Q5 legacy 예외 해당 없음 | design Q5(a) |

### 요구 변경분 리뷰 (3회차 — 결정 반영분 한정, 2026-11-16, Codex gpt-6-astra)

원문 `analysis/review/review-3-requirement-change-codex.md`. 판정 「착수 승인에 부족 — 시간 상한·공정성·통지 보장」. M2·M6 반영과 알림 필드(금지 필드
자동 추가 경로 없음, normal 은 dedup Key 안 씀)는 일치로 확인됨. 지적 7건 — 내가 코드로 확인함:

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| 1 | BLOCKER | 2s context 는 반환·wire 해제 상한이 아님(토큰 잠금 token.go:62~80, 예산 정리 `WithoutCancel` +2s transport_linux.go:405~407, 예산 파일 Sync) | 맞음. **문장 정정**(D11: 「취소 요청」, 정상 경로 2s). 상한을 더 강하게 만드는 것은 새 가드 → **M11** |
| 2 | MAJOR | collect 안 Notify 는 동기 로그(log.go:194~203)라 2s 로 안 묶임 | 맞음. 사실 정정(D6). 알림을 임계 경로 밖으로 빼는 것은 새 구조 → **M12** |
| 3 | MAJOR | 「안 물은 것 먼저」는 새 후보가 계속 들어오면 재조회 starvation | 맞음(추적 순서 = 주문번호 정렬, fills.go:1652). spec 시나리오를 고정 집합 전제로 정정. 공정 규칙은 새 정책 → **M13** |
| 4 | MAJOR | 「운영자에게 알린다」와 normal best-effort 는 다른 계약 | 맞음 — M1 결정(normal)과 문장을 맞춤: spec R4 를 「기록·전송 1회 시도, 전달 보장 없음」. 알림 context 는 조회와 별개 2s |
| 5 | MINOR | 「한 건」→「최대 한 건」, R4 를 D1 후보로 한정 | **수정** |
| 6 | MINOR | D10 수치는 인증 재시도 없는 정상 경로 수치 | **수정** |
| 7 | MINOR | M7 잔여 위험에 정상 종결 누락·지속 미해결도 포함 | **수정** |

## Manager 결정 필요 (2단계 중)

| # | 결정 | 선택지 | 추천 |
|---|---|---|---|
| M11 | 조회의 실제 반환 상한(토큰 잠금·예산 정리 포함) | 받아들임(오늘의 OPEN·단건 조회와 같은 성질, tasks 4.6 은 정상 경로 2s 만 단언) · 조회를 별도 goroutine 으로 빼 collect 가 2s 에 무조건 반환(늦은 응답은 버림) | **받아들임** — 같은 성질의 단건 조회가 이미 주문마다 있음. goroutine 은 새 동시성 표면 |
| M12 | 알림을 collect 임계 경로 밖으로 | 그대로(동기, 2s ctx) · 비동기 1칸 큐 | **그대로** — 같은 로거로 엔진 다른 루프도 동기 기록, 주문당 해결 전까지 1회 |
| M13 | 재조회 공정성 | 그대로(고정 집합에서 공정) · 대기시간 우선(lastLookup·firstSeen 중 오래된 것) | **대기시간 우선**으로 바꿔도 코드 1곳 — 결정 주면 반영. 지금은 Manager 지시(M8 ②) 그대로 구현 |

## 2단계 기록 (구현 팀메이트 Opus, 2026-11-16~17)

- RED: `analysis/red/red-1-core-behavioural.txt`(base 에서 행동 실패 3 — 스냅숏이 OPEN_UNFILLED 로 남음, 대사 로컬 열린 주문 남음),
  `red-2-new-api-compile.txt`(새 표면 컴파일 실패). 구현 뒤 작성한 시험 3개는 RED 관측이 없어 변이 m27~m29 로 대신 증명.
- GREEN: `analysis/green/`(filldetect 75 top-level PASS, 이웃 시험·`-race`·`make lint` rc 0). 기존 시험 46개 중 CLOSED 질의를 낸 시험 0
  (`existing-tests-closed-lookups.txt`, 임시 계측 뒤 cmp 로 원복 확인).
- 변이: `analysis/mutation/README.md` — m1~m32(m27~m29 포함) 중 m4 만 SURVIVED(동등 변이, 근거 README), 나머지 CAUGHT.
- FLM/BTM 번들 3(gate ⑤ 대상): `Detector.collect`, `JournalTracked.TrackedOrders`, `engineFillDetector`. 편집 전 AST `analysis/pre-edit-ast/`.
  새 함수 AST `analysis/new-code-ast/`. `check_analysis.py` 는 돌리지 않음(허용 목록 밖).
- retry matrix: a165·a166 선례(아카이브 뒤 같은 표를 고친 커밋 5eaa7f8d1·ffa2dadd4·b5bbae3b1)대로 아카이브 표에 GET 행·호출량 행, SLO 표에 행 추가.

### 코드 리뷰 1회차 (2026-11-17, Codex gpt-6-astra, read-only, 적대적 Eng)

원문 `analysis/review/code-review-1-codex.md`(대상 diff `code-review-1.diff`). 「변경 요청 — #1·#2」. 전부 내 코드 결함이라 고침:

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| 1 | MAJOR | settle 전에 실패한 주기의 후보가 쌓임 | `noteLeftOpen` 이 다른 주기 후보를 버림. 시험 `TestA191StaleCandidatesDoNotAccumulate`, 변이 m30 |
| 2 | MAJOR | 수용 즉시 간격 상태를 지워, 같은 주기 뒤 sweep 실패 시 다음 주기에 30s 간격 없이 재조회 | 수용 시 상태 유지(다음 주기 비후보면 정리). 시험 `TestA191AcceptedLookupKeepsTheIntervalWhenTheCycleFailsLater`, 변이 m31 |
| 3 | MINOR | 숫자 아닌 filledQuantity 가 row_unreadable 로 분류 | 파싱 실패 + 체결 수량 없음/불가독 → `closed_list_execution_missing`. 시험 사례 2 추가, 변이 m32 |
| 4 | MINOR | 알림 단언이 폴링 없이 시계만 진행해 공허함 | 시험 재작성 + `TestA191OrderThatStaysInTheOpenListIsNeverReported` |

### 코드 리뷰 2회차·최종 (2026-11-17, 같은 방식)

원문 `analysis/review/code-review-2-codex.md`. 「새 결함 없음, 1회차 4건 닫힘」. 정책 제안은 기존 M11~M13 뿐. 리뷰는 시험·변이를 다시 돌리지
않았다(정적). 회차 상한 2 도달.

## Manager 결정 (2026-11-17, 2단계 검증 뒤) · 잔여 위험

| # | 결정 | 잔여 위험(수용) |
|---|---|---|
| M11 | 받아들임 — goroutine 분리 없음 | 2s 는 취소 요청: 토큰 잠금(token.go:62~80)·a166 예산 정리(+2s)·예산 파일 Sync 가 ctx 를 안 따름. 기존 OPEN·단건 조회와 같은 성질 |
| M12 | 그대로(동기) | 경보의 동기 로그 쓰기는 시한으로 안 끊김(log.go:194~203) |
| M13 | 현행 유지 | 새 후보가 매 주기 들어오면 이미 물은 미해결 주문의 재조회가 밀림 |
| m4 | 동등 변이로 기록, 새 시험 없음, 방어 검사 유지 | `Derived` 생성자가 바뀌면 이 검사가 유일한 방어가 됨 |
| 검증 결함 | `check_analysis.py`(Manager 실행) rc 1 — collect BTM 에 B9~B12 누락 | **수리**: 시험별 커버리지 실측으로 행 채움, B12 거짓 갈래를 밟는 시험이 없어 `TestA191IncompleteOpenOrderNotTrackedIsStillApplied` 추가(변이 m33 CAUGHT) |

## Manager 검증 기록 (2026-11-17, 브랜치 커밋 9bf8d5796 뒤)

- **Q9 실측**(사용자 승인 2026-11-17, 콘솔 대안으로 수행 — 정확한 조회 형태는 미수행): 분쟁 주문 id 와 바이트 동일한 행 1건, 종결 그룹,
  CANCELED, 체결 수량 0(브로커가 보낸 값), 주문·체결 수량이 장부 단건 스냅숏과 같음 → Q3 가정 지지, M6 (가) 유지. 한계 ①~⑤ 는 design Q9 결과.
- **전체 시험**(사용자 요청 2026-11-17 선택지 1, Manager 실행, 요약만 기록 — 원문 로그는 저장소에 없음), worktree `9bf8d5796`(시작·끝 sha 동일,
  dirty 0, 캐시 적중 0): `make lint` rc 0(22s) · `make test` rc 0(911s, ok 122 패키지, FAIL 0) · `make test-seams` rc 0(975s, ok 123, FAIL 0) ·
  `make test-race` rc 0(248s, 7 패키지 + 엔진 이름 목록) · 추가 `go test -race -count=1 ./internal/filldetect` rc 0(67.9s, `make test-race` 목록 밖).
  미실행: `make gate`, `make sdd-check`, `make sdd-test`, `make validate`.

## 배포·복구 기록 (Manager 실측 2026-11-17 KST, 사용자 승인 — 계좌·종목·수량·가격·주문번호 미기재)

- main 반영 236a21d71(fast-forward). 이미지 `make image CHANGE=a191-…` rc 0 → `tossos:a191-fill-detection-resolves-orders-that-left-the-open-list-236a21d71`.
  바이너리 차등(`grep -a -c`): `fill.order_unresolved`·`closed_list_not_found` 새 이미지 각 1, 직전 a187 이미지 각 0. SchemaVersion 36 = 운영 저널 user_version 36.
  핀: release = a191, rollback = a187-7d1e3163a. 판정 파일 `~/.local/share/tossos-deploy/schema-compatibility-a191-236a21d71.txt`.
- 교체: httpapi 06:15:57~06:16:03 healthy, 엔진 06:16:03~06:16:10 healthy·engine run 1.
- **운영 동작**: 06:16:12 KST(2026-11-16T21:16:12Z) 첫 주기들 안에 분쟁 주문의 `scoped_fill_snapshots`·`fill_snapshots` 두 행이 terminal=1·fail_closed=0·
  `CANCELLED`. `fill.order_unresolved` 0건 → design Q9 한계 ① 운영 실행으로 지지. 그 시점 대사 `blocked` 2·미해제 `reconcile_states` 2(설계대로 저절로 안 풀림).
- 차단 해제: 엔진 정지 06:16:55 → 장부 백업 → `engine reconcile-resolve --confirm` rc 0(「quantity-mismatch reconcile blocks released after three stable official
  snapshots」) → 재시작 06:17:08 healthy·engine run 1. 엔진 부재 약 13초. 2026-11-16 같은 명령은 「1 order(s) the account does not show」로 거절됐었다.
- 복구 확인: 미해제 `reconcile_states` 0, `reconcile.clean` 두 주기 연속 `blocked` 0·`stable` true, 첫 주기 `adopted` 1, `exit.position_adopted` 1건
  (market kr, 공통 사다리 정책, synthetic stop 이 관측가 아래). 재시작 뒤 WARN/ERROR 0건. **관측 창은 재시작 뒤 약 70초뿐이다.**

## (기록) 1단계 시점 Manager 결정 요청 — 위 표로 닫힘



| # | 결정 | 선택지 | 추천 |
|---|---|---|---|
| M1 | 「지속되는 미해결」 경보 통로 | (A) `Detector.Alerts` ← `ectx.Notifier`, 새 이벤트 `fill.order_unresolved`(normal) · (B) 진입 게이트 심볼 래치 + 새 ReasonCode · (C) `Cycle` 만 | **A** — B 는 대사가 이미 같은 계좌를 막아 안전 효과 0, C 는 `Run` 이 결과를 버려 운영자에게 안 닿음 |
| M2 | 단건 파생이 UNKNOWN 인 목록 이탈 주문도 조회하나 | 포함 · 제외 | **포함** — US 정정 옛 id(`PENDING_REPLACE`+후속, M28)와 KR 정정 옛 id(`PENDING`+후속, M32)는 오늘 `open_with_successor` 로 영원히 UNKNOWN. 받아들이는 것은 CLOSED 의 깨끗한 종결 + 수량 일치뿐 |
| M3 | 간격·보고 시한 값 | 조회 간격 30s · 보고 시한 30s(SLO grace 와 같음) — 또는 다른 값 | 30s/30s |
| M4 | CLOSED 와 단건의 체결 수량이 다를 때 | (가) 오늘과 같음 + 경보(운영자) · (나) CLOSED 수량 채택 | **가** — 종결 목록 혼자서 체결을 만들거나 지우지 않는다(Q4). 비용은 KR 부분체결 후 취소에서 단건이 낡았을 때 운영자 1회 |
| M5 | Story 의 Feature | FEAT-TOS-002 · FEAT-TOS-003 | 002 |
| M6 | 수용 조건에 「행에 `execution`·`filledQuantity` 가 **있음**」을 더하나(리뷰 1 #1) | (가) 더함 — 체결 없이 취소된 정상 행이 execution null 이면 그것도 거부 [미측정] · (나) 안 더함 — 오늘 단건과 같은 성질 | **Q9 조회 결과를 보고 정함**. 조회 전이면 (가) — 거부의 비용은 「오늘과 같음 + 경보」라 작다 |
| M7 | `HasNext=true`+빈 커서를 거절하나(리뷰 1 #2) | 새 CLOSED 조회에만 별도 검사 · `ScanOrders` 자체(OPEN 걷기·IN_DOUBT·대사·flatten 에도 영향) · 안 함 | **새 조회에만**(새 파일 안, 기존 함수 무편집). `ScanOrders` 수정은 별도 change |
| M8 | 조회가 같은 주기의 다른 주문 반영과 공유 wire 를 늦추는 문제(리뷰 1 #5·#7, 리뷰 2 #3) — **2단계 착수 전 필수, 허용 지연·총량 수치를 정해야 함** | (가) **조회 전체**(페이지 합계)에 짧은 시한(예 2s) + 주기당 조회 1건 상한 + 같은 (종목·창) 질의 재사용 · (나) 조회를 `Apply` 뒤로 미룸(`pollLocked` 편집 — a100 인용 번들도 흔듦). **(나) 만으로는 진행 중 CLOSED 요청이 a166 공유 wire 를 응답 끝까지 점유하는 것을 못 막는다**(Critical 은 다음 소유자 선택에서만 우선 — budget_linux.go:549~584, transport_linux.go:177~188, 리뷰 2 인용) · (다) 둘 다 | **(가)** — wire 점유 시간 자체를 줄이고 `collect` 2줄 편집 범위를 유지. 상한을 넘은 주문은 `closed_lookup_deferred` |
| M9 | CLOSED 조회 실패·미해결이 있어도 OPEN freshness 를 도장하나(리뷰 1 #6) | 도장(오늘 동작 유지 — **잔여 위험 수용**) · 보류 | **도장, 단 안전장치라고 주장하지 않음**(리뷰 2 #4 로 근거 정정): 대사는 별도 루프라 첫 관측 전 공백이 있고 missing-order 의 첫 블록은 **그 심볼만** 막는다(mismatch.go:1418~1427). 이 공백은 오늘도 같다. 보류하면 CLOSED 429 하나가 계좌 전체 진입을 staleness 로 막는다 |
| M10 | 수용 뒤 안전망 없는 잘못된 종결(리뷰 1 #8: 브로커가 살아 있는 주문을 CLOSED 에 실음 등) | 잔여 위험으로 둠 · 종결 뒤 OPEN 재등장 감시(재등장 시 경보) · 운영자 확인 전용 | **잔여 위험, 재등장 경보는 후속 change** — 같은 위험이 오늘 단건 `CANCELED` 에도 있다 |

## 사용자 승인 필요

- (선택) design Q9 의 읽기 전용 조회 1회 — 정확한 형태는 새 일회용 도구가 필요하고, 새 코드 없는 대안은 콘솔 주문 화면 1회.
- 배포·운영 장부 복구(엔진 정지 창, `reconcile-resolve` 실행) — 2단계 완료 뒤.
