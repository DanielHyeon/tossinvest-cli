# Design: refresh-positions-screen

## D1 — 사유 안내의 자리: 행이 아니라 상태의 스코프

`positionRow.Reason`의 4분기 중 2개는 사실상 페이지 전역 상태다:

| 분기 | 스코프 | 새 자리 |
|---|---|---|
| 원장 미판독(`Unknown`) | 전역 — journalstate 안내와 같은 원인 | 페이지 공지 1회 |
| 원장에 포지션 없음(`!InJournal`) | 해당 상태의 모든 수동 보유에 동일 | 페이지 공지 1회 |
| 자격 기록 없는 원장 포지션(`!Eligible`) | 행 고유(어떤 기록이 없는지) | 행 유지 |
| 자격 있음·exit 미개설(`!HasExit`) | 행 고유(전이 상태) | 행 유지 |

전역 사유의 문장은 지금 문구를 유지하되 단수 지시("이 종목")만 복수 안내로 고친다. 라벨
컬럼("관리 외(미편입)"/"관리 여부 불명")은 행마다 남으므로 행→공지의 대응은 눈으로 따라갈
수 있다. 기존 표 하단 각주(관리 외의 의미·편입은 대사 루프 몫)는 그대로 두고, 공지는 표
위에 상태별 최대 1개씩만 렌더한다.

구현: `positionsView`에 새 leaf 도우미(`AnyUnknown`/`AnyJournalAbsent`)를 추가하고,
`Reason()`은 전역 2분기에서 빈 문자열을 반환한다. 행의 둘째 `<tr>`은 내용이 있을 때만
렌더한다(빈 회색 줄 방지, `HasDetail`).

## D2 — 갱신: 브라우저 재로드, 서버는 그대로

무스크립트 규칙(검증 화면과 동일 메커니즘)을 유지하므로 선택지는 meta refresh뿐이다.
주기는 `holdingsTTL`과 같은 30초로 못 박는다:

- 재로드마다 서버는 lazy 판정을 그대로 탄다 — TTL이 지나야 holdings 1콜. 열린 탭 하나의
  상한 = 1콜/30초(§0.4 계상: 기존 "수동 새로고침" 비용과 동일 상한, 빈도만 규칙화).
- 검증 실행 중에는 `verifyHold`가 캐시 서빙을 강제하므로 재로드 비용 0콜(기존 SHALL 유지).
- TTL과 같은 주기면 재로드 시점의 캐시 나이는 ≈TTL이라 사실상 매 재로드가 1콜이 된다 —
  이것이 의도한 상한이며, 표시 나이는 페이지에 그대로 남는다.

head 템플릿의 `content="2"` 하드코드는 `{{.RefreshSeconds}}`로 바꾼다. `Refresh`가 참인
페이지만 그 값을 요구하므로 verify/dashboard에 `RefreshSeconds() int { return 2 }`(동작
불변), positions에 `30`을 준다. history·report는 `Refresh` 거짓 유지.

주의: static_test가 `.Refresh(` 호출 문자열을 금지한다(runlock 가드) — Go 코드에서 이
메서드들을 호출하지 않고 템플릿에서만 읽는다(현행과 동일).

## D3 — 손대지 않는 것

- `holdingsCache`·TTL·single-flight·hold 로직 무변경(주석의 "left open overnight costs
  nothing"만 현실에 맞게 갱신)
- 라우트 표·CSRF 게이트·read-only 가드 무변경(GET 그대로)
- 스펙이 고정한 라벨 문자열·정의 위치(portfolio.go 1곳) 무변경
