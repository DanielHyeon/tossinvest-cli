# Review: refresh-positions-screen

## proposal-freeze (2026-07-27)

- 보이스 구성: 경량 리뷰 — Manager 셀프리뷰(WORKFLOW "위험 등급 가중": UI change는
  validate + 셀프리뷰 + 기록으로 충분). 주문·위험·원장·인증 경로 무접촉 확인이 전제다.
- `openspec validate refresh-positions-screen --strict --no-interactive`: PASS

### 발견과 처분

1. **[수용] 두 전역 공지의 배타성 명시** — "관리 여부 불명" 공지(원장 미판독)와 "원장 부재"
   공지(원장 판독됨)는 `JournalReadable`로 상호 배타다. 같은 페이지에 둘 다 뜨는 조합은
   없다. 템플릿 조건을 view 도우미로 몰아 테스트로 고정한다(1.1 ①②가 각각 검증).
2. **[수용] meta refresh 주기의 하한을 스펙에 명시** — 주기를 TTL 미만으로 줄이는 후속
   변경이 비용 상한을 깨지 않도록 delta에 "주기는 캐시 TTL 이상"을 SHALL로 못 박았다.
   구현은 상수를 `holdingsTTL`에서 유도해 드리프트를 막는다.
3. **[수용] 열린 탭의 24시간 비용 정직 계상** — 상한 holdings 1콜/30초 = 2콜/분. 검증·soak
   실행 중에는 runlock/in-process hold가 재로드의 브로커 콜을 0으로 만든다(기존 SHALL
   유지, 자동 재로드가 보류를 우회하지 못함을 delta에 추가). 야간 방치 탭의 콜은 무해한
   read이나 비용이 0이 아니게 되는 점은 holdings.go 주석에 남긴다(task 1.4).
4. **[거절] history 화면에도 자동 갱신** — 동결 값 표시라 갱신 요구가 없고, 사용자 요청
   범위(보유 종목) 밖. 필요해지면 별도 소품 change로.
5. **[거절] 종목별 편입 opt-in UI** — read-only 불변식과 route 가드에 정면 충돌. 편입은
   엔진 대사 루프 + `adoption` config가 소유(adopt-external-positions). 범위 밖 유지.
6. **[확인] static 가드와의 정합** — `.Refresh(` 호출 문자열 금지(runlock 가드)는 메서드
   선언·템플릿 필드 접근에 저촉되지 않는다. 라벨 "관리 외(미편입)"의 정의 위치
   (portfolio.go 1곳 + templates_portfolio.go 설명) 가드도 유지된다 — 공지 문구는 라벨
   문자열을 새로 철자하지 않는다.

### Function Logic Map

- 대상: `positionRow.Reason`(분기 이동), `positionsPage.Refresh`(반환 플립) — 산출물 작성
  (task 2.1). 신규 도우미(`AnyUnknown` 등)·`RefreshSeconds`는 새 leaf 함수라 대상 아님.
- High-risk 아님 → Pre-Edit 선언 불요. 다만 기존 테스트(TestAnUnmanagedHoldingIsLabelled
  ExactlyOnce 등)가 문구를 고정하므로 RED 선행으로 회귀 방향을 잡는다.

### 판정

FREEZE — 구현 착수 승인(위 처분 반영). Requirement 수준 재수정이 생기면 리뷰 재실행.

## 구현 검증 (2026-07-27)

### 게이트 실행 (격리 worktree — 구현 컨텍스트와 분리된 클린 체크아웃)

다른 활성 change(add-net-rr-measurement)의 미커밋 수정이 본 worktree에 있어
`check_analysis`의 base↔worktree diff가 오염되므로, 커밋 후 격리 worktree에서 게이트를
실행했다(2b 1.11 Manager 격리 검증과 동일 관행). 미커밋 상태의
`openspec/changes/add-net-rr-measurement/`는 registry stale 판정 방지를 위해 복사만 했다.

- RED: 대상 5 테스트 중 4 실패 확인(구현 전 — 공지 0회/반복 3회/meta 부재/신규 문자열 부재)
- GREEN: `go test ./internal/console/ -race -count=1` — ok (115 tests)
- `make gate CHANGE=refresh-positions-screen` — **PASS (8/8, exit 0)**
- 명시 재확인: `go test ./...` — **2907 passed / 52 packages**(upstream 회귀 0),
  `go vet ./...` — 이상 없음, `openspec validate --all --strict` — 14/14 valid

### 독립 검증 (별도 검증 에이전트 컨텍스트 — 작성자·검증자 분리)

판정 **ACCEPT**. 범위 초과 없음, 라우트 표·read-only 불변식 유지, rate budget 상한
(1콜/TTL·hold 우회 경로 없음) 재확인. non-blocking 발견 6건과 처분:

1. **[수정]** design.md의 `AnyUnadopted` 표기를 구현 명(`AnyJournalAbsent`)으로 정정(editorial)
2. **[수용-기록]** "관리 여부 불명" 문자열이 portfolio.go(Label)와 템플릿 공지 2곳에 존재 —
   스펙이 고정한 라벨은 "관리 외(미편입)"뿐이므로 가드 확장은 범위 밖. 드리프트 시 후속 처리
3. **[수용-기록]** reportPage는 `Refresh` 필드만 있고 `RefreshSeconds`가 없다 — 항상 false라
   도달 불가. true로 바꾸는 후속 change가 메서드를 함께 추가해야 한다(템플릿 실행 오류 방향
   이라 조용히 깨지지 않는다)
4. **[수정]** check_analysis.py null-branches 수정에 단위 테스트 부재 →
   `test_null_branches_from_the_go_extractor_are_accepted` 추가(테스트 규율)
5. **[수용-기록]** 검증 화면 2초 pin이 head 템플릿 직접 렌더(간접) — 핸들러 배선은 무변경
   기존 경로라 적정
6. **[수용-기록]** 원장 부재 공지의 "(편입은 엔진 대사 루프의 몫이다)" 문구는 기존 각주와
   정합인 서술 추가 — cosmetic

Function Logic Map: 수정 기존 함수 2건(positionRow.Reason·positionsPage.Refresh) + diff
교차 요구 9건(신규 leaf 7·본문 무변경 인접 1(base 리비전 AST)·테스트 1) 전부 산출,
`check_analysis.py` PASS. High-risk 경로 무접촉 — Pre-Edit 선언 불요.
