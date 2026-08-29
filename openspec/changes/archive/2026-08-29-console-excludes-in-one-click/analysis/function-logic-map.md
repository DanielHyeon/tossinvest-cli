# Function Logic Map: console-excludes-in-one-click

기존 함수 **내부 로직**을 고치는 전 대상. 신규 추가만 하는 심볼은 마지막 절에 목록으로만 둔다.

## §0. 구현 전 코드 확인에서 나온 사실

### §0.1 제외는 살아 있는 손절을 벗기지 못한다 (§0.4 확인)

이 change에서 답해야 했던 단 하나의 안전 질문. `ReconcileDriver.judgeHoldings`
(`internal/app/engine/adoption.go`):

| 순서 | 코드 | 이 change에 대한 의미 |
|---|---|---|
| 108 | `if p.ExitEligible() { … continue }` | **이미 관리 중인 포지션은 여기서 반환된다** |
| 116 | `if d.blocked(market, symbol) { continue }` | RECONCILE 중 |
| 119 | `if !fresh { continue }` | |
| 127 | `if d.opts.Adoption.Excludes(symbol) { … }` | 제외 판정은 **여기가 처음** |

제외 판정은 exit 자격 판정보다 **뒤**에 있다. 따라서 `exclude_symbols`에 심볼을 더하는
행위는 이미 exit 관리 중인 포지션의 손절·익절에 도달하지 못한다. §0.4(손절 즉시성)
위반 없음. 같은 사실이 `console-adoption-controls` 스펙에 이미 다른 말로 있다 —
"exclude 추가는 이미 편입된 포지션에 무효".

**귀결(D5)**: 관리 중인 행에 제외 컨트롤을 두면 아무 효과 없는 버튼이 된다 → 두지 않는다.

### §0.2 스펙은 이미 행별 버튼을 1차 경로로 정해 뒀다

`console-adoption-controls`의 "편입 설정 화면" 요구사항:
> 목록 직접 기입은 고급 접힘 안에만 둔다(SHALL — **1차 경로는 포지션 화면의 행별 버튼이다**)

그런데 그 1차 경로는 `include_symbols`에만 존재한다. 이 change는 새 정책을 만드는 것이
아니라 **이미 성문화된 정책이 한쪽 목록에만 적용돼 있던 것**을 마저 적용한다.
스펙 문장의 "1차 경로는 …행별 버튼이다"에 "두 목록 모두에 해당한다"를 명시한다.

### §0.3 제외 경로는 손절폭 검증을 부르지 않는다

`config.Adoption.validate()` (engine.go:129-139):
```go
if !a.Enabled && len(a.IncludeSymbols) == 0 && a.DefaultStopPct == 0 { return "" }
if err := exitpolicy.ValidateStopPct(a.DefaultStopPct); err != nil { return err.Error() }
```
조건에 있는 것은 `IncludeSymbols`뿐이고 `ExcludeSymbols`는 없다. 그러므로
`handleSettingsInclude`가 기본값 5%를 채워 넣는 이유(검증이 요구한다)가 제외에는
**존재하지 않는다**. 제외 경로는 `DefaultStopPct`를 건드리지 않는다 → 1.9 전용 테스트.

---

## 1. `console.(*Console).handleSettingsInclude` — settings.go:155

편입 지정의 공지 한 갈래만 고친다. **쓰기 경로는 손대지 않는다.**

### Before

| # | 조건 | 결과 |
|---|---|---|
| 1 | `Settings == nil` | 501 refuse |
| 2 | `symbol == ""` | 400 refuse |
| 3 | `Load()` 오류 | redirect "지정 안 됨 — 설정 파일을 읽을 수 없다" |
| 4 | `remove=1` | include에서 제거 → Save → "지정 해제됨" |
| 5 | `!Included(symbol)` | include에 추가 + 정렬 |
| 6 | `DefaultStopPct` 무효 | 5% 채움 + `usedDefault` 문구 |
| 7 | Save 오류 | redirect "지정 안 됨" |
| 8 | 성공 | **"편입 예약됨 — 상시 규칙이다…"** |

### After — 8번 갈래만 분기

| # | 조건 | 결과 |
|---|---|---|
| 8a | 성공 ∧ `¬current.Excludes(symbol)` | 오늘과 **바이트 단위로 같은** 문구 |
| 8b | 성공 ∧ `current.Excludes(symbol)` | 지정은 기록하되 공지가 "제외 목록에 있어 편입되지 않는다 — 제외를 먼저 해제한다"를 말한다 |

근거: 엔진은 제외를 편입보다 우선한다(adoption.go:124-130). 8b 상태에서 오늘의 문구
"편입 예약됨 — 상시 규칙이다"는 **거짓**이다. 화면이 컨트롤을 감추는 것(D6)은 UI이고
강제가 아니므로 직접 POST가 이 갈래에 도달할 수 있다.

1~7은 무변경. 특히 6(5% 채움)은 그대로다 — 제외된 심볼이어도 include 목록이 비어 있지
않게 되므로 검증은 여전히 요구된다.

## 2. `console.(*Console).handlePositions` — portfolio_pages.go:43

### Before

```go
if c.opts.Settings != nil {
    if block, _, err := c.opts.Settings.Load(); err == nil {
        page.CanDesignate = true
        for i := range page.Snap.Rows {
            page.Snap.Rows[i].Designated = block.Included(page.Snap.Rows[i].Symbol)
        }
    }
}
```

| # | 조건 | 결과 |
|---|---|---|
| 1 | seam 없음 | `CanDesignate=false`, 스탬프 없음 |
| 2 | `Load()` 오류 | `CanDesignate=false`, 스탬프 없음 |
| 3 | 정상 | `CanDesignate=true`, 행마다 `Designated` |

### After

3번에서 같은 루프가 `Excluded`를 **하나 더** 찍는다. 1·2는 무변경 —
seam이 없거나 읽히지 않으면 `Excluded`도 찍히지 않고, `CanDesignate=false`이므로
어느 컨트롤도 렌더되지 않는다(오늘과 동일).

`Load()`는 **여전히 1회**다. 두 번 읽으면 두 목록이 서로 다른 스냅샷에서 올 수 있다.

## 3. `console.positionRow.Label` — portfolio.go:291

### Before

| # | 조건 | 반환 |
|---|---|---|
| 1 | `Unknown()` | "관리 여부 불명" |
| 2 | `!Managed() ∧ Designated` | "관리 편입" |
| 3 | `!Managed()` | "관리 외(미편입)" |
| 4 | `HasExit ∧ Exit.Completed` | "관리 종료" |
| 5 | `HasExit` | "엔진 관리" |
| 6 | 그 외 | "엔진 관리(대기)" |

### After

| # | 조건 | 반환 | 변경 |
|---|---|---|---|
| 1 | `Unknown()` | "관리 여부 불명" | — (여전히 최우선: 원장을 못 읽었으면 아무것도 단정하지 않는다) |
| **2'** | `!Managed() ∧ Excluded` | **"관리 제외"** | **신규** |
| 3' | `!Managed() ∧ Designated` | "관리 편입" | 순서만 한 칸 밀림 |
| 4' | `!Managed()` | "관리 외(미편입)" | — |
| 5'-7' | (관리 중 3갈래) | 무변경 | — |

**2'가 3'보다 앞인 것이 계약이다** — 엔진이 제외를 편입보다 우선하므로 화면의 우선순위가
어긋나면 라벨이 엔진의 행동을 잘못 예고한다. 동시 등재 시 "관리 제외"가 이긴다.

`Excluded=false`인 모든 기존 행의 반환값은 오늘과 **동일**하다(zero value 안전).

## 4. `console.(*Console).routes` — console.go:497

`mux.HandleFunc("/settings/exclude", c.session0(c.mutating(c.handleSettingsExclude)))` 1줄
추가. 기존 등록 순서·wrapper 조합 무변경. `session0`+`mutating` 조합은
`/settings/include`와 **같아야 한다** — 정적 검사가 그것을 읽는다.

## 5. 정적 가드 3종 — static_test.go

스펙이 "스펙 문장과 정적 검사 목록이 같은 커밋에서 함께 움직여야 한다"를 요구하므로
세 곳 **전부**를 갱신한다. 하나라도 빠지면:

| 위치 | Before | After | 빠뜨렸을 때의 증상 |
|---|---|---|---|
| 라우트 수 하한 + 열거 주석 (286-302) | `< 19`, "2 its two edits" | `< 20`, "3 its three edits" | 카나리아가 실제보다 낮아져 파서가 멈춰도 통과 |
| `stateChanging` map (336-359) | 9개 | 10개 | 새 라우트가 "CSRF 게이트 뒤에 있는 읽기 라우트"로 판정돼 **실패** |
| `consoleStateChanging` (606-609) | 9개 | 10개 | 새 라우트가 미승인 상태변경으로 판정돼 **실패** |

즉 뒤의 둘은 빠뜨리면 빌드가 붉어지고, 첫째만 조용히 썩는다 → 첫째를 잊지 않도록
이 표를 남긴다.

## 6. 템플릿 — templates_portfolio.go (positions)

| 항목 | Before | After |
|---|---|---|
| 헤더 | `<th>관리 편입</th>` | `<th>관리 편입</th><th>제외</th>` |
| detail row `colspan` | Multi 9 / 8 | Multi **10** / **9** |
| 편입 셀 | `CanDesignate ∧ InBroker ∧ ¬Managed ∧ ¬Unknown`이면 체크박스 | 같은 조건 **∧ ¬Excluded**. `Excluded`면 "제외를 먼저 해제" 안내 |
| 제외 셀 | 없음 | 편입 셀과 **정확히 같은** 행 조건에서 체크박스 |

헤더 `<th>관리 편입</th>`는 `portfolio_label_test.go`가 못박은 문자열이므로 손대지 않고
**열을 하나 더 둔다**. 두 컨트롤의 행 조건이 하나면 서로 어긋날 수 없다.

## 7. 신규 추가 (기존 로직 무변경)

- `console.(*Console).handleSettingsExclude` — settings.go
- `console.positionRow.Excluded` bool — portfolio.go (zero value = 오늘)
- 라우트 `/settings/exclude`

`AdoptionSettings` seam은 **Load/Save 2개 그대로**다 — 새 능력을 얻지 않는다.

---

## Branch Test Map

| 분기 | 테스트 |
|---|---|
| 1-8a 오늘의 문구 유지 | `TestDesignatingASymbolFromThePositionsScreen`(기존) |
| 1-8b 제외된 심볼의 편입 요청 | `TestDesignatingAnExcludedSymbolSaysTheExclusionWins` |
| 2-3 `Excluded` 스탬프 | `TestAnExcludedRowIsLabelledAndOffersRelease` |
| 2-1·2-2 seam 없음/오류 | `TestNoSeamRendersNeitherControl` |
| 3-2' "관리 제외" | `TestAnExcludedRowIsLabelledAndOffersRelease` |
| 3-2' > 3-3' 우선순위 | `TestExclusionBeatsDesignationInTheLabel` |
| 3-1 Unknown 우선 | `TestAnUnreadableJournalStaysUnknownEvenWhenExcluded` |
| 4 라우트 등록·게이트 | `TestSettingsPostsWithoutCSRFWriteNothing`(확장), 정적 검사 |
| 5 가드 3종 | 정적 검사 자체 |
| 6 컨트롤 렌더 조건 | `TestTheExcludeControlOnlyRendersOnUnmanagedKnownRows` |
| 신규 쓰기 경로 | `TestExcludingASymbolFromThePositionsScreen` (추가·멱등·타 필드 보존) |
| 신규 해제 경로 | `TestReleasingAnExclusion` |
| D4 손절폭 미침범 | `TestExclusionNeverInventsAStopFraction` |
| D3 상호 배제 | `TestExcludingADesignatedSymbolDropsTheDesignation` |
| 마찰 금지 | `TestTheExcludeControlAsksForNoTyping` |
