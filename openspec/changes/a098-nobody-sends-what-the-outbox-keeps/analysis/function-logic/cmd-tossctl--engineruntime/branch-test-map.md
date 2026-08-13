# Branch Test Map: `engineRuntime`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: **편집 후** 분기 6 · 이탈 8 (편집 전 5 · 7).

기존 커버리지의 정본은
`TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess`
(`cmd/tossctl/engine_runtime_branch_test.go:49`)의 케이스 표다 — 다섯 갈래를
**하나씩 실패시켜** 본다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:349` 체결 감지 구성 실패면 거절 | 기존 분기 테이블 | no | **yes (기존)** |
| B2 | `:358` 대사 드라이버 구성 실패면 거절 | 같은 테스트 | no | **yes (기존)** |
| B3 | `:369` exit 관측 구성 실패면 거절 | 같은 테스트 | no | **yes (기존)** |
| B4 | `:374` 복구 시퀀스 구성 실패면 거절 | 같은 테스트 | no | **yes (기존)** |
| B5 | `:378` `strategy-entry` 구성 실패면 거절 | 같은 테스트 | no | **yes (기존)** |
| B6 | **신설** `:390` — **배달 실행자 구성 실패면 거절**(fail-closed) | — | — | **아래 ⛔ — 덮이지 않았다** |
| 이탈 `:403` | 정상 조립 — 감독 넷 + **보조 하나** | **a098** `TestProductionRuntimeStartsExactlyOneAlertDeliverer` · `TestTheAlertDelivererIsNotASupervisedLoop` · 기존 `TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop` | **yes — 뮤테이션 L**: 실행자를 만들고 `Auxiliary`에 안 넘기면 `auxiliary names=[]`로 FAIL | **yes (2026-08-12)** |

> **⛔ B6의 거짓 쪽만 덮였다 — 참 쪽(거절)은 테스트가 없다. 숨기지 않는다.**
>
> `AlertDeliverer`가 거절하는 조건은 *"`Journal` 이나 `Entry` 가 nil"*뿐인데,
> **그 상태의 `engine.Context`는 이 함수에 오기 전에 이미 거절된다** — B2의
> `ReconcileDriver`도 B4의 `Recovery`도 같은 핸들을 요구하고 먼저 실행된다.
> 즉 B6의 참 쪽은 **현재 조립 순서에서 도달 불가능**하다.
>
> **그렇다고 갈래를 지우지 않는다.** 순서가 바뀌면 도달 가능해지고, 그때
> **없는 것은 갈래가 아니라 거절**이다. `not-applicable`로 적는 대신
> **덮이지 않았다고 적는다** — 이 둘은 다르다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

B1~B5 — 다섯 자리 전부 기존 분기 테이블이 덮고 **a098은 그 조건을 안 바꾼다.**
§5의 VERIFY는 그 다섯의 판정이 안 바뀌었다만 확인한다.

## 덮이지 않은 것을 이름으로 적는다

- **B6의 참 쪽** — 위 ⛔.
- **`Loops` 리터럴의 내용** — 이 change 가 안 바꾸고, 바뀌지 않았음을 **R12**가 진다
  (`engine_strategy_entry_dormant_test.go:50`의 `reflect.DeepEqual`).
  `TestTheLoopSetIsTheSpecifiedThree`(`engine_test.go:347`)는 **그 증거가 아니다** —
  소스 문자열 포함만 보므로 루프가 넷이어도 통과한다.
