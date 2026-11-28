# Proposal

## Why

main HEAD `7d1e3163a`(= origin/main)에서 `go test ./...` 가 정적 가드 5개로 실패함. 완료 게이트 7단계 `make test` 를 어떤 change 도 통과할 수 없고, 그래서 a187 archive 가 막혀 있음. 실패 원문은 `openspec/changes/a187-strict-quote-readers-accept-kr-alphanumeric-symbols/analysis/verify/base-failures.txt` 에 있음.

1. `internal/official` `TestA112MBUSStaticNoProductCallerOrForbiddenPath` — `internal/accountbudgetruntime/v2_token_test.go: selector official.A112MBUSCandle`
2. `tools/a128-jev-shadow-probe` `TestImportsStayInsideTheAllowlist` — `internal/accountbudget`·`internal/accountbudgetruntime` import 가 허용 목록 밖
3. `tools/a128-jev-shadow-probe` `TestNoMutatingOrAccountSurfaceIsNamed` — `official.Option` 이 허용 목록 밖
4. `tools/a140-breakout-signal-shadow` `TestImportsStayInsideTheAllowlist` — 같은 두 import
5. `tools/a140-breakout-signal-shadow` `TestNoMutatingAccountOrWriteSurfaceIsNamed` — `official.Option`

원인 커밋은 `78d777cfe`(feat(a166): shadow 도구 예산 연결)와 `4d5884cbc`(Bind runtime v2 event store and token observations), 둘 다 2026-10-31. a166 이 두 그림자 도구의 `main.go` 를 계좌 공유 요청 예산에 연결하고 A112 M-B0 복제 전송이 예산에 계수되는지 보는 런타임 시험을 넣었지만, 세 가드의 허용 목록은 그대로 둠. 가드 파일을 고친 커밋은 어느 브랜치에도 없음(`git log --all --not main -- <세 가드 파일>` 0건, Manager 실측).

사용자 결정 2026-11-15: 「별도 작은 change 로 가드를 고치고 gstack 리뷰 뒤 a187 을 archive」.

추가로 같은 날 영향 패키지 시험에서 6번째 실패를 찾았음: `internal/accountbudgetruntime` `TestA166WiringCensusBaseline`·`TestA166WiringCensusRejectsBypasses/new-unclassified-http` — `reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions`. base `7d1e3163a` 에서도 같음. a179 입력 프로필(`9abcb631a`, 병합 `d7c078ffc`)이 그 함수를 바꾸면서 a166 census 의 고정 모양(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)을 갱신하지 않았음. 사용자 결정 2026-11-15(「추천대로」): 이 수리를 a189 에 포함.

### 범위 확장 2 — 낡은 시험 가드 3건 (사용자 결정 2026-11-16 「1번 — a189 에 전부 포함」)

a189 의 목적(main 의 `go test ./...` 를 다시 초록으로)을 막는 가드가 셋 더 있음. 셋 다 시험 파일의 고정값·인식기가 운영 코드의 의도된 편집을 따라가지 못한 것임. 각 원인 커밋과 그 부모에서 시험을 실제로 돌려 PASS→FAIL 경계를 확인함(`analysis/boundary/boundary.txt`).

7. `cmd/tossctl` `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` — 문자열 고정 `l.shared.resolve()` 가 a166 `3dc0c776b`(resolve 에 `ctx` 인자 추가) 뒤로 소스에 없음.
8. `internal/app/engine` `TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange` — `ExitObserver.record` 의 렌더 해시. a166 `9c4bdf4de` 가 `record` 에 예산 우선순위 표지를 넣고 고정값을 갱신하지 않음.
9. `internal/app/engine` `TestExactlyOneProductionSiteAdmitsIntoTheSeam` — 인식기 `admitSites` 가 이름이 `Admit` 인 식별자를 전부 세서 a087 `5c0aeb2b8` 의 인터페이스 메서드 선언 `lowerLimitAdmission.Admit`(exit_lower_limit.go:58)을 경계 자리 `type 블록:Admit` 으로 셈.

## What Changes

시험 파일의 가드 허용 목록만 고침. 더하는 이름은 지금 비시험 소스가 실제로 쓰는 이름(AST 열거, `analysis/enumeration/selectors-and-imports.txt`)뿐임.

### `tools/a128-jev-shadow-probe/static_test.go`

- `TestImportsStayInsideTheAllowlist` 의 `allowedModule` 에 두 경로를 더함:
  - `github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget`
  - `github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime`
  - 실패 문구의 괄호 목록도 이에 맞게 고침.
- `TestNoMutatingOrAccountSurfaceIsNamed`:
  - `allowedOfficial` 에 `Option` 하나를 더함.
  - 새로 들이는 두 패키지의 선택자 이름을 고정함(같은 파일 `TestWTSSourceIsTheThreeReads` 의 `allowedPackage`+`saw` 모양 그대로):
    - `accountbudget`: `WithPriority`, `Auxiliary`
    - `accountbudgetruntime`: `ParseLaunchConfig`, `Open`
  - 두 패키지 중 하나라도 선택자를 하나도 못 보면 실패함(기존 `sawOfficial`·`sawConfig` 와 같은 규칙).

### `tools/a140-breakout-signal-shadow/static_test.go`

- `TestImportsStayInsideTheAllowlist` 의 `allowed` 에 `accountbudget`, `accountbudgetruntime` 를 더함.
- `TestNoMutatingAccountOrWriteSurfaceIsNamed` 의 `allowed` 맵:
  - `official` 목록에 `Option` 을 더함.
  - 새 항목 `accountbudget`: `WithPriority`, `Auxiliary` / `accountbudgetruntime`: `ParseLaunchConfig`, `Open`. 기존 `for pkg := range allowed { if !saw[pkg] … }` 가 두 패키지의 「본 것 없음」도 자동으로 실패시킴.

### 보강 (Manager 결정 2026-11-15, gstack 발견 1)

- 두 도구의 이름 시험이 `accountbudget.WithPriority` 를 직접 호출로만 받고 둘째 인자가 정확히 `accountbudget.Auxiliary` 인지 고정함. 호출이 0 이거나 함수 값 참조가 있으면 실패.
- 두 도구의 `forbidden` 에 예산 핸들 메서드 7개를 더함: `Budget`, `Acquire`, `EnginePollKey`, `NewObservationFence`, `ObservationSnapshot`, `PollRemaining`, `VerifyIdentity`.

### a166 census 고정 모양 (사용자 결정 2026-11-15)

- `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 에서 키 `tools/a128-jev-shadow-probe/main.go||runCommandWithOptions` 의 값 **하나만** 현재 모양으로 갱신함. 다른 89개 키와 `sites.json` 은 무변경.

### `internal/official/a112_mbus_static_test.go`

- 선택지 A(Manager 결정 2026-11-15, design D4): `a112MBUSReferenceAllowed` 의 마지막 정확 경로 비교에 `internal/accountbudgetruntime/v2_token_test.go` 한 경로를 더함. 접두사·접미사 조건은 그대로임. 그 파일 하나만 통과하고 다른 모든 파일의 판정은 그대로임.

### 낡은 시험 가드 3건 (범위 확장 2, design D11~D13)

- `cmd/tossctl/console_test.go` `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`: 파일 전체 문자열 검사 `strings.Contains(src, "l.shared.resolve()")` 를 `(*lazyOrders).Orders` 본문 안의 `<수신자>.shared.resolve` 참조 수 == 1 AST 단언으로 바꿈(D11). 나머지(빌드 횟수 1, `official.New(` 금지)는 그대로.
- `internal/app/engine/a163_judgement_pin_test.go`: 패키지 수준 `a163JudgementRenderHashes["record"]` 값 하나를 a166 `9c4bdf4de` 의 렌더로 갱신하고, 가드 주석이 요구하는 이유를 그 자리에 적음(D12). 나머지 9개 값과 시험 함수 본문은 그대로.
- Manager 결정 1(2026-11-16, gstack 리뷰 발견): `tools/a128-jev-shadow-probe/static_test.go`·`tools/a140-breakout-signal-shadow/static_test.go` 의 `TestImportsStayInsideTheAllowlist` 가 `official`·`accountbudget`·`accountbudgetruntime` 를 별칭·dot·blank 로 import 하면 실패함(이름 시험이 선택자 철자로 패키지를 가리므로 별칭으로 Critical 우선순위를 달 수 있었음).
- `internal/app/engine/strategy_dispatch_handoff_guard_test.go` `admitSites`: 경계 패키지(import 경로 끝 `/strategyhandoff`)를 가리키는 지역 이름(별칭 포함)의 선택자 참조와, dot import 가 있는 파일의 선택자 아닌 문 이름만 셈(D13). 새 fixture 시험은 새 파일에 둠.

### 정본 스펙

- `breakout-retest-strategy-lane` 「breakout 그림자 신호 측정은 주문 권위가 없다」를 MODIFIED: 측정 도구가 계좌 공유 요청 예산에 보조 등급으로 참여할 수 있고(허용이지 보장 아님), 우선순위는 보조 등급만 쓰며, 정적 검사가 이름과 우선순위 인자를 고정함. 예산 참여가 요청마다 예산의 관측 상태를 기록하되 그것은 측정 evidence 가 아님을 적음(gstack 발견 9). 주문·계좌·보유·매수 가능 표면 금지는 그대로. 시나리오 하나를 더함.
- `breakout-retest-strategy-lane` 「M-B0/M-B1 measurement는 production evidence authority와 분리된다」를 MODIFIED(A-2): 「M-B0 exported-symbol references are allowed only …」 한 문장에 a166 런타임 시험 파일 하나를 더함. 다른 문장은 무편집(`analysis/spec-a2-mb0-diff.txt`).

## Capabilities

### New Capabilities

없음

### Modified Capabilities

- `breakout-retest-strategy-lane`: 그림자 측정 도구의 계좌 공유 요청 예산 참여와 그 표면의 이름 고정, M-B0 참조 허용 위치에 a166 런타임 시험 파일 하나 추가.

## Non-goals

- 도구의 `main.go` 등 비시험 코드, `internal/accountbudget*`, 주문·인증 경로. a166 의 배선은 의도된 것이므로 되돌리지 않음.
- 금지 이름 목록(`forbidden`)과 기존 검사의 약화. 하나도 빼지 않음.
- 허용 목록에 이미 있지만 지금 안 쓰는 이름의 정리(a128 `official.New`·`Client`·`Credentials`, a140 `official.New`·`Client`). 좁히기는 이 change 가 요청받은 범위가 아님.
- 별칭 import 로 패키지 이름 추적을 비켜 가는 경로 막기(design Risks R1). `official`·`tossconfig` 에도 이미 있는 한계라 범위 밖(Manager 결정 2026-11-15).
- a128 프로브의 요구사항 신설. 정본 스펙에 a128 프로브의 정적 검사 문장이 없음(`semantic-judgment-shadow` 검색 결과는 평가 해석 문장 3건뿐, review.md) — 지어내지 않음.

## Impact

- 편집 파일: 시험 파일 3개와 새 시험 파일 1개(`a112_mbus_a189_static_test.go`), a166 census corpus `shapes.json` 의 값 하나. `.go` 비시험 파일 0개.
- 고치는 기존 함수(gate ⑤ FLM 번들 대상): 선택지 A 기준 **5개** — `a112MBUSReferenceAllowed`, a128 `TestNoMutatingOrAccountSurfaceIsNamed`·`TestImportsStayInsideTheAllowlist`, a140 `TestNoMutatingAccountOrWriteSurfaceIsNamed`·`TestImportsStayInsideTheAllowlist`. 편집 전 AST 는 `analysis/pre-edit-ast/`.
- 판정이 바뀌는 시험: 위 5개와 a166 census 2개(`TestA166WiringCensusBaseline`·`RejectsBypasses`)가 실패 → 통과. 같은 파일의 다른 시험과 가드를 부르는 픽스처 시험(`TestA112MBUSResolvedASTGuard*` 3개)은 판정 불변이어야 함.
- 런타임·운영 영향 0: 시험 파일만 바뀜. 운영 토글·설정·DB·journal 변경 없음.
- 범위 확장 2: 시험 파일 3개(`console_test.go`, `a163_judgement_pin_test.go`, `strategy_dispatch_handoff_guard_test.go`)와 새 시험 파일 1개. 운영 Go 파일 0줄. 본문을 고치는 기존 함수는 `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`·`admitSites` 둘로 예상함(`a163JudgementRenderHashes` 는 패키지 수준 변수 — 수는 `check_analysis.py` 로 실측). 판정이 바뀌는 시험: 위 세 시험이 실패 → 통과. `ExitObserver.record` 등 운영 청산·주문 진입 경로는 무편집.
- 스펙: 범위 확장 2 는 시험 가드의 고정값·인식기 정정이라 Requirement 문장을 바꾸지 않음. 정본 스펙에 세 가드를 기술한 문장이 없음(`openspec/specs` 에서 `render hash`·`admitSites`·`ExactlyOneProductionSite`·`orders seam`·`shared resolver` 검색 0건). exit-policy 「브로커 데이터 시각은 보고의 입력이며 청산 판정 자격을 바꾸지 않는다」의 금지는 브로커 시각·보고 토글을 읽는 것에 대한 것이고, 그것을 지키는 금지 census 두 시험은 이 change 에서 무편집임.
