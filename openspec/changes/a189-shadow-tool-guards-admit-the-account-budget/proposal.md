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

### `internal/official/a112_mbus_static_test.go`

- 선택지 A(Manager 결정 2026-11-15, design D4): `a112MBUSReferenceAllowed` 의 마지막 정확 경로 비교에 `internal/accountbudgetruntime/v2_token_test.go` 한 경로를 더함. 접두사·접미사 조건은 그대로임. 그 파일 하나만 통과하고 다른 모든 파일의 판정은 그대로임.

### 정본 스펙

- `breakout-retest-strategy-lane` 「breakout 그림자 신호 측정은 주문 권위가 없다」를 MODIFIED: 측정 도구가 계좌 공유 요청 예산에 보조 등급으로 참여할 수 있고, 그 표면은 정적 검사가 패키지별 이름으로 고정하며, 주문·계좌·보유·매수 가능 표면 금지는 그대로임을 적음. 시나리오 하나를 더함.
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

- 편집 파일: 시험 파일 3개(위). `.go` 비시험 파일 0개.
- 고치는 기존 함수(gate ⑤ FLM 번들 대상): 선택지 A 기준 **5개** — `a112MBUSReferenceAllowed`, a128 `TestNoMutatingOrAccountSurfaceIsNamed`·`TestImportsStayInsideTheAllowlist`, a140 `TestNoMutatingAccountOrWriteSurfaceIsNamed`·`TestImportsStayInsideTheAllowlist`. 편집 전 AST 는 `analysis/pre-edit-ast/`.
- 판정이 바뀌는 시험: 위 5개가 실패 → 통과. 같은 파일의 다른 시험과 가드를 부르는 픽스처 시험(`TestA112MBUSResolvedASTGuard*` 3개)은 판정 불변이어야 함.
- 런타임·운영 영향 0: 시험 파일만 바뀜. 운영 토글·설정·DB·journal 변경 없음.
