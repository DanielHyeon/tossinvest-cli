# Tasks: refresh-positions-screen

> 선행: proposal-freeze 경량 리뷰(UI change — validate + Manager 셀프리뷰 + review.md 기록).
> 파일 표면은 `internal/console`과 이 change 디렉터리·PM registry뿐이다 — 활성 change
> add-net-rr-measurement·verify-execution-capability와 겹치지 않는다.

## 1. 화면 동작 [T]

- [x] 1.1 RED: ① 원장에 없는 보유 2종목 이상에서 원장 부재 사유 안내가 페이지에 정확히 1회
      렌더됨(행 라벨 "관리 외(미편입)"은 행마다 유지) ② 원장 미판독 상태에서 "단정하지
      않는다" 안내 1회 ③ positions 응답에 `meta http-equiv="refresh" content="30"` 존재
      ④ verify 실행 중 화면의 `content="2"` 유지
- [x] 1.2 GREEN — 안내 dedup: `Reason()` 전역 2분기(미판독·원장 부재)를 빈 문자열로 옮기고
      `positionsView`에 leaf 공지 도우미 추가, 템플릿은 표 위 상태별 공지 1회 + 내용 있는
      행만 둘째 `<tr>` 렌더. 기존 문구는 복수 안내로만 손본다(라벨 정의 위치 불변 —
      static_test 가드 유지)
- [x] 1.3 GREEN — 자동 갱신: head 템플릿 `content="{{.RefreshSeconds}}"` 파라미터화,
      verify/dashboard `RefreshSeconds()==2`(동작 불변), positions `Refresh()==true`·
      `RefreshSeconds()==holdingsTTL초`. history·report 무변경. Go 코드에서 `.Refresh(`
      호출 금지(가드 테스트) — 템플릿 전용
- [x] 1.4 주석 정합: portfolio_pages.go 헤더("no meta refresh") 및 holdings.go("left open
      overnight costs nothing")를 새 비용 상한(열린 탭 1개 = holdings ≤1콜/TTL, 검증 중
      0콜)으로 갱신

## 2. 완료 게이트 [M]

- [x] 2.1 Function Logic Map: 수정된 기존 함수(`positionRow.Reason`,
      `positionsPage.Refresh`) 산출물 작성 + `check_analysis.py` 통과
- [x] 2.2 PM registry `bootstrap_change_allowlist`에 등재 + `generate_master_tracker.py
      --check` 통과
- [x] 2.3 `go test ./internal/console/ -race` + `make test`(upstream 회귀 0) + `make vet` +
      `openspec validate refresh-positions-screen --strict --no-interactive` + `make
      sdd-sync` + `make gate CHANGE=refresh-positions-screen`
