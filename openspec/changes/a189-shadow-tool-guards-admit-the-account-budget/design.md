# Design

## Context

### 가드가 이름을 세는 방법 (편집 전 AST, `analysis/pre-edit-ast/`, base `7d1e3163a`)

- 두 도구의 `nonTestSources` 는 `<도구 디렉터리>/*.go` 중 `_test.go` 를 뺀 파일을 `parser.ParseFile(…, 0)` 로 읽음. 빌드 태그는 보지 않음.
- a128 `TestNoMutatingOrAccountSurfaceIsNamed`(static_test.go:78~126, 분기 8):
  - B1 range(99) 파일마다 `ast.Inspect`.
  - B2 if(102) `SelectorExpr` 아니면 계속.
  - B3 if(105) **모든** 선택자의 `Sel` 이름이 `forbidden` 에 있으면 실패 — 패키지와 무관함.
  - B4 if(108) X 식별자가 `official` 이면 B5 if(110) `allowedOfficial` 밖이면 실패.
  - B6 if(114) X 가 `tossconfig` 이면 B7 if(116) `allowedConfig` 밖이면 실패.
  - B8 if(123) `official`·`tossconfig` 중 하나라도 못 봤으면 Fatal.
- a128 `TestImportsStayInsideTheAllowlist`(162~181, 분기 3): B1·B2 range(173·174), B3 if(176) 경로에 `.` 이 있고 `allowedModule` 밖이면 실패.
- a140 `TestNoMutatingAccountOrWriteSurfaceIsNamed`(70~116, 분기 8):
  - B3 if(97) `forbidden` 은 a128 과 같은 전역 검사.
  - B4 if(100)·B5 if(101) X 식별자가 `allowed` 맵의 키이면 B6 if(103) 그 패키지 목록 밖이면 실패.
  - B7 range(111)·B8 if(112) `allowed` 의 키마다 `saw` 가 없으면 실패.
  - 그래서 `allowed` 에 패키지를 더하면 「본 것 없음」 검사도 자동으로 그 패키지에 걸림.
- a140 `TestImportsStayInsideTheAllowlist`(118~136, 분기 4): B3 if(128) 표준 라이브러리 건너뜀, B4 if(131) 모듈 접두사 밖이거나 접미사가 `allowed` 밖이면 실패.
- 두 도구 모두 패키지는 **선택자의 X 식별자 철자**로 가림. import 경로가 아님.
- A112 `a112MBUSReferenceAllowed`(a112_mbus_static_test.go:489~498, 분기 2, return 3):
  - B1 if(491) `tools/a112-mb-us-source/` 접두사 → true.
  - B2 if(494) `internal/official/a112_mbus_` 접두사이면서 `_test.go` 접미사 → true.
  - 마지막 return(497) 정의 파일 세 경로와의 정확한 비교.
  - 호출자는 `a112MBUSFindForbiddenRefs` 하나(CodeGraph 1.6.0, `analysis/code-context/codegraph-1.6.0.txt`). 그 함수가 true 이면 파일을 파싱하지도 않고 위반 0 을 돌려줌.

### 지금 비시험 소스가 쓰는 이름 (D2 열거, `analysis/enumeration/selectors-and-imports.txt`)

열거는 `nonTestSources` 와 같은 파일 집합(a128 12개, a140 13개)을 같은 파서 모드로 읽음.

| 도구 | import (새 것) | `official.*` | `accountbudget.*` | `accountbudgetruntime.*` |
|---|---|---|---|---|
| a128 | `internal/accountbudget`, `internal/accountbudgetruntime`(둘 다 별칭 없음) | ErrRateLimited, LoadCredentials, MarketCalendarResponse, **Option**(main.go:80), StrictMinutePage, StrictTopOfBook | Auxiliary·WithPriority(main.go:164) | Open(main.go:154), ParseLaunchConfig(main.go:149) |
| a140 | 같은 둘(별칭 없음) | ErrRateLimited, LoadCredentials, MarketCalendarResponse, **Option**(main.go:74), StrictLastPrice, StrictMinutePage, StrictTopOfBook | Auxiliary·WithPriority(main.go:156) | Open(main.go:146), ParseLaunchConfig(main.go:141) |

- 굵은 `Option` 만 현 허용 목록 밖임. 나머지 `official.*` 는 이미 허용됨.
- 열거에 따르면 현 허용 목록 중 a128 `New`·`Client`·`Credentials`, a140 `New`·`Client` 는 지금 쓰이지 않음. 이 change 에서는 정리하지 않음(Non-goals).
- 예산 쪽 다른 호출(`factory.NewClient`, `factory.Close`, `budgetConfig.CleanupError`)은 X 가 지역 변수라 패키지 추적 대상이 아님. 이름이 `forbidden` 에 없음을 base 실패 원문으로 확인함: B3 실패 줄이 0개.

## Goals / Non-Goals

- Goal: 실패하는 가드 5개를 시험 파일 편집만으로 통과시킴. 넓히는 폭은 위 표의 이름만큼만이고, 새로 들이는 두 패키지는 선택자 이름까지 고정함.
- Non-goal: 비시험 코드, 금지 목록 약화, 미사용 허용 이름 정리, 별칭 우회 막기, a128 요구사항 신설.

## Decisions

### D1. 고치는 것은 시험 파일의 허용 목록뿐

a166 의 배선(`78d777cfe`)은 의도된 것임. 되돌리면 두 도구가 공유 예산 밖에서 요청을 보내므로 a166 요구사항(계좌 공유 요청 admission)과 어긋남.

### D2. 새 패키지는 기존 패키지별 허용 맵 모양으로 고정

- a140: 기존 `allowed` 맵에 두 키를 더함. B7·B8 이 「본 것 없음」을 같이 잼.
- a128: `TestNoMutatingOrAccountSurfaceIsNamed` 안에 같은 파일 `TestWTSSourceIsTheThreeReads` 의 `allowedPackage`+`saw` 모양을 그대로 둠(새 헬퍼 없음). 같은 `ast.Inspect` 콜백에 갈래 하나를 더하고, 루프 뒤에 두 패키지의 `saw` 검사를 더함.
- import 허용 목록은 두 경로만 더함.

### D3. 금지 목록과 기존 검사는 그대로

`forbidden` 맵, `allowedConfig`, `journal`·`tossconfig` 항목, `marketSource`·`wtsSource` 메서드 고정 시험은 글자 하나 바꾸지 않음.

### D4. A112 가드 — 선택지 비교 (Manager 결정)

| | A. 경로 한 줄 허용 | B. 참조를 없앰(시험 이동) | A-wide. 모든 `_test.go` 면제 |
|---|---|---|---|
| 편집 | `a112MBUSReferenceAllowed` 마지막 return 의 정확한 경로 비교에 `internal/accountbudgetruntime/v2_token_test.go` 를 더함(1줄) | `TestA166RuntimeV2A112CloneRecordsPending`(v2_token_test.go:392~470, 분기 14·호출 76)을 지우고, `internal/official/a112_mbus_*_test.go` 에 다시 씀 | `strings.HasSuffix(path, "_test.go")` 면 true |
| 새로 통과하는 입력 | 그 파일 하나. 그 안에서는 M-B0 이름 다섯(`A112MBUSResult`·`HoldError`·`Candle`·`Orderbook`·`Calendar`) 전부 허용 — 파일 단위 면제라 이름 단위로는 못 좁힘 | 없음(가드 무편집) | 저장소의 모든 시험 파일 |
| 비용 | FLM 번들 1개 | 옮길 자리에서 `accountbudgetruntime` 의 비공개 시험 도우미 `setupGroup`·`manifest`·`a166RuntimeStore` 를 쓸 수 없음. 패키지 `official` 안 시험은 `accountbudgetruntime` 을 import 하면 순환이 되므로 `official_test` 외부 시험으로 써야 하고 매니페스트 준비를 새로 짜야 함. 삭제한 base 함수에 `revision: base` 번들 필요. 활성 change a166 의 파일을 건드림 | 1줄 |
| 위험 | 판정 범위가 파일 하나로 닫힘 | 옮기는 중 시험이 지키던 것(복제 전송의 대기 관측 계수)을 잃을 수 있음 | 시험 파일에서 M-B0 를 부르는 새 경로를 아무도 못 봄. 가드 목적(제품 caller 0)에는 맞지만 범위가 열림 — 추천하지 않음 |
| 스펙 | 정본 문장 「M-B0 exported-symbol references are allowed only in their definitions, M-B0 tests and the exact tool」(breakout-retest-strategy-lane:91)과의 관계가 결정 대상(아래) | 문장 그대로 | 문장을 넓혀야 함 |

**추천: A.**
- 가장 작은 변경임.
- 그 시험은 M-B0 복제 전송이 공유 예산에 대기 관측으로 계수되는지를 증명함. 그대로 두는 것이 a166 증거를 지킴.
- 이미 같은 계열 시험이 M-B0 명명 규칙 안에 있음: `internal/official/a112_mbus_a166_test.go` `TestA166A112CloneSharesBudget`. 다만 그것은 `accountbudget` 만 쓰고 런타임 팩토리·관측 저장소는 쓰지 않아 대체가 아님.

A 를 고를 때 스펙 처리 둘 중 하나를 Manager 가 정함:
- A-1(해석): 그 시험은 실질적으로 「M-B0 test」이고, 가드의 경로 접두사는 구현의 더 좁은 정의임 → 스펙 개정 없음. 문장 해석이 기록에만 남음.
- A-2(개정): `analysis/option-a-mb0-delta-draft.md`(정본 90~126행 블록을 그대로 복사하고 한 문장만 바꾼 MODIFIED 초안)를 specs/ 로 옮김. 블록이 커서 옮겨 적는 오류 위험은 생성 스크립트(정확한 문자열 대조 assert)로 줄였음.

A 의 부수 영향:
- `internal/official/a112_mbus_static_test.go` 는 M-B1 도구의 측정 소스 신원 집합에 들어 있음(`tools/a112-mb-us-source/identity.go:802` `measurementSourceDigests`).
- 그래서 앞으로의 M-B 영수증에 적히는 이 파일의 sha256 이 바뀜(현재 `a6cbe55c…bf9659`).
- 저장소에서 이 값을 상수로 고정한 곳은 없음(`grep` 0건, `.gz` 는 안 봤음). 신원 비교는 한 실행 안의 전후 대조임.
- 이 파일이 바뀌는 것은 2026-08-17(`29f050545`) 이후 처음임.
- B 는 이 파일을 건드리지 않음.

### D5. `official.Option` 허용의 영향

원문: `analysis/enumeration/official-option-impact.txt`.

- 가드는 `official.<이름>` 을 이름 하나씩 봄(B5/B6). `Option` 을 더하면 새로 통과하는 것은 선택자 `official.Option` 하나뿐임. 위치는 타입 자리(가변 인자 `...official.Option`, 변수·슬라이스 타입)와 변환식 `official.Option(f)` 둘 다임.
- `Option` 을 돌려주는 생성자 `WithBaseURL`·`WithHTTPClient`·`WithAccountSeq` 는 각자 다른 선택자 이름이라 여전히 개별 허용이 필요함 — 둘 다 허용 목록에 없음. `WithAccountSeq` 는 `forbidden` 에도 있어 어느 패키지 앞에서든 실패함. `WithAttemptObserver`(trace.go:29)는 `Option` 이 아니지만 역시 목록 밖임.
- 지금 비시험 소스에서 `official.With*`·`official.Option(` 출현은 0건임. 운영 경로 `runCommand` 는 옵션 없이 `runCommandWithOptions` 를 부르고, 옵션을 넘기는 곳은 시험 파일뿐임(가드 대상 밖).
- 변환식 경로: `official.Option(func(c *official.Client){…})` 는 이름 검사를 통과할 수 있음. 그 클로저가 얻는 것은 `*official.Client` 하나임.
  - `Client` 에는 내보낸 필드가 없음(`go doc`: 「Has unexported fields」).
  - 그래서 할 수 있는 것은 내보낸 메서드 호출뿐이고, 그 이름도 B3 `forbidden` 검사를 그대로 받음.
  - 도구는 이미 `factory.NewClient` 가 돌려준 `*official.Client` 를 들고 있음. 그러니 도달 범위가 새로 늘지 않음.
  - 생성 뒤에는 `configurationSealed` 로 표준 옵션 재적용이 막힘(client.go:30~32 주석).

### D9. 보조 등급 고정과 예산 핸들 메서드 금지 (보강, Manager 결정 2026-11-15)

- 이유: 공유 예산에서 Critical 은 엔진 보호 주문 몫임. 측정 도구가 Critical 을 달면 손절 즉시성을 깎을 수 있음(안전 불변식 4). 2.2 의 이름 고정은 패키지 한정 선택자만 봐서 `WithPriority(ctx, 1)`(리터럴 1 = Critical)과 `factory.Budget()` 핸들 경로를 못 막았음(gstack 발견 1).
- 우선순위: `types.ExprString(call.Fun) == "accountbudget.WithPriority"` 인 직접 호출마다 인자 2개·둘째 인자 `accountbudget.Auxiliary` 를 요구함. 같은 선택자의 참조 수와 직접 호출 수가 같아야 하고(함수 값으로 빼서 부르면 실패), 호출이 0 이면 실패(전칭 판정이 표본 0 으로 통과하지 않게).
- 금지 이름: `go doc -all` 로 두 패키지의 내보낸 메서드를 전수 열거함 — `accountbudget`: `Budget.{Acquire,Close,EnginePollKey,NewObservationFence,ObservationSnapshot,PollRemaining,VerifyIdentity}`, `Lease.Release`, `ObservationCollector.Snapshot`, `ObservationFence.Context`, `ObservationStore.{Begin,Finish,PollRemaining,Snapshot}`, `Transport.RoundTrip`; `accountbudgetruntime`: `Factory.{Budget,Close,NewClient}`.
  - 더한 것 7개: `Budget`, `Acquire`, `EnginePollKey`, `NewObservationFence`, `ObservationSnapshot`, `PollRemaining`, `VerifyIdentity`.
  - 뺀 것: 도구가 쓰는 `NewClient`·`Close`, 흔한 이름이라 오탐 위험이 큰 `Release`·`Begin`·`Finish`·`Snapshot`·`Context`·`RoundTrip`.
  - 0건 근거: 두 도구 비시험 소스에서 `\.(Budget|Acquire|EnginePollKey|NewObservationFence|ObservationSnapshot|PollRemaining|VerifyIdentity|Release|Begin|Finish)\b` grep 0건, 보강 뒤 두 시험 PASS.
  - 앞으로 거절하게 될 정상 사용: 다른 패키지의 같은 이름 — a140 이 import 하는 `internal/scheduler` 결정 구조의 `Budget` 필드(`decision.go:39·47`)와 `internal/strategyrouter` `QuotaAuthority.Acquire`(quota.go:205) 를 도구가 언젠가 쓰면 그것도 거절함(이름 단위 금지라 받는 쪽 타입을 안 봄).

### D10. a166 census 고정 모양 갱신을 a189 에 포함 (사용자 결정 2026-11-15)

- 이유: a189 의 목적은 main 의 `go test ./...` 를 다시 초록으로 만들어 a187 archive 를 푸는 것임. census 실패는 그 목적을 막는 같은 종류의 낡은 가드 값이었음 — a179 가 `runCommandWithOptions` 를 바꾸며 a166 의 고정 모양을 갱신하지 않음.
- 범위: `shapes.json` 의 값 하나. 고정 모양의 의미(그 함수 본문 전체를 리뷰된 모양으로 못 박음)와 다른 89개 키, 호출 자리 분류 `sites.json`, census 코드는 무편집.
- 리뷰: Manager 가 모양 diff(`analysis/census-investigation/shape-diff.txt`)를 읽고 승인 — 변화는 전부 a179 입력 프로필, 예산 배선 줄 무변화, 새 브로커 HTTP 호출 자리 없음.
- 생성: 저장소에 corpus 생성 스크립트·`-update` 플래그가 없음(`shapes.json` 을 읽는 곳은 `a166_wiring_census_test.go:35~36` 뿐). census 와 같은 추출(`format.Node`, Doc 제거)로 값을 만들고, Go `encoding/json` 문자열 인코딩으로 기존 값의 표기를 바이트 그대로 재현함을 먼저 단언한 뒤 바꿈.
- FLM: 비 Go 파일이라 gate ⑤ 대상 아님(required 수 5 그대로).
- 반증: 사본에서 그 함수 안에 예산을 거치지 않는 `http.Get` 을 넣으면 모양 검사가, 새 함수에 넣으면 호출 자리 수 검사가 여전히 실패함(`analysis/census-repair/falsify.txt`).

### D6. 고칠 함수와 FLM 번들

- 선택지 A 기준 5개. 편집 전 AST 는 `analysis/pre-edit-ast/` 의 5개 파일이고, B 비교용으로 `optionB-…` 1개를 더 둠. 편집 뒤 번들은 2단계에서 `scaffold_analysis.py` 와 `go run ./tools/logic-map` 으로 만듦.
- 같은 이름 `TestImportsStayInsideTheAllowlist` 가 두 패키지에 있음. 번들 디렉터리는 패키지 경로로 갈림.

### D7. 위험도와 검증

- 주문·손절 경로가 아니고 시험 전용임(Small). 다만 읽기 전용 프로브의 안전 가드 판정을 바꿈.
- 2단계에서 사본 변이(양성 대조군)를 돌림. 넓힌 뒤에도 다음을 각각 넣으면 해당 가드가 실패하는지 봄:
  - 목록 밖 예산 이름(예: `accountbudget.Critical`, `accountbudgetruntime.Config`)
  - 금지 이름(예: `PlaceOrder`)
  - 목록 밖 official 이름(예: `official.WithBaseURL`)
  - 목록 밖 import
  - 예산 패키지 선택자 전부 제거(「본 것 없음」)
  - A112: 허용 경로 밖 시험 파일의 M-B0 참조
- 다중 적대 리뷰는 `not-applicable`(비례 원칙: High-risk 경로·게이트 판정 직접 변경 아님). gstack `review` 1회.

### D11. 콘솔 orders seam — 문자열 고정을 함수 본문 AST 단언으로 (범위 확장 2, #1)

- 경계 실측(`analysis/boundary/boundary.txt`): `3dc0c776b^` PASS, `3dc0c776b` FAIL(console_test.go:755). 그 커밋이 `consoleBroker.resolve()` 를 `resolve(ctx)` 로 바꿈. 철자 수(`git show <c>:cmd/tossctl/console.go`): 부모 `l.shared.resolve()` 2·`(ctx)` 0, 원인 커밋과 base 는 0·2.
- **철자 수가 2 인 것이 문제임**: 같은 파일의 `(*lazyHoldings).Holdings` 도 수신자 이름이 `l` 이라 같은 철자를 씀. 그래서 철자만 고치면 `Orders` 가 공유 resolver 를 안 써도 `Holdings` 의 철자로 통과함 — 이 시험이 지키는 「orders seam 이 공유 resolver 로 읽음」을 못 봄.
- 결정: 파일 전체 `strings.Contains` 를 `(*lazyOrders).Orders` 본문 안에서 `<수신자 이름>.shared.resolve` 선택자 **참조** 수 == 1 로 바꿈(호출이 아니라 참조 — 메서드 값으로 빼도 셈). `Orders` 선언을 못 찾으면 실패. 수신자 이름은 선언에서 읽음(철자 `l` 에 묶지 않음).
- 변이로 두 판본(철자만 고친 S, AST 판본 A)을 비교해 선택을 잼(`analysis/mutation-2/`). 빌드 횟수 1·`official.New(` 금지는 그대로.

### D12. a163 렌더 해시 — `record` 값 하나 갱신 (범위 확장 2, #2)

- 경계 실측: `9c4bdf4de^` PASS, `9c4bdf4de` FAIL(a163_judgement_pin_test.go:102). 렌더 하네스(`analysis/render-hash/render.go.txt`, 시험과 같은 파싱·렌더)로 `9c4bdf4de^..7d1e3163a` 의 rev-list 전수 135 커밋 + 부모 1 을 잼: `exitloop.go` 의 서로 다른 blob 3개, `record` 해시는 둘뿐(`f3458b84…` 17 커밋, `0a776329…` 119 커밋). 부모 대비 `record` 가 바뀐 커밋은 `9c4bdf4de` 와, 9c4bdf4de 를 둘째 부모로 들여온 병합 둘(`5f681f2a5`·`5b379e08d`)뿐. 그 뒤 base 까지 `record` 불변(`render-at-commits.txt`). 나머지 9개 함수는 경계 세 지점에서 기대값과 같음.
- 그래서 새 값 `0a776329…` 는 a166 `9c4bdf4de` 의 한 편집임. 편집 요지(`analysis/render-hash/cause-9c4bdf4de-exitloop.diff`): `record` 의 청산 정리 갈래에서 `o.clearTheSymbol(ctx, …)` 에 넘기는 context 를 `accountbudget.WithPriority(ctx, accountbudget.Normal)` 로 만들고, `isProtective(proposal)`(9c4bdf4de 이전부터 있던 술어 — 손절선 이탈·래더 손절)이면 `Critical` 로 바꿈. `WithPriority` 는 `context.WithValue` 한 줄임(transport_linux.go:21·unsupported.go:44).
- AST 대조(`tools/logic-map`, `analysis/render-hash/record-ast-*.json`·`record-ast-compare.txt`): 편집 전→후 분기 18→19, return 6→6, 호출 24→27, 대입 21→23. 더해진 항목은 정확히 여섯 — 대입 `:=`·`=` 각 1, 호출 `accountbudget.WithPriority` 2·`isProtective` 1, 분기 `if` 1(우선순위 고르기). 빠진 항목 0, 같은 항목 69. `9c4bdf4de`→base 는 75 항목 전부 같음.
- 판정 불변의 근거는 이 대조뿐임: 새 `if` 는 context 의 우선순위 값만 고르고 return·`orderable`·`judgement` 대입이 없음. 청산 여부·순서를 정하는 기존 분기·return 은 전부 같은 항목으로 맞물림. 우선순위가 계좌 공유 예산의 전송 허가 시점에 주는 영향은 a166 의 계약이고 `record` 의 판정 트리 밖임.
- 갱신 자리: 가드 주석(「갱신하려면 새 change 가 이유를 적어야 함」)의 바로 그 맵 항목 위에 이유 주석을 둠. 맵은 패키지 수준 변수라 FLM 대상 함수가 아님(`check_analysis` 실측으로 확인).

### D13. dispatch seam Admit 자리 — 경계 패키지의 이름만 세기 (범위 확장 2, #3)

- 경계 실측: `5c0aeb2b8^` PASS, `5c0aeb2b8` FAIL(strategy_dispatch_handoff_guard_test.go:669, `type 블록:Admit`). a087 이 `exit_lower_limit.go:58` 에 `type lowerLimitAdmission interface{ Admit(context.Context) error }` 를 넣음. 이 `Admit` 은 경계(strategyhandoff) 와 무관한 엔진 사설 인터페이스의 메서드 이름임.
- 결정(Manager 권장 「E」): `admitSites` 가 세는 것을 둘로 좁힘.
  1. 이 파일의 import 중 경로가 `/strategyhandoff` 로 끝나는 것의 지역 이름(별칭이면 별칭, 아니면 `strategyhandoff`)을 X 로 갖는 선택자 `X.<문 이름>` — 호출이 아니라 **참조**를 셈(함수 값 빼돌리기·합성 리터럴 안 함수 값 포함).
  2. dot import(어느 경로든)가 있는 파일에서, 선택자의 `Sel` 이 아닌 맨 식별자 `<문 이름>` — 보수적으로 셈(dot import 된 패키지가 무엇이든 맨 이름은 그 패키지 것일 수 있음).
- 경로를 정확한 모듈 경로가 아니라 끝 원소로 맞추는 이유: 기존 fixture 시험 `TestAdmitCensusSeesEverySpellingItClaims` 가 `sh "example.com/strategyhandoff"` 로 별칭 철자를 재므로, 그 시험 본문을 고치지 않고 그대로 통과해야 함. 끝 원소 비교는 정확 비교보다 넓음(같은 끝 이름의 다른 패키지도 셈 → 실패 쪽으로 틀림). 생산 엔진의 import 는 정확 경로 하나뿐임.
- 좁힘이 새로 놓치는 모양은 변이로 세어 남은 위험에 적음(R4). 새 fixture 시험은 새 파일(기존 파일 속 새 함수는 logic-map 범위를 늘림).

## Risks / Trade-offs

- R1 별칭 우회: 두 도구의 가드는 X 식별자 철자로 패키지를 가림. 그래서 `ab "…/accountbudget"` 처럼 별칭으로 import 하면 import 검사는 통과하고 선택자 이름 고정은 비켜 감. `official`·`tossconfig` 도 지금 같은 한계가 있음(기존). (gstack 리뷰로 정정, 2026-11-15) 이 change 는 이 구멍을 새 두 패키지로 **넓힘**: 편집 전에는 두 예산 경로가 import 단계에서 막혔고, 이제 import 는 통과하므로 별칭 쪽 이름은 안 보임. 「본 것 없음」 검사는 **모든** 사용을 별칭으로 바꾼 경우만 잡음(변이 f-alias-runtime) — 정식 이름 사용을 남긴 채 같은 경로를 별칭으로 한 번 더 import 하면(Go 는 허용) 통과함. 막으려면 import 의 지역 이름을 함께 고정해야 함. Manager 결정(2026-11-15)으로 범위 밖.
- R2 A 의 파일 단위 면제: v2_token_test.go 에 M-B0 이름이 더 들어와도 가드가 안 봄. 그 파일은 시험 파일이라 제품 caller 0 이라는 가드 목적은 유지됨.
- R3 M-B1 신원 집합의 파일 변경(D4 부수 영향).

## Migration Plan

없음 — 시험 파일만 바뀜. 되돌리기는 커밋 revert.

## Open Questions

없음. Manager 결정(2026-11-15): D4 = A, 스펙 = A-2, R1 = 범위 밖.
