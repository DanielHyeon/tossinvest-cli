# Function Logic Map: `openHTTPAPIResources`

- Source: `cmd/tossctl/httpapi.go` (592-661)
- AST evidence: `ast.json` — AST branches 10 · returns 6 · defers 0
- Risk scan: `risk-pattern-report.md`
- 편집 대상: **B9 의 `else`** 하나와 그것을 쓰기 위한 인자 하나(`errOut io.Writer`).
  gstack 리뷰 A3 — 무경고 강등 금지.

## 이 함수가 하는 일

조회 데몬이 읽을 저장소(원장·성과·최적화)를 열고 `httpAPIReader` 를 조립한다.
**두 부류의 실패가 섞여 있는 것이 이 함수의 성격**이고, 그 구분이 이 편집의 문맥이다.

| 부류 | 예 | 결과 |
|---|---|---|
| 없어도 데몬이 뜬다 | 원장(B2)·성과(B3)·엔진 디렉터리(B9) | 해당 화면만 비고 진행 |
| 없으면 못 뜬다 | journal 경로(B1)·최적화 저장소(B4·B5)·`reader.validate()`(B10) | 거절 |

편집 전에는 「없어도 뜬다」 셋 중 **엔진 디렉터리만 아무 말도 하지 않았다.** 원장·성과는
실패를 `journalErr`·`performanceErr` 로 reader 에 넘겨 화면이 사유를 그리지만, 엔진
디렉터리 실패는 `engineMarker` 와 `managementRuntime` 을 **빈 채로** 두고 지나갔다.
운영자가 보는 것은 「엔진 상태를 못 읽는다」이고, 그것은 「엔진이 죽었다」와 화면에서
구별되지 않는다. a108 이 httpapi 쪽에서 내린 판정과 같은 판정을 여기에 적용한다:
**삼키지 않는다는 요구는 fatal 이 아니라 경고 문구가 진다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` (`--config-dir`) | nil 허용 | `consoleJournalPath` | 해석 실패는 **거절** (B1) |
| 원장 파일 | 없어도 된다 | `journal.OpenReadOnly` | `journalErr` 로 reader 에 전달 (B2) |
| 성과 저장소 | 없어도 된다 | `openConsolePerformanceCapabilities` | `performanceErr` 로 전달 (B3) |
| 최적화 registry·DB | 필요하다 | `optimization.CoreRegistry/Open` | 거절 + 이미 연 것 닫기 (B4·B5) |
| adoption 설정 seam | 없어도 된다 | `newAdoptionSettingsSeam` | nil 이면 desired 화면만 빈다 (B8) |
| **엔진 디렉터리** | 없어도 된다 | `engineJournalDir` | **경고 + 강등** (B9 와 그 else) |
| reader 조립 결과 | 유효해야 한다 | `httpAPIReader.validate` | 거절 + 닫기 (B10) |

**관통 불변식:** 이 함수가 거절하는 이유는 **자기 저장소**뿐이다. 엔진 쪽 산출물
(마커·정책 런타임 descriptor)은 하나도 필수가 아니다 — 엔진이 없어도 조회 데몬은 뜬다.
그 결정은 그대로 두고, 이 편집은 **그 강등이 말을 하게** 만든다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal 경로 해석 실패 | 없음 | 원인 그대로 | 미고정 |
| B2 | 원장을 열었다 | `resources.journal` | 없음 | `TestADeadDescriptorDoesNotStopTheDaemon` (원장 없는 tmpdir 로 뜬다) |
| B3 | 성과 저장소를 열었다 | `resources.performance` | 없음 | 간접 |
| B4 | 최적화 registry 실패 | `resources.Close()` | 원인 그대로 | 미고정 |
| B5 | 최적화 DB 열기 실패 | `resources.Close()` | 원인 그대로 | 미고정 |
| B6 | (accountRef 클로저 안의 오류 전달) | 없음 | 호출자에게 | 간접 |
| B7 | journalReader 가 있다 → performanceSource | reader 상태 | 없음 | 간접 |
| B8 | adoption seam 이 있다 | reader 상태 | 없음 | 간접 |
| B9 | **엔진 디렉터리를 정했다 / 못 정했다** | 정했으면 marker·management 대입, **못 정했으면 stderr 경고** | 없음 (양쪽 다 진행) | 미고정 — 아래 사유 |
| B10 | `reader.validate()` 실패 | `resources.Close()` | 원인 그대로 | 미고정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleJournalPath` | 프로파일의 원장 경로 | 실패는 거절 | AST |
| `journal.OpenReadOnly` | 조회 전용 핸들 | 실패는 reader 로 전달 | AST |
| `optimization.Open` | 최적화 저장소 | 실패는 거절 + 이미 연 것 닫기 | AST |
| `engineJournalDir` | 엔진 마커·정책 런타임 descriptor 위치 | **실패는 경고 + 강등** (이 편집) | AST |
| `httpAPIReader.validate` | 조립 검증 | 실패는 거절 | AST |

## State mutations and fallbacks

- 열린 자원의 소유권은 `*httpAPIResources` 에 있고, 거절 경로는 전부 `resources.Close()`
  를 먼저 부른다. 이 편집은 그 경로를 하나도 늘리거나 줄이지 않는다.
- 새로 생긴 side effect 는 **stderr 한 줄**뿐이다. 반환값·reader 상태·거절 조건은 불변.
- 같은 원인(`engineJournalDir` 실패)이 `strategyRuntimeReaderFor` 에서도 경고를 찍는다.
  두 줄이 뜨는 것은 중복이 아니라 **잃은 화면이 둘**이기 때문이다: 하나는 전략 화면,
  하나는 엔진 상태·관리 런타임 화면이다.

## Safety conclusion

- Safe edit boundary: B9 의 `else` 와 인자 하나. 판정·순서·거절 조건 불변.
- High-risk impact: **no**. 이 함수는 조회 전용 데몬의 조립 경로이고 주문·손절·사이징·
  Guardian 어디에도 닿지 않는다. 변경 방향은 「덜 조용해진다」 한 방향이다.
