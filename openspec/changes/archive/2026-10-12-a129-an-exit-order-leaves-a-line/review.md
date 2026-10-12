# a129 review

## Pre-Edit Gate (2026-10-12, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a129-an-exit-order-leaves-a-line / 1.2
- 대상 심볼(패키지.함수): engine.ExitObserver.submit · engine.ExitObserver.ObserveOnce · main.engineRuntime
  (+ 타입 engine.ExitObserverOptions 필드 하나, 새 파일 internal/app/engine/exit_order_log.go — 새 leaf)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/)
  — submit 호출자 생산 1(record) · ObserveOnce 생산 2(Run · Tracer.Run) · engineRuntime 생산 1(engineRuntimeFactory)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
  (R2: CodeGraph 의 ExitObserver.log 호출자 해석 오류 — 계좌 판정은 HEAD grep 아홉 자리로)
- 기존 동작 파악 근거: base 381e73da HEAD 직접 읽기 · AST 번들 3 · 엔진 패키지 전체 커버리지(set, 228s, rc 0) ·
  기존 시험 TestABaselineBreachProposesTheWholePosition / TestAnInDoubtSubmissionKeepsTheProposalArmed /
  TestARefusedProposalReleasesTheLevelAndAlerts / TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable /
  TestA094AnUnrecordedOutcomeKeepsTheProposalArmed / TestA090EngineRuntimeWiresOnlyTheDedicatedUnobservedLogger
- Function Logic Map / Branch Test Map: analysis/function-logic/{internal-app-engine--exitobserver.submit,
  internal-app-engine--exitobserver.observeonce,cmd-tossctl--engineruntime}/ (편집 전, 분기 13 / 8 / 6)
- upstream 상속 테스트 영향: no — 엔진 패키지는 TossOS 고유. 두 패키지 전체 go test -race 로 회귀 확인
- 실패 테스트 선행 작성: yes — analysis/red/red-1.txt (엔진 시험 컴파일 실패 · cmd 구조 핀 FAIL, production 무편집 HEAD 381e73da)
- 설정·DB·journal 변경과 rollback: 없음 — 옵션 필드 하나(선택), 원장 읽기(LookupFill)만. 롤백 = 이미지 핀 되돌림
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문·해제·판정 무변경(줄은 Place 뒤 · 해제 뒤), 로거 패닉 recover, nil = 종전,
  계좌 필드 화이트리스트 밖, 브로커 호출 0 추가, 토글 flip 없음
```

## Proposal-freeze review (2026-10-12)

- 보이스: 적대적 Eng 1(분리된 리뷰 컨텍스트, code-reviewer, 읽기 전용 — 저장소 무편집 확인). 대상: proposal · design 1판 · spec 1판 · tasks · FLM/BTM 3 · code-context.
- 판정: **APPROVE-WITH-FIXES** (P0 0 · P1 2 · P2 6 · P3 10). 「`Place` 앞·해제 앞에 줄이 서지 않는다」는 코드 대조로 성립, 손절 즉시성·주문 동작 무접촉 확인.
- 처리는 GREEN 전에 design 2판 · spec 2판 · 시험에 반영했다.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| R-1 | P1 | `LookupFill` 은 재사용된 브로커 주문 번호에서 다른 날 주문의 종결 기록을 이 주문의 체결로 읽고 따라가기를 끝냄(거짓 「손절 체결」 줄) · 두 행이면 매 주기 ambiguous | **수용** — `LookupIntent` 로 정규 신원 → `LookupFillScoped`(D5). 시험 H(다른 날 같은 번호) · 변이 M10 |
| R-2 | P1 | engine.log 는 이미 계좌를 평문으로 실음(`reconcileloop.go:538/543/565`, `runtime.go:523`) — D1 의 전제가 거짓 | **수용** — D1 정정(정당화 = 「a129 는 계좌를 더하지 않는다」), spec 시나리오를 a129 줄로 한정, 기존 불변식 8 잔여는 **범위 밖 · 별도 change 필요**로 보고 |
| R-3 | P2 | `position_id` 는 소금 없는 해시라 짧은 계좌 참조는 대입으로 복원 가능 | **잔여로 명명**(D1) — Manager 화이트리스트가 요구하고 a090 이 이미 실음. R-2 잔여를 고칠 때 함께 |
| R-4 | P2 | 넓은 recover · `defer` 줄은 `Place`/`release` 패닉을 삼켜 nil 반환이 될 수 있음 | **수용** — 갈래별 명시 호출(defer 금지, spec 명시), recover 는 쓰기 하나만 감쌈. 시험 K(`Place` 패닉 전파) |
| R-5 | P2 | 해제된 거절은 5초마다 재발의 → 무한 반복 줄(engine.log 회전 없음) | **수용** — 포지션별 거절 서명 래치(D4-4). 시험 J · 변이 M6 |
| R-6 | P2 | 체결 0 으로 종결된 손절이 줄을 안 남김 | **수용** — `exit.order_ended`. 시험 I · 변이 M11 |
| R-7 | P2 | `fill_lookup_failed` 를 `fill.observed` 로 쓰면 세는 규약이 깨짐 | **수용** — `exit.fill_follow_failed` 신설(엔진 패키지 상수, `obs` 편집 범위 밖 — Manager 판단 사항) |
| R-8 | P2 | spec 「늘 때마다」 과장 · 재시작 잔여가 design 에만 | **수용** — 「주기마다, 마지막으로 쓴 값보다 크면」 + 잔여 문장을 spec 에 |
| R-9 | P3 | submit 안 동기 쓰기는 같은 주기 뒤 포지션의 손절 판정 앞에 섬 | **수용** — 대기열 + 주기 끝 쓰기(D4-3) |
| R-10 | P3 | 같은 로거의 다른 쓰기엔 recover 없음 — 생산 보호는 국소적 | 명명(D6) |
| R-11 | P3 | B1 은 하한 읽기 실패가 아니라 산술 실패 | **수용** — `floor_arithmetic_failed` |
| R-12 | P3 | `UNRESOLVED_IN_DOUBT` 는 `order.unresolved_in_doubt` | **수용** — 시험 D 「B8 unresolved」 |
| R-13 | P3 | 빈 사유가 `unrecognised` 로 바뀜 | 해당 없음 — 구현은 빈 값을 빈 값으로 둠(`exitOrderToken`) |
| R-14 | P3 | `Log` · `ExitOrderLog` 둘 다면 확정 두 줄 | 명명(D2) — 생산 불가능 구성 |
| R-15 | P3 | 자리 수 9 → 10 · 수량 비교는 `CompareDecimal` · ctx 끝나면 읽지 않기 | **수용** |
| R-16 | P3 | 변이 목록에 defer → 직접 호출 · 번호만 조회 추가, 줄 쓰기 nil 가드 삭제는 동등 변이 | **수용** — M8 · M10, 동등 변이는 D7 에 명시 |
| R-17 | P3 | engine.log `severity` 를 읽는 코드 없음 | 확인 — 조치 없음 |
| R-18 | P3 | 키 충돌 없음 · defer 가 반환값 못 바꿈 · 단일 goroutine · a090 핀 호환 · ObserveOnce 가 filldetect 장식보다 나음 · slog 잠금 defer 해제 | 확인 — 조치 없음 |

## Implementation review (2026-10-12)

- 보이스: 적대적 Eng 1(분리된 리뷰 컨텍스트, code-reviewer, 읽기 전용 — 저장소 무편집·`git status` 전후 동일 확인). 대상: working tree vs HEAD `381e73da`.
  리뷰어 실측: `-race -run TestA129` 13/13 PASS · 두 패키지 전체 `go test`(무 race) rc 0 · `go vet` 두 태그 · 실제 `gofmt -l` 깨끗 ·
  Go 1.26.5 에서 「defer 된 end() 안의 중첩 recover 는 바깥 패닉을 삼키지 않음」 재현.
- 판정: **APPROVE-WITH-FIXES** (P1 1 · P2 2 · P3 7). 제출 경로 반환값·해제·알림·지연 시계는 모든 갈래에서 base 와 같음(err 가림 없음) 확인.
- spec 요구 수정(I-1 의 「연속한 주기」 · I-2 의 수명)은 이 리뷰 보이스의 제안을 그대로 반영한 것이라 이 리뷰를 requirement 변경 리뷰로 기록한다.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| I-1 | P1 | 거절 래치가 에피소드를 넘어 남음 — 며칠 뒤 같은 거절이 0 줄 | **수용** — 주기 끝에 이 주기 거절이 없던 포지션의 래치를 풂(연속한 주기만 묶음). 구현 중 「래치로 거른 거절도 이 주기 표시를 먼저」 결함을 시험 J 가 잡아 고침. 시험 M · 변이 M13 |
| I-2 | P2 | 따라가기 항목 무기한 — 주기당 원장 읽기(연결 1개 풀) | **수용** — 수명 24h 뒤 `follow_expired` 한 줄 · 버림. 시험 N · 변이 M14 |
| I-3 | P2 | `return o.release` 를 푼 다섯 자리의 해제 실패 전파 시험 0 | **수용** — 원장을 닫아 해제 실패: B2 · B3 · B10 · B13 → 주기 오류. 시험 O · 변이 M15~M18. B5 는 도달 불가(count 0)라 변이 없음 |
| I-4 | P3 | 패닉 중 defer 의 원장 읽기가 막힌 연결에 영원히 설 수 있음 | **수용(경량)** — 주기 끝 읽기 전체에 1초 시한 |
| I-5 | P3 | 거절 래치 맵이 닫힌 포지션마다 남음 | I-1 수리로 해소(주기 끝 정리) |
| I-6 | P3 | 따라가기 실패 래치가 회복 뒤에도 안 풀림 | **수용** — 읽기 성공 시 래치 풂 |
| I-7 | P3 | 날을 넘어 재사용된 같은 주문 번호의 두 주문이 한 항목으로 겹침 | **수용** — 키를 intent id 로 |
| I-8 | P3 | 줄의 지정가가 `sellIntent` 규칙의 손 사본 | **부분 수용** — `sellIntent`(High-risk 기존 함수)는 편집하지 않고, 시험 A 가 줄 지정가 = 실제 제출 지정가를 고정 |
| I-9 | P3 | 시험 D 는 submit 안 해제·대기열 순서를 못 잼 · M1 은 info 길만 | **수용** — 시험 D 주석을 「주기 끝 쓰기를 고정」으로 고침, warn 길 변이 M1b 추가 |
| I-10 | P3 | `position_id` 잔여 · id 필드는 모양 거름 없음 | 명명(D1) — 원장·브로커 키, 계좌 없음 확인 |
