# a067 · 설계

## D1 — 보호선의 신선도는 snapshot의 나이가 아니라 유지 주체의 생존이다

**결정**: `row.Exit.Snapshot.WithFreshness(asOf, holdingsTTL)`를 콘솔의 판정으로
교체한다. `WithFreshness` 자체는 건드리지 않는다 — journal의 순수 함수이고 다른
호출부(`cmd/tossctl/httpapi_reader.go`, journal 테스트)가 있다.

**근거**: 다섯 값(익절·손절·추적 회수·기준·고점)은 관측이 아니라 **상태**다.
`ExitLineSnapshot`은 마지막 판정이 확정한 보호선이고, 다음 판정이 그것을 바꾸기
전까지 그것이 현재 실효선이다. 시간이 지난다고 틀려지지 않는다.

`observed_at`을 나이 판정에 쓰는 것이 왜 틀렸는지는 proposal에 있다. 여기서
중요한 것은 그 컬럼이 **여전히 유용하다**는 점이다 — 무엇의 시각인지만 정확히
말하면 된다. 상세 popup의 `평가 시각`은 그대로 두고, 그것이 마지막 *변화* 시각임을
아는 것은 이제 코드가 아니라 문서와 후속 change의 몫이다.

## D2 — 미배선은 정지가 아니다: 세 갈래 판정

```go
type protectionLiveness struct {
    Wired   bool   // 엔진 마커 경로가 주입됐다
    Running bool   // 마커가 신선도 창(enginelock.StaleAfter) 안이다
}
```

| 상태 | 판정 | 왜 |
|---|---|---|
| `Wired && Running` | 나이 판정을 걸지 않는다 | 엔진이 이 journal을 5초마다 관측 중이다. 값이 안 바뀌었다는 것은 바뀔 이유가 없었다는 뜻이다 |
| `Wired && !Running` | `engine_not_running` | 아무도 이 선을 갱신하지 않는다. 지금 가격이 그 선을 넘어도 판정이 없다 |
| `!Wired` | **기존 `holdingsTTL` 나이 판정 그대로** | operator-console spec은 "미배선을 정지로 표시해서는 안 된다"고 못박는다. 알 수 없는 것을 정지로 단정하지 않고, 이 빌드에서는 지금과 동일하게 동작한다 |

세 번째가 회귀 0을 보장한다. 기존 테스트 하네스는 `EngineMarker`를 배선하지 않으므로
전부 지금 경로를 그대로 탄다. 프로덕션 콘솔은 `cmd/tossctl/console.go`가 항상
`enginelock.MarkerPath(engineJournalDir(root))`를 넣으므로 배선돼 있다.

**대안으로 검토하고 버린 것**: 미배선을 fail-closed로 처리(전부 `—`). 지금보다
보수적이지만 spec의 명시적 SHALL NOT과 충돌하고, 마커 없이 도는 개발용 콘솔에서
모든 라인이 사라진다. 미배선의 정직한 답은 "모른다"이고, 모를 때 이 화면이 이미
쓰던 답이 나이 판정이다.

## D3 — quarantine은 이 change의 안전 전제다

**결정**: 활성 exit-snapshot quarantine을 콘솔이 읽고, 걸린 포지션의 라인을 닫는다.
판정 순서는 quarantine이 liveness보다 **먼저**다.

**근거**: D1은 "엔진이 돌고 있으면 저장된 선이 살아 있다"고 주장한다. quarantine은
그 주장의 **정확한 반례**다 — 엔진은 돌지만 그 포지션만 판정에서 빠져 있다
(`exitloop.go`가 `ErrExitSnapshotQuarantined`로 건너뛴다). quarantine을 읽지 않은 채
D1만 적용하면, 보호되지 않는 포지션에 실행 가능한 보호선을 그린다. 지금 그 행이
비어 있는 것은 우연히 나이로 걸린 덕이므로, 나이를 치우는 순간 그 우연이 사라진다.

**generation 일치가 조건이다.** quarantine의 활성 유일 인덱스는
`(position_id, position_generation)`이고, 엔진은
`ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)`로 **현재 generation만**
확인한다. 콘솔의 read도 `q.position_generation = p.instance_seq`를 건다 — 해제된 과거
세대의 quarantine이 재편입된 현재 세대를 막으면 안 된다.

**읽는 자리**: `ReadOnly.LivePositionExits`가 이미 (position, exit_state)를 병합한다.
같은 계좌 read에 세 번째 statement를 더해 `PositionExit.Quarantine`을 채운다. 콘솔이
따로 상관관계를 계산하지 않는다는 것이 요점이다 — 상관관계를 두 곳에서 계산하면
언젠가 두 답이 갈린다.

**비용**: 계좌당 SELECT 1회, 로컬 SQLite, 화면당. `AccountRefs`·`accountExitStates`·
`positionSelect`가 이미 도는 자리다. 브로커 호출은 0이고 §0.4 예산과 무관하다.

## D4 — 새 사유는 문자열이고, 값 계산은 여전히 없다

`operatorview.BuildExitLine`의 구조는 바꾸지 않는다. 이미 계약이 "호출자가 신선도를
판정하고 이 어댑터는 fail-closed 표시만 한다"이므로, 콘솔이 새 `StaleReason`을
넣으면 그대로 동작한다. 더하는 것은 두 가지뿐이다.

1. `reasonText`에 사유 3건: `engine_not_running`, `snapshot_quarantined`,
   `engine_liveness_unknown`.
2. stale의 **상태 문구**를 사유에 따라 나눈다. 지금은 무조건 `오래된 평가`인데,
   엔진이 멈춘 것과 판정이 격리된 것은 운영자의 다음 행동이 다르다.

| 사유 | 상태 문구 |
|---|---|
| `snapshot_quarantined` | `판정 격리` |
| `engine_not_running` | `엔진 정지` |
| 그 외(기존 3건 포함) | `오래된 평가` (변경 없음) |

`Fresh()`/`Stale()`/`Unknown()` 서술어와 상태 코드(`fresh`/`stale`/`unknown`)는
그대로다. 템플릿의 `{{if .ExitLine.Fresh}}ok{{else if .ExitLine.Stale}}bad{{else}}notice{{end}}`가
계속 맞고, quarantine과 엔진 정지는 `bad`(빨강)로 나온다 — 맞는 색이다.

**대안으로 버린 것**: 새 Status 코드 추가. 템플릿 분기 3곳과 JSON 어댑터
(`httpapi.ExitLineFrom`)가 함께 바뀌고, 얻는 것은 문구 하나다.

## D5 — `/position-management`의 이름은 peek에서 온다

**결정**: `positionpolicy.State`에 `Name`을 넣지 않는다. 뷰 모델
`policyRowView.Name`에 핸들러가 붙인다.

**근거**: `positionpolicy.State`는 엔진이 소유하는 lifecycle 도메인 타입이고
journal이 그 필드를 저장하지 않는다. 표시용 문자열을 도메인 타입에 실으면 그 필드는
어떤 경로에서는 채워지고 어떤 경로에서는 비는 값이 된다.

**출처**: `c.holdings.peek(now)`. `get`이 아니다 — 이 화면은 지금 브로커 호출을
하지 않고, 그 성질을 유지한다. 캐시는 `/positions`가 채우고 `/dashboard`가 이미 같은
방식으로 읽는다(overview).

**키**: `symbolKey(market, symbol)` — join이 쓰는 것과 같은 함수. 시장을 키에 넣는
이유도 같다(KR `005930`과 미국 상장 동명 티커는 다른 종목).

**빈 캐시**: 이름 없음. 지금 화면과 동일하다. 없는 이름을 만들지 않는다.

## D6 — 정지선

- 새 route 없음, 새 POST 없음, 새 config 쓰기 없음.
- `HoldingsReader`는 여전히 메서드 1개다. `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`가
  그것을 지킨다. 새 journal read는 `*journal.ReadOnly`에 들어가므로 이 가드의 대상인
  주입 seam이 늘지 않는다.
- exit 판정 로직·주문·손절 즉시성에 닿지 않는다. 이 change가 만지는 것은 **표시**뿐이다.
- 나이 판정을 지우는 방향은 "덜 보수적"으로 보일 수 있다. 그러나 지워지는 것은
  **틀린 근거로 가리던 것**이고, 대신 들어오는 두 조건(엔진 정지, quarantine)은
  지금 존재하지 않던 보호다. quarantine 표시는 순증가다.
