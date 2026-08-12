# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (L199-312)
- AST evidence: `ast.json` — 분기 4, return 6
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

a100은 **이 함수를 편집한다.** 수렴 워커(design D3)가 조립되는 지점이 여기다. 기존 함수 내부를
바꾸므로 `.claude/CLAUDE.md` 「필수 진입과 완료 조건」 3항에 따라 편집 전에 이 산출물이 있어야
한다.

## AST가 말해 준 것 — 이 함수에는 조건부 조립이 하나도 없다

분기 4개는 **전부 에러 검사**다.

| Branch | 종류 | 조건 |
|---|---|---|
| B1 (L204) | 에러 | `checkProjectionWired(in.journal)` 실패 |
| B2 (L226) | 에러 | `tracker.Restore(ctx)` 실패 |
| B3 (L250) | 에러 | `protection.NewPairedReadinessAdapter` 실패 |
| B4 (L274) | 에러 | `execgw.New` 실패 |

토글·설정으로 갈라지는 분기는 **없다.** 모든 구성요소가 무조건 생성되고, 켜고 끄는 일은
생성된 뒤 다른 곳에서 판단한다(`Attested: nil` 주석 L269-272이 그 관례를 말한다 — "Turning
replay on is 2b's job and it is **one field**").

⇒ **수렴 워커를 `adoption.enabled` 같은 조건 아래 조립하면 이 함수의 첫 조건부 조립 분기가
된다.** 관례를 깨는 편집이므로 그 선택을 설계에서 명시해야 한다. 관례를 따르려면 워커는
무조건 생성하고 **실행 여부를 워커 자신이 판단**해야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.journal` | apply hook가 바인딩된 journal | `bindApplyHooks`(L152) | B1이 미바인딩을 거부 |
| `in.official` | official Open API client | 상위 조립 | nil이면 이후 호출에서 패닉 — 이 함수는 검사하지 않는다 |
| `in.accountRef` | 계좌 참조 | 자격증명 해석 결과 | 검사 없음 |
| `in.clock` | 엔진 시계 | 상위 조립 | 검사 없음 |
| `in.configDir` | 설정 디렉터리 | 상위 조립 | readiness provider가 읽는다 |
| `in.manifestPin` | manifest digest | 상위 조립 | readiness가 검증 |

**불변식 1 — 순서가 의미를 갖는다.** 주석이 두 번 명시한다. L200-203: journal에서 믿음을
읽기 **전에** 그 믿음의 생산자가 있어야 한다. L278-279: Retrier가 Notifier를 통해 전이를
알리므로 **Notifier가 먼저** 존재해야 한다. 새 구성요소를 끼워 넣을 때 이 두 제약을 깨면
컴파일은 되고 런타임이 깨진다.

**불변식 2 — RECONCILE projection 복원이 블록 집합을 되살린다.** L216-219: `tracker.Restore`
없이 재기동하면 "silently clears every block a disagreement raised". 수렴 워커가 tracker보다
먼저 돌면 아직 복원되지 않은 블록 상태를 보게 된다.

**불변식 3 — 보호 readiness는 이미 여기서 조립된다.** L242-252가
`productionProtectionAssemblies` → `NewProductionProvider` → `NewPairedReadinessAdapter`를
만들고 L261이 그것을 `execgw`에 넘긴다. **a100은 이 경로를 바꾸지 않는다** — readiness는
진입(exposure 증가) 판정용이고, a100이 다루는 것은 reduce-only 매도라 `checkProtection`이
즉시 반환하는 쪽이다(design 「범위를 이렇게 자를 수 있는 이유」).

## Branches and early returns

| Branch | 조건 (L) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `checkProjectionWired` 실패 (204) | 없음 | `engineWiring{}, err` | apply hook 미바인딩 |
| B2 | `tracker.Restore` 실패 (226) | 없음 | wrap된 err | projection 복원 실패 |
| B3 | `NewPairedReadinessAdapter` 실패 (250) | 없음 | wrap된 err | readiness 구성 실패 |
| B4 | `execgw.New` 실패 (274) | 없음 | wrap된 err | 게이트웨이 구성 실패 |
| — | 정상 (297) | — | `engineWiring{...}, nil` | 정상 조립 |

**어떤 분기도 side effect를 남기지 않는다.** 실패하면 부분 조립된 것이 밖으로 나가지 않고,
이미 만들어진 `tracker`·`entry`는 참조가 끊긴다. 새 워커를 추가할 때 **고루틴을 여기서 띄우면
이 성질이 깨진다** — B3·B4에서 반환할 때 이미 돌기 시작한 워커를 멈출 경로가 없다.
⇒ 워커는 여기서 **생성만** 하고 기동은 호출자가 한다.

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `checkProjectionWired` | apply hook 바인딩 확인 | 에러 = 조립 중단(B1) | AST L204 |
| `execgw.NewEntryGate` | 진입 게이트 | 에러 없음 | AST L214 |
| `tracker.Restore` | RECONCILE projection 복원 | 에러 = 조립 중단(B2) | AST L226 |
| `productionProtectionAssemblies` | supervisor assembly 목록 | 에러 없음 | AST L243, 별도 산출물 있음 |
| `protectionreadiness.NewProductionProvider` | 보호 readiness 생산자 | 에러 없음(생성만) | AST L244 |
| `protection.NewPairedReadinessAdapter` | 읽기 전용 readiness 어댑터 | 에러 = 조립 중단(B3) | AST L249 |
| `execgw.New` | 실행 게이트웨이 | 에러 = 조립 중단(B4) | AST L253 |
| `newNotifier` / `newRetrier` | 알림·재시도 | 에러 없음 | AST L280-281 |

## State mutations and fallbacks

- `entry.SetAuthorityRefresh`(L229)가 tracker를 진입 게이트에 **뒤로 연결**한다. 순환 참조를
  콜백으로 끊는 기존 방식이며, 워커가 tracker를 필요로 하면 같은 방식을 쓴다.
- 반환하는 `engineWiring`은 값이 아니라 **살아 있는 포인터 묶음**이다. 워커를 여기에 추가하면
  호출자가 그 수명을 책임진다.

## Safety conclusion

- Safe edit boundary: 구성요소 **생성**만 추가한다. 고루틴 기동·설정 분기·순서 재배치는 이
  함수 밖.
- High-risk impact: **yes.** 이 함수가 실패하면 엔진이 뜨지 않고, 뜨지 않으면 손절도 없다
  (기억 `tossos-engine-stop-removes-stoploss`).
- **설계 귀결:** 워커는 (1) 무조건 생성, (2) 기동은 호출자, (3) tracker 의존은 콜백으로,
  (4) 실패는 새 return이 아니라 기존 4개 중 하나의 형태를 따른다.
