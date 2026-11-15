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
