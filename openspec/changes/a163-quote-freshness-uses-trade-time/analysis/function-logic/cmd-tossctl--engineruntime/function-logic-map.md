# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (655-752)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 6 · 반환 7 · defer 0 · go 0
- 위험 등급: High-risk(생산 엔진 런타임 조립 — 청산 관측자 배선)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 검증된 엔진 컨텍스트 | `engine.NewContext` | — |
| `logger` | 엔진 로거 | 호출자 | nil 이면 전용 줄 0 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 658 `if err != nil {` | — | 탐지기 조립 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B2 | if 667 `if err != nil {` | — | 대사 드라이버 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B3 | if 684 `if err != nil {` | — | `ectx.ExitObserver` 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B4 | if 689 `if err != nil {` | — | 복구 조립 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B5 | if 693 `if err != nil {` | — | 전략 진입 감독 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B6 | if 705 `if err != nil {` | — | 다음 조립 실패 반환 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |

반환 좌표(AST): 659 · 668 · 685 · 690 · 694 · 706 · 709

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.ExitObserver` | 청산 관측자 조립(옵션 리터럴 하나) | 오류는 B3 | AST |
| `engine.NewRuntime` | 감독자 | — | AST |

## State mutations and fallbacks

- 없음(조립만).

## a163 편집

- a163 3.4: `engine.ExitObserverOptions` 리터럴에 `SourceAgeLog: logger` 한 키. 분기 6 · 반환 7 불변. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 옵션 리터럴 키 하나.
- High-risk impact: yes(배선) — 구조 시험 2.11(a) · 변이 ⑫.
