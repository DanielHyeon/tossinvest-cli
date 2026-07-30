# Review: console-excludes-in-one-click

날짜: 2026-07-30 · 위험 등급: **High-risk**

이 change가 쓰는 것은 config 블록 하나뿐이고 계좌·journal·브로커에 닿지 않는다.
그럼에도 High-risk로 다루는 이유는 그 블록의 소비자가 `ReconcileDriver`이고,
제외 여부가 보유에 **합성 손절이 생기는지**를 정하기 때문이다. 화면의 클릭 하나가
"이 보유는 보호하지 않는다"를 뜻할 수 있으면 그것은 손절 경로다.

## Pre-Edit Gate

```text
change id / task id:  console-excludes-in-one-click / 1.1-1.20
대상 심볼 (기존 함수 내부 수정):
  console.(*Console).handleSettingsInclude (settings.go:155)        — 공지 1갈래 분기
  console.(*Console).handlePositions       (portfolio_pages.go:43)  — 같은 Load에서 스탬프 1개 추가
  console.positionRow.Label                (portfolio.go:291)       — 세 번째 상태
  console.(*Console).routes                (console.go:497)         — 라우트 1줄
  console 정적 가드 3종                     (static_test.go:300·336·606)
  positions 템플릿                          (templates_portfolio.go:100-129)
가산 (신규):
  console.(*Console).handleSettingsExclude, console.positionRow.Excluded,
  라우트 /settings/exclude
기존 동작 파악 근거:
  analysis/function-logic-map.md (전 분기 표 + Branch Test Map)
  읽은 파일: internal/console/{settings,portfolio,portfolio_pages,templates_portfolio,
             console}.go, internal/config/{engine,adoption_io}.go,
             internal/app/engine/adoption.go, cmd/tossctl/adoptionsettings.go,
             static_test.go·settings_static_test.go·portfolio_label_test.go·settings_test.go
  소비자 전수: rg Excludes\(|Included\(|ExcludeSymbols (전 호출부 12곳)
upstream 상속 테스트 영향: no — internal/console은 TossOS 전용 패키지다
실패 테스트 선행 작성: yes (1.4-1.19 전부 RED 선행)
안전 불변식 §0 위반 여부 검토: 통과 — 조항 4·6·7은 아래 표에서 명시적으로 다룬다
```

## §0 대조

| 조항 | 이 change |
|---|---|
| 1 승인 없는 LIVE side effect 금지 | 무관 — 주문 경로 없음. 쓰는 것은 `engine.adoption` 블록뿐이고 계좌·브로커 무접촉 |
| 2 mutating 자동 실행 금지 | 콘솔 클릭은 사람의 행위다. 에이전트가 실행하지 않는다 |
| 3 토글 OFF = upstream 동작 | `exclude_symbols`가 비어 있으면 `Excluded=false`이고 `Label()`·템플릿·핸들러 전 갈래가 오늘과 동일. zero value 안전 |
| 4 손절·비상 청산 즉시성 | **확인 대상이었고 통과한다.** `judgeHoldings`가 `ExitEligible()`에서 제외 판정보다 먼저 반환하므로(FLM §0.1) 제외는 살아 있는 손절에 도달하지 못한다. 그 무효성을 화면에서도 지킨다 — 관리 중인 행에는 컨트롤을 두지 않는다 |
| 5 High-risk 경로 | 해당(합성 손절의 생성 여부) → full TDD + FLM + 적대적 리뷰 + gate |
| 6 손절·사이징은 보수 방향만 | 제외 추가 = 엔진이 덜 한다(보수) → 자동. 제외 해제 = 엔진이 더 한다(확장) → 자동화하지 않는다. 편입 지정과의 상호 배제도 같은 비대칭을 따른다(design D3) |
| 7 운영 토글 flip은 사람 | **이 change의 요지다.** 사람이 누르는 버튼은 사람이 한 것이다. §0.7이 요구하는 것은 사람의 결정이지 손으로 하는 파일 편집이 아니다. 루프백 리스너 + 세션 + CSRF + 확인 대화상자가 그 결정을 표시하고, `recordAdoptionSave`가 시점을 기록한다 |
| 8 시크릿·개인정보 미저장 | 기록은 심볼 목록뿐 |
| 9 주문은 공식 Open API만 | 무관 |
| 10 실계좌 자동 테스트 금지 | 전 테스트가 `fakeSettings` seam + fake broker |

## 사용자 결정과의 대조

| 결정 | 이 change |
|---|---|
| 타이핑 확인·추가 승인 마찰 금지 (2026-07-27) | 지킨다. 기존 편입 컨트롤과 **같은** `confirm()` 1회뿐이고 새 마찰은 없다. 전용 테스트로 못박는다 |
| 미체크=미편입 / 체크=편입 (2026-07-27) | 건드리지 않는다. 3-상태 라디오 재설계를 검토했다가 **버렸다** — 사용자가 정한 컨트롤을 요청받지 않은 방향으로 재설계하는 것이기 때문이다. 제외는 별도 열의 별도 컨트롤이다 |
| 한 번 클릭으로 설정 (2026-07-30) | 이 change 자체 |

## 적대적 Eng 리뷰

날짜 2026-07-30 · 시점 구현 후 · 방식: "이 클릭 하나가 운영자에게 무엇을 믿게 만드는가,
그리고 그 믿음이 어디서 깨지는가"를 경로별로 추적.

### A1. 공지가 현재형으로 거짓을 말했다 — **P1**

초안의 공지는 "**엔진은 이 심볼을 편입하지 않는다**"였다. 엔진이 가동 중이면 그것은
거짓이다 — 엔진은 기동 시점 스냅샷으로 돌고, **다음 대사 주기에 바로 그 심볼을 편입할
수 있다**. 그리고 한 번 편입되면 §0.1에 따라 그 포지션에 대해 제외는 영영 무효다.
즉 이 문장을 믿고 측정을 시작하면, 측정 중에 엔진이 그 종목을 손에 쥔 채로 시작한다.

`effectNotice`가 뒤에 "재시작해야 반영된다"를 붙이고 있었으므로 한 공지 안에서 두
문장이 서로를 부정하고 있었다.

→ 현재형 보장을 버리고 "편입 제외 **기록됨**"으로 바꿨다. 언제 효력이 생기는지는
`effectNotice` 한 곳만 말한다.
테스트 `TestTheExclusionAnswerDefersToTheEngineRestart`.

### A2. 없는 컨트롤을 가리키는 안내 — **P2**

제외된 행의 편입 칸에 "편입하려면 오른쪽 제외를 먼저 해제한다"를 넣었는데, 그 안내의
조건이 `Excluded ∧ ¬Managed ∧ ¬Unknown`이라 **컨트롤의 조건보다 넓었다**. 브로커 보유에
없는 원장 전용 행(`BrokerMissing`)은 오른쪽 칸이 `—`인데 안내만 떴다.

→ 안내의 조건을 컨트롤과 **글자 그대로 같게** 맞췄다.
테스트 `TestTheReleaseHintOnlyAppearsWhereTheControlIs`가 "안내 수 ≤ 컨트롤 수"라는
구조 불변식으로 못박는다 — 조건이 다시 갈라지면 문자열이 아니라 산술이 깨진다.

### A3. 정적 가드가 `/settings/exclude`를 아예 못 봤다 — **P2** (선행 결함)

`routeFindings`의 `actVerbs`에 `"include"`는 있는데 `"exclude"`가 없었다. 두 문자열은
부분문자열 관계가 **아니므로**(`exclude`에 `include`가 들어 있지 않다), 이 change 이전에도
누군가 `/settings/exclude`를 `consoleStateChanging`에 올리지 않고 등록했다면 "논증 없이
행위하는 라우트" 검사를 **조용히 통과**했을 것이다. 그 가드의 주석이 경계하던 바로 그
경우다("a future unlisted /settings/anything cannot sail past this guard").

→ `actVerbs`에 `"exclude"`를 더했다. 이 change의 라우트는 목록에 올라 있으므로 영향은
없고, 막는 것은 다음번이다.

### 닫지 않고 남긴 것

- **읽기-수정-쓰기의 lost update** — `Load`가 `config.Service`의 flock **밖**에서 일어나므로,
  두 탭이 각각 Load한 뒤 차례로 Save하면 먼저 쓴 심볼이 사라진다. 이것은 이 change가
  만든 것이 아니라 seam(Load/Save)의 성질이고 `/settings/save`·`/settings/include`가
  이미 같다. 고치려면 버전 토큰이나 CAS가 필요한데 `settings_static_test.go`가 seam을
  2메서드로 못박고 있어 그 확장은 별도 논증을 요구한다. 실사용 위험은 낮다 — 루프백
  단일 운영자이고 한 클릭이 전체 왕복 한 번이다. **다만 이 change가 클릭을 싸게 만든
  만큼 연타 가능성은 올라간다**는 것을 기록해 둔다.
- **제외가 관리 중 포지션에 무효라는 사실 자체** — 코드로 줄일 수 없다. 화면에서
  컨트롤을 감추고(D5) 공지가 명시하는 것이 할 수 있는 전부다.

## 검증 결과

| 항목 | 결과 |
|---|---|
| `go test ./...` | **3821 passed**, 0 failed (57 packages) — 직전 3806 대비 신규 15 |
| `make vet` | 통과 |
| `make validate` | 31 passed, 0 failed |
| `openspec validate --strict` | valid |
| 상속 테스트 회귀 | 0 — `internal/console` 263건 전부 통과, 라벨·편입 지정 기존 테스트 무수정 |
| zero value 안전 | `exclude_symbols`가 비면 `Excluded=false`이고 `Label()`·템플릿·핸들러 전 갈래가 변경 전과 동일 |
