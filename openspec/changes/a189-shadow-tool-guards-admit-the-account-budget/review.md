# Review

## 1단계 기록 (구현 팀메이트 Opus, 2026-11-15)

### 기억 회고

- ai-memory(MCP `memory_query`, 질의 `static guard allowlist a128 a140 accountbudget`, 현재 프로젝트): 세션 발췌 8건뿐임(a184 리뷰 세션, a128 예약 작업, a140 기본 모드 문구 등). 이 가드들의 허용 목록·a166 배선 누락을 다룬 결정·함정 페이지는 없음.
- Hindsight(bank `tossos-dev-episodes`, 질의 `static guard allowlist a128 a140 A112 accountbudget shadow probe import allowlist`): 게이트 종결 교훈만 나옴(a178 a112 e2e 시한, a165 stray `.git`·ext4 홈, a172 「시험 함수 본문 편집도 gate ⑤ 에 걸림」, a176 번들 복사). 이 가드 실패를 다룬 기록은 없음. 그중 「시험 함수 본문을 고치면 gate ⑤ 가 번들을 요구」는 이 change 에 해당하므로 D6 에 반영함.
- 파일 기억 `tossos-fresh-worktree-gate-pitfalls.md`·`tossos-logic-map-scope-creep.md`·`flm-exemption-only-when-no-function-modified.md` 를 읽고 현재 HEAD 로 확인함. worktree 는 저장소 밖(`~/tossos-work/`)에 둠. 시험 파일의 기존 함수 5개를 고치므로 번들 5개가 필요함(면제 불가).

### 번호 배정

- 확인 범위: `git branch -a`(a189·a190 0건), `git worktree list` 58개(a188 = `~/tossos-work/a188-us-gainers` `feat/a188-us-gainers-ranking-exploration`, a189 0건), `~/tossos-work`(a189 0건), `openspec/changes`·`archive`(a18x 활성은 a180·a187), `docs/pm/portfolio/`(Story a180·a187 까지), `git log --all --oneline -3000`.
- 제목에 `a189` 가 나온 커밋은 3건이고 전부 커밋 해시 접두사 `a189e74f7`(a094·a126 기록)였음. change 번호가 아님.
- a189 를 씀.

### Story·Feature

- `STORY-TOS-a189` → `FEAT-TOS-002`(Broker execution verification and measurement).
- 근거: 원인과 결과가 모두 a166(`STORY-TOS-a166`, FEAT-TOS-002)의 배선 로트(`78d777cfe`·`4d5884cbc`)에 묶임. 스펙 델타도 측정 도구의 공유 예산 참여를 적음.
- 대안이었던 FEAT-TOS-001(SDD toolchain)은 게이트를 푸는 효과만 보면 맞지만, 바뀌는 것은 SDD 도구가 아니라 측정 도구의 안전 가드라 고르지 않음.

### 증거

- 열거: `analysis/enumeration/enumerate_surface.go.txt`(하네스).
  - `.go.txt` 로 커밋한 이유: 저장소 전수 census(A112 `WalkDir` 등)·`go test ./...` 에 안 잡히게 하려는 것임.
  - 원문: `analysis/enumeration/selectors-and-imports.txt`(rc 0), `analysis/enumeration/official-option-impact.txt`.
- 편집 전 AST: `analysis/pre-edit-ast/`.
  - 선택지 A 의 5개: `official--a112mbusreferenceallowed`, `a128--testnomutatingoraccountsurfaceisnamed`, `a128--testimportsstayinsidetheallowlist`, `a140--testnomutatingaccountorwritesurfaceisnamed`, `a140--testimportsstayinsidetheallowlist`.
  - 선택지 B 비교용 1개: `optionB-accountbudgetruntime--testa166runtimev2a112clonerecordspending`.
  - 명령: `GOFLAGS=-trimpath go run ./tools/logic-map --file … --func …`.
  - design Context 의 분기 번호·행은 이 AST 에서 옮김.
- CodeGraph: `analysis/code-context/`(1.6.0, 원문 포함).
- 확인한 원인: a166 diff(`git show 78d777cfe -- tools/*/main.go`)가 두 도구에 같은 배선을 넣음. 넣은 것은 import 2개, `runCommandWithOptions(…, opts ...official.Option)`, `accountbudgetruntime.ParseLaunchConfig`/`Open`, `factory.NewClient`, `accountbudget.WithPriority(ctx, accountbudget.Auxiliary)` 이고, 시험 파일은 `a166_budget_test.go` 만 더함.

### Pre-Edit 선언

```text
Pre-Edit Gate:
- change id / task id: a189-shadow-tool-guards-admit-the-account-budget / 2.2
- 대상 심볼(패키지.함수):
  - official.a112MBUSReferenceAllowed (선택지 A; B 면 accountbudgetruntime.TestA166RuntimeV2A112CloneRecordsPending 삭제 + 새 시험 파일)
  - main(tools/a128-jev-shadow-probe).TestNoMutatingOrAccountSurfaceIsNamed
  - main(tools/a128-jev-shadow-probe).TestImportsStayInsideTheAllowlist
  - main(tools/a140-breakout-signal-shadow).TestNoMutatingAccountOrWriteSurfaceIsNamed
  - main(tools/a140-breakout-signal-shadow).TestImportsStayInsideTheAllowlist
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md — 운영 호출자 0, a112MBUSReferenceAllowed ← a112MBUSFindForbiddenRefs ← (RepositoryForbiddenRefs, 픽스처 시험 3)
- CodeGraphContext 후보와 evidence reconciliation: CGC 안 봤음(사유 codegraphcontext-context.md), reconciliation 불일치 0
- 기존 동작 파악 근거: HEAD 7d1e3163a 소스·편집 전 AST 5개, base 실패 원문(a187 analysis/verify/base-failures.txt), AST 열거 원문
- Function Logic Map / Branch Test Map: 편집 전 AST analysis/pre-edit-ast/ (번들은 2.3 에서 편집 뒤 생성)
- upstream 상속 테스트 영향: no — 세 파일 모두 TossOS 고유(a112·a128·a140) 시험
- 실패 테스트 선행 작성: no 새 시험 — RED 는 이미 실패하는 기존 다섯 시험(2.1 에서 이 worktree 원문 확보); 양성 대조는 3.1 사본 변이
- 설정·DB·journal 변경과 rollback: 없음 — 시험 파일만, rollback 은 revert
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문·손절·인증 경로 무편집, 금지 이름 목록 무편집, 운영 토글·live 호출 없음
```

### not-applicable 사유

- 다중 적대 리뷰: not-applicable. 비례 원칙에 따라 High-risk 경로나 게이트 판정을 직접 바꾸는 변경이 아님(시험 전용 허용 목록). gstack `review` 1회로 함(tasks 3.3).
- 변이 원장: 비례 원칙상 의무는 아님. 다만 안전 가드의 판정을 넓히므로 사본 변이(양성 대조군)를 3.1 에서 돌림(Manager D7).
- 스펙 델타의 a128 프로브 요구사항: 대응 요구사항이 정본에 없음. `semantic-judgment-shadow` 를 `import|정적|static|allowlist|허용|이름|계좌|account` 로 찾으면 `account` 3건뿐이고 셋 다 평가 해석 문장(「independent account trades」·「account return」 등)이라 정적 검사 문장이 아님, `openspec/specs` 전체에서 `jev-shadow-probe`·`a128-jev` 0건. 그래서 a128 델타는 쓰지 않음.
- a166 의 기존 델타(`openspec/changes/a166-account-shared-rate-limit/specs/order-execution/spec.md`)는 admission·Retry Matrix 만 다루고 측정 도구의 정적 검사 문장이 없음(`tool|도구|그림자|shadow` 검색 0건). 겹침 없음.
- 같은 capability 를 바꾸는 활성 change: a169(ADDED 새 요구사항 하나)뿐. 이 change 는 다른 요구사항을 MODIFIED 하므로 충돌 없음.

### 남은 위험·확인 못 한 것

- design Risks R1~R3.
- `tools/a112-mb-us-source` 시험이 가드 파일 sha 변경에 반응하는지는 아직 안 돌려 봄(3.2).
- `.gz` 영수증 안의 sha 고정 여부는 안 봤음.

## Manager 결정 (2026-11-15)

- D4 = A: `a112MBUSReferenceAllowed` 의 마지막 정확 경로 비교에 `internal/accountbudgetruntime/v2_token_test.go` 하나만 더함. 접두사·접미사 조건 무편집. A-wide 기각.
- 스펙 = A-2: M-B0 요구사항의 한 문장만 개정. 정본 블록과의 차이가 그 문장 하나뿐임은 `analysis/spec-a2-mb0-diff.txt`(글자 단위 차이 2건이 같은 문장 안, 옛 문장 제거·새 문장 제거 뒤 블록 동일 True).
- R1(별칭 import 우회) = 범위 밖. 후속 항목은 아래 「후속」.

## 2단계 기록 (구현 팀메이트 Opus, 2026-11-15)

- 2.1 RED(`analysis/tdd/red-2.1.txt`, HEAD cade0c638): 다섯 시험 FAIL 5·PASS 0, 문구는 base 실패 원문과 같음.
- 2.2 GREEN(acc14449d, `analysis/tdd/green-2.2.txt`): PASS 5·FAIL 0. 편집은 proposal 「What Changes」의 이름뿐. a128 import 실패 문구의 괄호 목록에 두 패키지를 더함(proposal 에 적은 대로 — 문구가 허용 목록과 어긋나지 않게).
- 3.1 변이 중 **시험 1개 추가**(6e6b29a5e): `internal/official/a112_mbus_a189_static_test.go` `TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile`. 1차 변이에서 정확 경로를 접두사(g2)·접미사(g3)로 넓힌 변이가 **SURVIVED** 했기 때문임(`analysis/mutation/mutation-before-fixture.txt`). 그 수리 없이 Manager 지시 g 항목(「접두사로 넓힌 변이 → 그 픽스처에 잡혀야 함」)을 만족할 커밋된 시험이 없었음. 새 파일·새 함수라 gate ⑤ required 는 5 그대로.
- 2.3 FLM/BTM 번들 5개(d929433ba), `check_analysis.py` 원문 `analysis/tdd/check-analysis-2.3.txt`: `evidence complete or diff-proven exempt`, rc 0.
- 3.1 변이: 1차 대조군은 사본에 `docs/migration` 이 없어 a140 `TestA169SecurityTypeOpenAPIEnum` 이 FAIL → 판정 중단(`history-1-control-missing-docs-migration.txt`). 사본 목록을 고쳐 다시 돎. 픽스처 전 18종 중 16 CAUGHT·2 SURVIVED(g2·g3), 픽스처 뒤 18/18 CAUGHT(`mutation-after-fixture.txt`, 기준 6e6b29a5e). gstack 리뷰 반영 뒤 하네스(기대 시험 이름 단언·HOME/XDG 격리)로 3차(`mutation-after-review.txt`).

## gstack 리뷰 처리 (tasks 3.3, 2026-11-15)

- 원문: `analysis/gstack-review.md`. 디스패치: specialists 5(testing·maintainability·security·performance·simplification), Red Team 1, Claude 적대 패스 1, Codex(gpt-6-astra) 적대 패스 1(OUTSIDE_STATUS completed, `VERDICT: FINDINGS`, P3 1건). maintainability·security·performance·simplification 은 NO FINDINGS.
- 다중 적대 리뷰는 비례 원칙상 not-applicable 이지만, gstack `review` 스킬이 DIFF_LINES(5370, 대부분 증거 파일)로 Red Team·적대 패스를 자동으로 켰음. 그 결과를 그대로 기록함.
- 결론: 오늘 시험을 깨거나 운영을 깨는 발견은 0. Medium 2건(발견 1·2)과 스펙 문구 2건(발견 9·10)은 Manager 결정으로 올림 — 「기록만」으로 닫지 않음.

| # | 출처 | 등급 | 내용 | 처리 | 근거 위치 |
|---|---|---|---|---|---|
| 1 | Claude 적대 1 + Red Team 2 | Medium, Blocks no | 예산 이름 고정이 패키지 한정 선택자만 봄. `accountbudget.WithPriority(ctx, 1)`(상수 1 = Critical)나 `factory.Budget()` → `Acquire`·`EnginePollKey`·`PollRemaining`·`ObservationSnapshot`·`NewObservationFence` 가 가드를 통과함. 스펙 델타의 SHALL(「예산 시작과 보조 우선순위 표시 … 로 한정」)과 시나리오(「목록 밖 이름 … 여전히 실패」)가 가드보다 넓게 말함 | **Manager 결정 대기**. 안 (a) 두 이름 시험에 `WithPriority` 둘째 인자가 정확히 `accountbudget.Auxiliary` 선택자인지 AST 로 고정하고, `Budget`·`Acquire` 등 예산 메서드 이름을 두 `forbidden` 에 더함(두 도구 비시험 소스에 0건 — 오탐 없음), 변이 2종 추가. 기존 두 함수 본문 안이라 번들 수 불변. 안 (b) 스펙 문장을 가드가 실제로 재는 범위로 좁힘 | 아래 「Manager 에게 올림」 |
| 2 | testing + Claude 적대 2 | Medium, Blocks no | 별칭 import 우회가 새 두 패키지까지 넓어짐(편집 전에는 import 단계에서 막혔음). 「본 것 없음」 검사는 전부 별칭일 때만 잡음 | design R1 문장을 사실대로 정정(9dbe116c7). 수리는 Manager 결정(2026-11-15)으로 범위 밖 → 후속 | design.md Risks R1 |
| 3 | Claude 적대 3 | Low | M-B0 허용이 파일 전체 면제(오늘 참조 1개, v2_token_test.go:463) | 기록만 — design R2 에 이미 적힌 한계, Manager D4 결정 범위 | design.md R2 |
| 4 | Claude 적대 4 | Low | 하네스가 패키지 안 아무 시험이나 실패하면 CAUGHT 로 셈 | 수정: 변이마다 기대 시험 이름, `--- FAIL: <이름> (` 가 없으면 WRONG-CATCH | mutation.py, `mutation-after-review.txt` |
| 5 | Claude 적대 5 | Low | 하네스가 HOME/XDG 격리 없이 돎(기존 a166 도구 시험이 기본 토큰 캐시 경로를 읽음) | 수정: 하네스의 go 실행을 HOME·XDG_* 임시 디렉터리로 격리(모듈 캐시·go env 파일 경로는 원래 값 명시). 기존 a166 시험 자체의 경로 의존은 a166 파일이라 후속 | mutation.py `isolated_env` |
| 6 | Claude 적대 6 | Low | 하네스 무결성 검사가 `assert`(python -O 에서 사라짐) | 수정: `Fail` 예외 + `check()` | mutation.py |
| 7 | Codex | P3/Low | 알 수 없는 변이 ID 가 아무것도 안 돌고 성공 | 수정: 사본 만들기 전에 거절, rc 2(실측 `알 수 없는 변이 ID: ['nope-id']`, rc=2) | mutation.py |
| 8 | Red Team 1 | Low | 커밋된 `mutation-after-fixture.txt` 가 빈 파일(진행 중에 커밋됨) | 수정: 완성본 커밋(a50e7bec9, 기준 6e6b29a5e 명시는 파일 첫 줄) | analysis/mutation/ |
| 9 | Red Team 3 | Low(c6) | 스펙 델타 「예산 참여는 읽기 요청의 전송 허가와 시점만 정한다」가 사실보다 좁음 — ModeRequired 에서 클라이언트 전송이 요청마다 계좌 공유 관측 상태(engine 이 읽는 event.state)에 Begin/Finish 를 씀. 같은 요구사항의 「evidence 쓰기는 자기 저장소에만」과 대조가 없음 | **Manager 결정 대기**(요구사항 문장 변경 → 리뷰 재실행 대상). 제안 문장은 아래 | 아래 |
| 10 | Red Team 4 | Low(c5) | 예산 환경 변수가 없으면 ModeOff = 예산 없이 돎(a166 기존 동작). 스펙 델타가 그것을 말하지 않음 | **Manager 결정 대기**(스펙에 한 문장 또는 운영 문서) | 아래 |
| G1~G3 | gate scan | — | 가드 허용 목록을 넓힌 hunk 3개(`80d9fc903a8e` a112 정확 경로, `cfe5b07efdd8` a128 import, `976f27364b5d` a140 이름) | **open**(서브에이전트는 사람에게 묻지 못함). 추천: B) Keep — 정당화됨(검증됨: AST 열거 `analysis/enumeration/selectors-and-imports.txt` 와 변이 18/18) | gate scan 원문 |

### Manager 에게 올림

1. 발견 1: 안 (a) 를 추천함 — 안전 불변식 4(손절 즉시성)와 직결되는 우선순위 상승을 정적으로 막고, 편집은 이미 고친 두 함수 안이라 번들 수가 늘지 않음. 대략 함수당 10줄 안팎 + 변이 2종. 안 (b) 는 스펙이 약속을 덜 하게 될 뿐 구멍은 남음.
2. 발견 9 제안 문장: 「예산 참여는 읽기 요청의 전송 허가와 시점을 정하고, 그 요청의 시작·끝을 계좌 공유 관측 상태에 기록한다. 이 기록은 측정 도구가 자기 evidence 저장소 밖에 쓰는 유일한 상태이며, 주문 · 계좌 · 보유 · 매수 가능 표면을 열어서는 안 된다 (MUST NOT).」
3. 발견 10 제안 문장: 「예산 환경 변수가 없으면 측정 도구는 예산 없이 돈다(OFF — upstream 동작).」 또는 운영 문서에만 적음.

### 후속 (범위 밖, 기록)

- R1 별칭 import 우회 — 두 도구 가드를 import 경로 기준으로 바꾸기(별도 change).
- 기존 a166 도구 시험(`tools/a128-jev-shadow-probe/a166_budget_test.go`·`tools/a140-breakout-signal-shadow/a166_budget_test.go`)이 `--config-dir` 없이 기본 토큰 캐시 경로를 읽음 — a166 쪽 시험 격리.

### 리뷰 재실행

- 리뷰 뒤 편집은 하네스(Python)·문서뿐이고 Go 변경 없음. 수정분은 변이 3차로 다시 잼(`mutation-after-review.txt`). gstack 의 전체 재리뷰 패스(Steps 3~5 재실행)는 돌리지 않음 — converged 아님으로 기록.

## 3.2 영향 패키지 시험 (2026-11-15, HEAD 844611265)

- 명령·원문: `analysis/verify/go-test-5pkg.txt`(`rtk proxy go test -count=1 -v` 5개 패키지), `go-vet-5pkg.txt`(rc 0), `gofmt.txt`(0줄).
- 결과: `internal/official` ok, `tools/a128-jev-shadow-probe` ok, `tools/a140-breakout-signal-shadow` ok, `tools/a112-mb-us-source` ok(163.1s, SKIP 2 — 기존 skip), `internal/accountbudgetruntime` **FAIL**.
- 수치: 최상위 PASS 551, 서브 포함 PASS 961, FAIL 3, SKIP 2.
- FAIL 3 은 모두 a166 배선 census:
  - `TestA166WiringCensusBaseline` — `a166_wiring_census_test.go:303: reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions`
  - `TestA166WiringCensusRejectsBypasses` 와 하위 `new-unclassified-http` — `mutant new-unclassified-http not caught by "callsite inventory count changed": reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions`
- base `7d1e3163a` 사본(git archive)에서 같은 두 시험이 같은 문구로 실패함(`analysis/verify/base-a166-census.txt`). 이 change 는 `main.go` 를 건드리지 않음. `4d5884cbc..7d1e3163a` 사이에 그 파일을 바꾼 커밋은 `9abcb631a`(a179 입력 프로필 비교)와 병합 `d7c078ffc` 둘 — census 의 검토된 함수 모양을 a179 가 바꾼 것으로 보임(원인 확정은 안 함).
- 의미: Manager 배경의 「정적 가드 5개」 목록에 없는 실패가 하나 더 있음. a187 의 3.2 범위(6개 패키지)에 `internal/accountbudgetruntime` 가 없어서 안 보였던 것으로 보임. 이 change 를 main 에 넣어도 `go test ./...` 는 이 census 로 계속 FAIL 함 → `make test` 7단계는 아직 못 통과함.

## 보강 (Manager 결정 2026-11-15 — gstack 발견 1·9·10, gate hunk)

- Manager 확인: Go diff 4파일은 D1~D3 범위 안. 새 시험 파일 `a112_mbus_a189_static_test.go` 추가와 a128 실패 문구의 괄호 목록 갱신을 승인함.
- 발견 1 = (a) 적용(커밋 ab7023901): 두 도구 이름 시험이 `accountbudget.WithPriority` 의 둘째 인자를 선택자 `accountbudget.Auxiliary` 로 고정함. 직접 호출만 받음(함수 값 참조 수 ≠ 직접 호출 수면 실패), 호출이 0 이면 실패. `forbidden` 에 예산 핸들 메서드 7개(`Budget`·`Acquire`·`EnginePollKey`·`NewObservationFence`·`ObservationSnapshot`·`PollRemaining`·`VerifyIdentity`)를 더함. 열거·0건 근거·앞으로 거절할 정상 사용은 design D9.
- 발견 9 = 스펙 문장 정정(같은 커밋). 전·후 원문 `analysis/spec-shadow-requirement-revision.txt`.
  - 전: 「예산 참여는 읽기 요청의 전송 허가와 시점만 정하며, 주문 · 계좌 · 보유 · 매수 가능 표면을 열어서는 안 된다 (MUST NOT).」
  - 후: 「예산 참여는 읽기 요청의 전송 허가와 시점을 정하고, 요청마다 계좌 공유 예산의 관측 상태(요청의 시작 · 끝)를 기록한다. 이 관측 상태는 예산이 소유하는 공유 상태이며 측정 evidence 가 아니다 — 위 「evidence 쓰기는 자기 전용 evidence 저장소에만」 은 측정 evidence 에 대한 것이고 이 기록에는 적용되지 않는다. 측정 도구는 예산 핸들을 직접 잡거나 관측 상태를 읽어서는 안 되고 (MUST NOT), 예산 참여로 주문 · 계좌 · 보유 · 매수 가능 표면을 열어서는 안 된다 (MUST NOT).」
  - 「보조 등급」 SHALL 은 유지하고 「우선순위는 보조 등급만 써야 한다 (SHALL)」·「우선순위 인자가 보조 등급임을 확인해야 한다 (SHALL)」 를 더함(이제 가드가 잼). 시나리오 THEN 에 「우선순위 인자가 보조 등급이 아니거나 우선순위 표시 호출이 없거나 예산 핸들 메서드 이름이 나오면 실패」 를 더함.
- 발견 10 = 스펙에 넣지 않음. 델타 문장은 「보조 등급 참여자로 연결될 수 있다」(허용)이고 「참여한다」(보장)가 아님을 확인함. 후속: ModeOff(예산 환경 변수 없음 = 예산 없이 돎)는 a166 launch 계약 — a166 에 인계.
- gate hunk 3건(`80d9fc903a8e`·`cfe5b07efdd8`·`976f27364b5d`) = Keep(B). 근거: AST 열거와 변이 결과.
- 변이 4차(`analysis/mutation/mutation-after-priority-pin.txt`, 기준 ab7023901): **28/28 CAUGHT**(기존 18 + 새 10 = 두 도구 × h·i·j·k·l). 새 변이는 모두 기대 시험(이름 시험)의 새 문구로 잡힘 — i 「source calls accountbudget.WithPriority(ctx, 1); … only accountbudget.Auxiliary …」, l 「direct calls 0 / references 1」, k 「direct calls 0 / references 0」, j 「source names Budget」·「source names Acquire」. h 는 인자 고정과 이름 허용 목록 둘 다에 걸림. l(함수 값)은 Manager 지시 밖으로 더함 — 참조 수 검사가 실제로 일하는지 재려고.
- 재리뷰: Manager 지시대로 이번 호출에서는 gstack 재리뷰를 돌리지 않음(범위 확정 뒤).

## a166 census 조사 (읽기 전용, Manager 결정 5)

- 실패 시험: `internal/accountbudgetruntime/a166_wiring_census_test.go` `TestA166WiringCensusBaseline`(301행) → `wiringInspect`(96행) 의 모양 검사(209~217행) 「reviewed function shape changed」.
- corpus: `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`(고정 함수 모양 90개)·`sites.json`(분류된 호출 자리 116개), `loadWiringCorpus`(29행)가 읽음. 마지막 갱신 커밋 `2764aa701`(2026-10-31).
- 고정 모양 = `78d777cfe` 시점의 `runCommandWithOptions` 모양(같은 `format.Node`·Doc 제거 방식으로 뽑아 비교, 바이트 동일). 현재와의 diff 원문 `analysis/census-investigation/shape-diff.txt`(추출기 `shape.go.txt`).
  - 바뀐 것(전부 a179 입력 프로필 — `9abcb631a`, 병합 `d7c078ffc`): `--input-profile` 플래그 추가, `--out` 기본값 `""` + `resolveProfileOutput` 로 해석, `newJudgeClient` 를 플래그 파싱 뒤로 옮김, `cohort` 무작위 토큰 생성, `config{InputProfile, Cohort, …}`, 시작 로그 한 줄.
  - 예산 배선 줄(`ParseLaunchConfig`·`Open`·`factory.NewClient`·`WithPriority(ctx, accountbudget.Auxiliary)`)은 바뀌지 않음.
- census 자체의 판정(모양 검사 외 단계): HEAD 사본(git archive)에서 `shapes.json` 의 그 키 하나만 현재 모양으로 바꾼 뒤 census 시험 셋을 돌리면 **전부 PASS** — `TestA166WiringCensusBaseline`·`RejectsBypasses`(`new-unclassified-http` 는 의도대로 「callsite inventory count changed: got 117 expected 116」 로 잡힘)·`RejectsDirectorySymlink`. 즉 호출 자리 116개 분류(`sites.json`)는 그대로 맞고, HTTP 호출 자리가 전부 예산을 거친다는 census 판정은 a179 변경 뒤에도 성립함. 원문 `analysis/census-investigation/trial-shape-updated-copy.txt`. 저장소 파일은 바꾸지 않음.
- `RejectsBypasses/new-unclassified-http` 실패는 같은 원인의 부수 효과임: 변이가 기대한 「callsite inventory count changed」 보다 모양 검사 실패가 먼저 남.
- 고치려면: `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 키 `tools/a128-jev-shadow-probe/main.go||runCommandWithOptions` 값을 현재 모양으로 갱신(사람이 새 모양을 리뷰한 뒤). 다른 파일은 안 바꿔도 됨(사본 실측). 그 파일은 활성 change a166 의 시험 corpus 이고 비 Go 파일이라 gate ⑤ FLM 대상은 아님.
- 판단: 이 실패는 a179 가 병합 때 a166 의 모양 corpus 를 갱신하지 않아 생긴 것으로, 안전 결함이 아니라 corpus 낡음임(사본에서 다른 단계 전부 통과). 범위 결정은 사용자에게 올라가 있음.

## a166 census 수리 포함 (사용자 결정 2026-11-15 「추천대로」)

### Manager 리뷰 (원문 그대로)

`analysis/census-investigation/shape-diff.txt` 를 읽었다. `runCommandWithOptions` 의 변화는 전부 a179 입력 프로필이다: `--input-profile` 플래그, `--out` 기본값과 `resolveProfileOutput`, `newJudgeClient` 호출을 플래그 파싱 뒤로 옮김, cohort 무작위 토큰, `config{InputProfile, Cohort,…}`, 시작 로그 한 줄. 예산 배선 줄(`ParseLaunchConfig`·`Open`·`factory.NewClient`·`WithPriority(ctx, accountbudget.Auxiliary)`)은 diff 에 없다. 새 브로커 HTTP 호출 자리가 없고, 네 사본 실측에서 호출 자리 116개 분류가 그대로 통과했다. 판정: 모양 갱신 승인(Manager, 2026-11-15).

### 수리 기록 (구현 팀메이트 Opus)

- 갱신(커밋 3af104af9): `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 키 `tools/a128-jev-shadow-probe/main.go||runCommandWithOptions` 값 하나. 정식 생성 경로가 저장소에 없어 census 와 같은 추출로 만들고, 기존 값의 Go 인코딩 표기를 바이트 그대로 재현함을 먼저 단언함(`analysis/census-repair/enc.go.txt`). `git diff --numstat`: `1 1 shapes.json`. 키별 비교: 키 90/90, 추가·삭제 0, 바뀐 키 그 하나, 무변경 89, `sites.json` 116개 동일(`analysis/census-repair/shapes-key-diff.txt`).
- 반증(`analysis/census-repair/falsify.py`·`falsify.txt`, 사본·컴파일 확인·기대 문구 단언): 대조군 GREEN. `inside-pinned-function`(고정 함수 안 `http.Get`) → `TestA166WiringCensusBaseline` 「reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions」. `new-unpinned-function`(새 함수 안 `http.Get`) → 「callsite inventory count changed: got 117 expected 116」. 둘 다 CAUGHT — 갱신이 census 를 눈멀게 하지 않음.
- 시험(`analysis/verify-3/`): 4개 패키지 ok, FAIL 0. `internal/accountbudgetruntime` 최상위 PASS 53·서브 포함 77. `tools/a112-mb-us-source` 는 Manager 지시로 안 돌림(이전 `verify-2` 에서 ok 173s, 그 뒤 그 패키지가 보는 파일은 안 바뀜 — 바뀐 것은 shapes.json·두 도구 static_test.go 이고 측정 소스 신원 집합에는 `a112_mbus_static_test.go` 만 있음).
- 스펙 델타: census 때문에 바꿀 것 없음 — 무편집.

### 후속 (a166 인계)

- a166 은 a179 같은 병합 때 corpus(`testdata/a166-wiring/shapes.json`) 갱신이 필요하다.

## 범위 확장 2 — 낡은 시험 가드 3건 (사용자 결정 2026-11-16 「1번 — a189 에 전부 포함」)

### 수리 직전 재측정 (구현 팀메이트 Opus, 2026-11-16)

- 워킹트리(HEAD `fc078f1ea`)에서 세 시험 모두 FAIL — 문구는 Manager 조사와 같음(console_test.go:755, a163_judgement_pin_test.go:102 `0a776329… want f3458b84…`, strategy_dispatch_handoff_guard_test.go:669 `type 블록:Admit`).
- 원인 커밋과 부모에서 **실제 시험**으로 경계를 다시 잼(`analysis/boundary/boundary.sh`·`boundary.txt`, git archive 사본): `3dc0c776b^` ok / `3dc0c776b` FAIL, `9c4bdf4de^` ok / `9c4bdf4de` FAIL, `5c0aeb2b8^` ok / `5c0aeb2b8` FAIL.

### #1 orders seam (D11) — 철자 고정 대신 AST 단언을 고른 이유

- 한 줄: 철자 `l.shared.resolve(ctx)` 는 `(*lazyHoldings).Holdings` 에도 있어 파일 전체 검사로는 `Orders` 를 못 봄 — 변이 `o4-unlocked-cache`(잠금을 건너뛰고 공유 캐시를 직접 채움, 빌드 1회 유지)와 `o1-second-resolve` 가 철자만 고친 S 판본에서 PASS(못 잡음), AST 판본에서 FAIL.
- 판본 이력(gstack 리뷰): 1차(참조 수만)는 안 쓰는 참조 한 줄로 뚫림(1회차, 보이스 셋 일치) → 직접 호출이 broker 를 채우고 broker 대입은 그 문뿐 → range·함수 리터럴 가림·둘째 단언(2회차) → 단언 대상을 파서가 해소한 객체로 대조하고 단언은 하나만 → `&broker` 포인터 쓰기(3회차) → 다시 대입으로 셈. 최종 판본은 다섯 수(선언·참조·broker 를 채우는 직접 호출·broker 대입·broker 객체에 대한 유일한 단언)가 모두 1 이어야 함.
- 행동 절반(빌드 1회)·`official.New(` 금지는 무편집. 남는 한계는 design R7(단언 뒤 흐름·다른 파일의 `official.New(` — 편집 전부터 있던 한계).

### #2 a163 렌더 해시 (D12) — 갱신 이유

- 원인: a166 `9c4bdf4de`(「Classify protective priority and replay origin」) 가 `ExitObserver.record` 의 청산 정리 갈래에서 `clearTheSymbol` 에 넘기는 context 를 `accountbudget.WithPriority(ctx, Normal)`, `isProtective(proposal)` 이면 `Critical` 로 만듦(`analysis/render-hash/cause-9c4bdf4de-exitloop.diff`). `WithPriority` 는 `context.WithValue` 한 줄.
- 새 값의 출처: `9c4bdf4de^..7d1e3163a` rev-list 전수(135 커밋 + 부모 1)에서 `record` 렌더는 둘뿐 — `f3458b84…`(옛 고정값) 17 커밋, `0a776329…` 119 커밋. 부모 대비 바뀐 커밋은 `9c4bdf4de` 와 그것을 둘째 부모로 들인 병합 `5f681f2a5`·`5b379e08d` 뿐(`render-at-commits.txt`). 그래서 새 값 = 그 한 편집.
- 렌더 원문 diff(주석 없는 go/printer, 조건식·인자·대입식 포함): 차이는 clearCtx 를 만드는 네 줄 추가와 `clearTheSymbol` 호출 첫 인자 `ctx`→`clearCtx` 한 곳뿐. `9c4bdf4de`→base 0 바이트 차이. 판정 불변의 근거는 이 원문 diff 임(AST 항목 대조 — 분기 18→19, return 6→6, 빠진 항목 0 — 는 보조. gstack 1회차 Codex 지적으로 정정).
- 나머지 9개 고정값: 경계 세 지점에서 엔진 생산 파일 88개 전수 렌더로 기대값과 같음(1차 기록은 exitloop.go 만 렌더해 `noteHeldProposal` 을 안 쟀음 — 1회차 지적으로 정정, 옛 원문 `history-render-at-commits-exitloop-only.txt`).
- 시험 함수 본문 무편집 — 패키지 수준 변수 값 하나와 그 위 이유 주석(gate ⑤ required 수 변화 0 을 실측: 7 = 앞 로트 5 + #1·#3 둘).
- 변이: `r1-record-priority`(Critical→Normal)·`r2-record-revert-9c4bdf4de`(부모 모양) 둘 다 새 고정값이 거절.

### #3 admitSites (D13) — 좁힘과 잃은 검출

- 좁힌 뒤에도 잡는 것(전부 FAIL 확인): 둘째 `strategyhandoff.Admit` 호출(a1), 별칭 import 호출(a2), dot import 맨 이름 호출(a3), 둘째 `AdmitEachOwnerScope`(a4), 지역 함수 값(a5), 패키지 수준 함수 값(a6), 원시 문자열·이스케이프 import 경로(a7·a8 — 1회차 리뷰로 `strconv.Unquote`).
- 안 잡는 무해한 것(PASS 확인): 다른 타입의 같은 이름 메서드 `b.admission.Admit(ctx)`(b1 — 앞 판본은 이것을 자리로 셈, b1-OLD), 사설 인터페이스 메서드 선언(b2).
- census 밖 길을 막는 새 시험(리뷰로 더함, 새 파일 `strategy_dispatch_admit_census_a189_test.go`): 엔진 생산 파일 `unsafe` import 금지(a9 — 엔진 안 linkname), 모듈 전체 `//go:linkname` 지시어 금지(a10~a14 — 다른 패키지·testdata 중계·엔진 dispatch 대상·push 형태·디렉터리 심링크; 점 디렉터리는 무해 대조 b3).
- **잃은 검출(앞 판본은 잡고 새 판본은 못 잡음)**:
  - l1: 허용 자리 `dispatchHandoff` 안에서 문을 문 이름 필드로 빼돌리고 다른 자리에서 부름. 앞 판본은 이름이 같아 우연히 잡음(더해진 자리 4개). **남은 위험**(design R4).
  - l3: 엔진 밖 래퍼 패키지 `a189wrap` 이 `var Admit = strategyhandoff.Admit` 로 다시 내보냄. 새 판본에서는 strategyhandoff `TestOnlyTheEngineImportsThisSeam` 이 FAIL 로 막음.
  - 두 판본 다 못 잡는 것(l2): 같은 빼돌리기를 다른 이름(`mint`)으로 — 앞 판본의 l1 검출은 이름 우연이었다는 뜻.
- 인식기 변이 g1~g8 전부 기대 시험 FAIL.
- 원문 `analysis/mutation-2/mutation2-run7.txt`: 대조군 3 + 변이 51 = 54 줄 전부 기대대로, 원복 sha 일치, 새 파일 0. 하네스 결함 이력(run1 g1 대상 3회·OLD 대조군 비교 불가, run2 EXTRA 판정이 합계에 안 더해짐, 리뷰 2회차 중간 실패 시 원복)은 원문·이력 파일로 남김.

### gstack 리뷰 (tasks 3c.8) — 3회, 수렴 안 함

- 원문·처리 표 `analysis/gstack-review-2.md`. 보이스: 전문가 6(testing·maintainability·security·performance·simplification·state-machines), red team, Claude 적대 패스, **Codex 적대 패스(다른 계열, read-only 샌드박스)** 1·3회차.
- 1회차 9건(고침 6·기록 3), 2회차 9건(고침 6·기록 2·Manager 1), 3회차 6건(고침 5·Manager 1). CRITICAL 로 분류된 것은 Manager 결정 대상 하나(별칭/dot import 로 보조 등급 고정 우회 — 앞 로트 범위, design R1).
- 3회차 수리 뒤 4회차 리뷰는 skill 상한(편집 회차 3)으로 돌리지 않음 — 3회차 수리는 변이(a11~a14·b3·o14)로만 확인했고 **새 발견 0 은 확인되지 않음**.

### not-applicable

- 다중 적대 리뷰: 이번엔 Manager 지시로 생략 불가 — gstack `review` + 다른 계열 보이스(Codex) 를 붙임(위).
- 스펙 델타: 세 가드를 기술한 정본 문장 없음(proposal Impact) — 무편집.
- `make gate`·전체 시험·`make test-seams` 전체: 지시대로 안 돌림. 태그 seam 은 바뀐 가드 시험만 `-tags tossos_testseams` 로 돌림(`verify-4/go-test-testseams-targeted.txt`).
- Exploratory QA: 시험 전용 diff 라 기능 표면 없음 — 대상 패키지 시험·변이 하네스가 판정 근거.

### 후속 (기록 — 두 change 디렉터리는 수정하지 않음)

- a166(`3dc0c776b`, `9c4bdf4de`)과 a087(`5c0aeb2b8`)이 전체 시험 없이 main 에 들어와 시험 가드 3건이 낡은 채 남았음. 같은 기전이 앞 로트의 가드 5건·census 1건(a166 `78d777cfe`·`4d5884cbc`, a179 `9abcb631a`)에도 있었음 — 병합 전 영향 패키지 시험이 빠짐.
- `TestEveryNameTheEngineTakesFromTheSeamIsClassified`(이 diff 가 안 바꾼 기존 함수)의 import 경로 읽기도 `strings.Trim` — 원시 문자열·이스케이프 철자에 눈멂. 주조 자리는 admitSites 수리로 잡히지만, 분류 검사 자체의 수리는 FLM 번들 하나가 더 드는 별 작업.
- 앞 로트 `census-repair/falsify.py` 의 HOME 격리 없음(잠재만 — 그 시험은 HOME 을 안 읽음).

### Manager 에게 올림 (범위 확장 2)

1. **(High, 안전 불변식 4) 두 그림자 도구의 보조 등급 고정이 별칭·dot import 로 뚫림**(gstack 2·3회차 red team 실측, design R1 「재검토 필요」). 앞 로트 Manager 결정(2026-11-15, R1 범위 밖)은 보조 등급 고정(ab7023901)보다 앞섬.
   - (A) a189 에서 고침 — 두 도구의 import 시험에서 accountbudget·accountbudgetruntime(·official·tossconfig) import 의 `spec.Name` 이 nil 이 아니면 실패. 두 시험 함수 본문 편집(번들 갱신), 변이 추가. (추천)
   - (B) R1 결정 유지, 후속 change.
2. R4 l1(허용 자리 안 문 이름 필드 빼돌리기): (A) 기록만(현재) (B) 후속 change 로 「경계 문 참조는 호출의 Fun 이어야 함」 — 기존 fixture 시험의 함수 값 철자 셋을 「거절」로 바꾸는 계약 변경.
3. R5(표지 없는 context 는 Normal)·R6(reflect 이름 우회): 스펙 문장을 가드가 재는 만큼 좁힐지, 런타임 단언을 넓힐지.
4. gstack 리뷰 비수렴: 4회차(3회차 수리에 대한 재리뷰)를 돌릴지.
5. Gate 처분 3건(`gstack-review-2.md` 「Gate edits」): 추천 B(Keep — justified, verified).

## Manager 결정 반영 (2026-11-16, 범위 확장 2 마무리 로트)

### 결정 원문 요지와 처리

1. **결정 1 = (A) a189 에서 고침**: a128·a140 의 `TestImportsStayInsideTheAllowlist` 가 `official`·`accountbudget`·`accountbudgetruntime` import 의 `spec.Name` 이 nil 이 아니면(별칭·dot·blank 전부) 실패함. 별칭을 쫓지 않고 금지함. 두 시험 함수 FLM 번들 갱신(편집 전 AST `pre-edit-ast/*.before-decision1.ast.json`). 오늘 두 도구 비시험 소스에서 세 패키지를 이름 붙여 import 한 자리는 0건(grep) — 정상 입력 거절 0.
   - 변이(`analysis/mutation-2/mutation2-run10.txt`): d1a(둘째 파일 별칭 + `ab.WithPriority(ctx, ab.Critical)`)·d1b(dot + `WithPriority(ctx, Critical)`)·d1c(blank) 를 두 도구마다 — 전부 `TestImportsStayInsideTheAllowlist` FAIL(문구 「is renamed to …」). 결정 전 판본(PRE1 — 금지 갈래만 뺀 사본)에서는 d1a·d1b 가 import·이름 시험 둘 다 PASS(못 잡음). 무변이 대조군 GREEN(두 도구 import·이름 시험 PASS).
2. **결정 2**: `TestA189EngineProductionNeverImportsUnsafe`·`TestA189NoPackageLinknamesTheSeamDoors` 삭제, 새 파일 머리 주석의 (마)(바)와 안 쓰게 된 import(`os`·`path/filepath`·`strconv`) 정리. 이유(Manager 원문 그대로):
   - 요청 범위는 낡은 가드 3건 수리다. 모듈 전체 `//go:linkname` 금지·엔진 `unsafe` 금지는 아무도 요청하지 않은 새 저장소 정책이고, 리뷰 루프가 만든 규칙이다.
   - linkname 경로는 좁힘으로 잃은 검출이 아니다(이름 기반 앞 판본도 이름이 다르면 못 셌다) — 편집 전부터 있던 맹점이다.
   - 모듈 전체 walk 는 환경에 따라 실패한다(비-점 디렉터리 심링크, `scanned < 500` 문턱). 엔진 스위트가 seam 과 무관한 이유로 빨개지는 길을 새로 만든다.
   - 변이 a9~a14 는 「잡지 못함(기존 맹점, 범위 밖)」 으로 재분류하고 하네스를 다시 돌려 원장이 실제 결과와 같게 함(run10: 여섯 다 census PASS 로 기록). 그 시험만 겨누던 무해 대조 b3 는 뺌. `strconv.Unquote` 해독과 `TestA189AdmitCensusDecodesTheImportPathLiteral` 은 유지.
3. **결정 3**: R4(l1)·R5·R6 은 기록만 함(아래 「남은 위험」). 가드를 넓히지 않음. 스펙 무편집.
4. **결정 4 — Manager 가 리뷰 루프를 닫음**: gstack 4회차는 돌리지 않음. 사유: 시험 전용 diff(운영 Go 0줄)라 비례 원칙상 경량 리뷰가 기본이고, 3회차까지의 새 발견은 모두 「가드를 의도적으로 우회하는 저자」 모양이며 남은 것은 아래 「남은 위험」에 열거됨. **수렴하지 않았고 닫았음** — 「새 발견 0 수렴」 을 주장하지 않음.
5. **결정 5**: gate 처분 3건(a163 고정값 `49234f885d5c` · admitSites 좁힘 `aad7cf1b05f0` · shapes.json `c98da0370da7`) 모두 Keep(B), Manager 승인.
6. **결정 6**: #1 콘솔 시험의 다섯 수 AST 단언 유지(Manager 가 실제 `(*lazyOrders).Orders` 본문과 대조해 확인).

### 남은 위험 (범위 확장 2 최종)

- R8 linkname/unsafe 로 census 밖에서 경계 문에 닿는 길 — 편집 전부터의 맹점, Manager 결정으로 범위 밖, 후속(변이 a9~a14 「잡지 못함」).
- R4 l1: 허용 자리 안에서 문을 문 이름 필드로 빼돌림 — 좁힘으로 잃은 검출(앞 판본은 이름 우연으로 잡음). 기록만, 후속.
- R5: 표지 없는 context 는 `Normal` 로 돎 — 보조 등급 고정은 명시 `WithPriority` 표지만 잼. 기록만, **a166 후속으로 넘김**.
- R6: 예산 핸들·주문 이름 금지는 선택자 이름 단위라 `reflect.MethodByName` 으로 비켜 감(기존 한계). 기록만.
- R7: 콘솔 단언 뒤 흐름(`reads`·`plain` 다시 대입)과 다른 파일의 `official.New(` — 편집 전부터의 한계, 후속.
- R1 잔여: `tossconfig`(internal/config) 별칭은 결정 1 범위 밖 — 기존 한계 그대로.

### 스펙 델타 중 가드가 검증하지 않는 것을 단언하는 문장 (고치지 않음 — Manager 에게 올림)

- `specs/breakout-retest-strategy-lane/spec.md:10` 「우선순위는 보조 등급만 써야 한다 (SHALL)」·「측정 도구의 정적 검사는 … 우선순위 인자가 보조 등급임을 확인해야 한다 (SHALL)」: 정적 검사가 확인하는 것은 **명시적 `accountbudget.WithPriority` 직접 호출의 둘째 인자**뿐임. 표지 없는 context 의 기본값 `Normal`(transport_linux.go:24~29)로 도는 요청은 검사 밖(R5). 결정 1 로 별칭·dot 우회는 닫힘.
- 같은 줄 「측정 도구는 예산 핸들을 직접 잡거나 관측 상태를 읽어서는 안 되고 (MUST NOT)」: 검사는 메서드 이름 7개의 선택자 금지라 `reflect` 문자열 호출은 못 봄(R6).

## Manager 재검증 (3.4) — 2026-11-16

Manager(Fable) 가 HEAD `974a9c623` 에서 한 독립 재검증. 원문은 `analysis/verify-manager-2/`(Manager 스크래치패드에서 원문 그대로 복사, 편집 없음 — 종결 팀메이트 Opus 가 옮김).

- 전체 시험을 격리 worktree `~/tossos-work/a189-fulltest`(detached `974a9c623`)에서 실행: `make test` rc 0(ok 122 패키지, FAIL 0) · `make test-seams` rc 0(ok 123, FAIL 0) · `make test-race` rc 0 · `make vet` rc 0. 실행 07:21:59~07:43:27 KST, 전후 HEAD 동일·dirty 0. 원문 `full2-status.txt`·`full2-test.txt`·`full2-test-seams.txt`·`full2-test-race.txt`·`full2-vet.txt`.
- 이전에 빨갛던 패키지가 모두 ok: `cmd/tossctl` · `internal/app/engine` · `internal/accountbudgetruntime` · `internal/official` · `tools/a128-jev-shadow-probe` · `tools/a140-breakout-signal-shadow`(`full2-test.txt` 2·4·6·51·127·130행).
- `tools/logic-map/check_analysis.py --change a189-shadow-tool-guards-admit-the-account-budget`: required 7, rc 0. 원문 `a189-mgr-check-analysis.txt`.
- `openspec validate --strict`: valid.
- base 대비 비시험 Go 변경 0 파일.
- Manager 독립 변이 6개(팀메이트 하네스와 별개, 사본 worktree, 원복 뒤 dirty 0), 원문 `a189-mgr-mutation2.txt`:
  - M1 엔진 생산 파일에 둘째 `strategyhandoff.Admit` 참조 → `TestExactlyOneProductionSiteAdmitsIntoTheSeam` FAIL(기대 FAIL).
  - M2 같은 이름 인터페이스 메서드 선언 → PASS(기대 PASS).
  - M3 `record` 의 clearCtx Critical→Normal → `TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange` FAIL.
  - M4 `(*lazyOrders).Orders` 에 resolve 둘째 호출 → `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` FAIL.
  - M5 a128·a140 에 별칭 import → 두 도구 `TestImportsStayInsideTheAllowlist` FAIL.
- Manager 결정(스펙 문장 2건, 위 절): 고치지 않음. 스펙 문장은 도구가 지켜야 할 요구이고 정적 검사는 그 일부를 강제함. 강제하지 못하는 부분은 남은 위험 R5·R6.
- tasks 3.2 의 `tools/a112-mb-us-source` 는 위 Manager 전체 시험(`make test`)에 포함돼 ok 임. 종결 팀메이트가 `full2-test.txt` 125행을 직접 확인함: `ok  	github.com/JungHoonGhae/tossinvest-cli/tools/a112-mb-us-source	(cached)` — go 시험 캐시 적중(같은 입력의 앞선 통과 결과 재사용)으로 찍힌 ok 임. 이 근거로 3.2 와 3.4 를 체크함.
