# a085 · Tasks

## 1. 근거 고정

- [x] 1.1 알림 조립부 14곳을 열거하고 각각이 종목을 가리키는지 분류한다.
- [x] 1.2 CodeGraph로 `reconcile.Holding`의 사용처를 확인하고 비교 경로가 이름을
      읽지 않음을 기록한다.
- [x] 1.3 `operatorview.BuildExitLine`의 호출부와 stale 사유 목록을 확인한다.
- [x] 1.4 Function Logic Map: `BuildExitLine`은 기존 함수 내부 분기를 바꾸므로
      작성한다. `Collector.holdings`는 필드 추가만이므로 함께 판단해 기록한다.

## 2. RED — 이름 배선

- [x] 2.1 `official.RawHolding`에 `Name`이 담기고 `HoldingsRaw`가 채운다. **실패한다.**
- [x] 2.2 `reconcile.RawHolding`·`reconcile.Holding`까지 이름이 전달된다. **실패한다.**
- [x] 2.3 회귀: 이름이 다른 두 holding의 수량 비교 결과는 같다 (비교 오염 없음).
- [x] 2.4 회귀: 이름이 비어도 스냅샷·비교·편입이 지금과 동일하게 동작한다.

## 3. RED — registry

- [x] 3.1 대사 주기가 registry를 갱신한다.
- [x] 3.2 보유가 사라져도 이름이 남는다 (청산 직후 알림이 이름을 잃지 않는다).
- [x] 3.3 registry에 없는 심볼은 코드만 쓰고 추가 브로커 요청을 하지 않는다.

## 4. RED — 알림 문구

- [x] 4.1 종목을 가리키는 알림 제목·본문이 한국어이고 `이름(코드)` 형식이다. **실패한다.**
- [x] 4.2 이름을 모르면 코드만 쓴다.
- [x] 4.3 회귀: `Fields` payload의 키와 값이 영문·원문 그대로다.
- [x] 4.4 회귀: 구조화 로그 필드가 바뀌지 않는다.
- [x] 4.5 회귀: 알림에 계좌번호·잔고·세션이 들어가지 않는다.
- [x] 4.6 원문 에러를 포함하는 알림은 한국어 설명 뒤에 원문을 그대로 붙인다.

## 5·6. operatorview / console — **a086으로 분리**

- [x] 5.0 a077의 승인된 요구사항과 충돌함을 확인하고 proposal에 분리 근거를 기록한다.
      승인된 requirement 수정에는 별도 리뷰 게이트가 필요하다(WORKFLOW 리뷰 게이트 표).

## 7. GREEN

- [x] 7.1 `Name` 필드를 official → reconcile까지 배선한다.
- [x] 7.2 엔진 registry와 `이름(코드)` 포맷터를 추가한다.
- [x] 7.3 알림 14곳의 제목·본문을 한국어로 바꾸고 포맷터를 쓴다.
- [x] 7.6 2~4장 전부 GREEN.

## 8. VERIFY

- [x] 8.1 §0.4: 이 change가 브로커 요청 수를 늘리지 않음을 테스트로 고정한다.
- [x] 8.3 실제 alert_outbox 사례("the exit policy could not judge 032820")가
      한국어 + 이름(코드)로 렌더되는지 확인한다.
- [x] 8.4 upstream 상속 테스트 회귀 없음 (650 green).
- [x] 8.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 8.6 `make gate CHANGE=a085-an-alert-says-which-stock`.

## 9. 리뷰와 기록

- [x] 9.1 경량 리뷰(validate + Manager 셀프리뷰)를 `review.md`에 기록한다.
- [x] 9.2 `issues.md`에 남긴다 — 대사 차단 evidence 문자열이 첫 관측에 고정되어
      화면이 오래된 숫자를 현재값처럼 보여주는 문제(a083 issues와 동일 건).

## 10. 개정 2 — 독립 리뷰 (2026-08-05)

- [x] 10.1 D5: 이름을 `stabilise`의 두 수집 각각에서 배운다. 요청 수 불변.
- [x] 10.2 D6: §0.8 주장을 실제 검사 범위로 좁히고, 발행 표면의 계좌 노출을
      `issues.md` B2에 별도 change 후보로 남긴다.
- [x] 10.3 I6: 한글 판정을 14음절 표본에서 Hangul 범위로, 이름 판정을 괄호 금지에서
      실제 레이블 비교로 바꾼다.
- [x] 10.4 `issues.md`에 B1·B2와 I5~I8을 기록한다.
- [ ] 10.5 gstack 독립 리뷰 재실행 + `make gate`.

## 11. 게이트 슬라이스 정합 (2026-08-05)

a083·a084·a085가 한 브랜치에 쌓이면서 `make gate`가 change를 서로 섞어 보게 된 문제를
슬라이스 단위로 정리한다. `check_analysis.py`는 각 change의 `base-commit.txt`부터
worktree까지를 diff하므로, 스택 위쪽 change의 게이트가 아래쪽 change의 수정까지 자기
것으로 계산한다. 사용자 결정(2026-08-05): change별로 자기 커밋 시점에서 게이트를 돈다.

- [x] 11.1 a083은 base `53626032` → `8dba0173` worktree에서 1~4단계 통과.
- [x] 11.2 a084는 base `8dba0173` → `ac2bbfcc` worktree에서 1~4단계 통과.
- [x] 11.3 a085 슬라이스(base `ac2bbfcc` → HEAD)는 개정 2~4의 수정까지 포함한다.
      그 구간에서 수정된 기존 함수 21개의 Function Logic Map을 만들고 stale AST 10건을
      갱신했다. 이 슬라이스가 "실제 배포되는 코드"를 증명하는 유일한 슬라이스다.
- [x] 11.4 `isFullExit`은 본문이 바뀌지 않았고 base 쪽 hunk만 줄 범위에 닿아
      `revision: base` 증거로 남긴다. 바로 아래 새로 들어온 `isProtective`가 원인이다.
- [x] 11.5 `Tracker.Observe`의 Branch Test Map은 a083 것을 복사하지 않고 다시 만든다 —
      개정 3이 `answerableBlockFor`(시계 비교)를 `hasBlockFor`(존재 비교)로 바꾸면서
      B21의 조건 자체가 달라졌다.
- [ ] 11.6 gate 5~8단계(`sdd-check`·`test`·`vet`·`validate`)는 change별이 아니라
      저장소 전체 검사이므로 HEAD에서 한 번만 돈다.

## 12. 증거 감사 (2026-08-05, 네 번째 독립 리뷰)

- [x] 12.1 B3: Branch Test Map의 테스트·판정 열을 주장에서 **측정**으로 바꾼다.
      `go test -covermode=count` 프로파일에서 분기별로 어떤 테스트가 그 줄을 실행하는지
      재어 렌더한다. 덮이지 않은 분기는 그렇게 표시된다 — 45개 target에서 183개.
      `_test.go` target은 계측 대상이 아니므로 "함수 자체가 테스트, 통과가 실행 증거"로
      구분해 적는다.
- [x] 12.2 B4: `Tracker.Observe`와 `Journal.recordExitJudgementTx` map을 HEAD에서 다시
      쓴다. 전자는 a083 복사본이 개정 3이 없앤 `t.adjusted = nil`을 현재 동작으로,
      후자는 a084 복사본이 개정 3 이전 시그니처와 폐기된 각인 추론 규칙을 적고 있었다.
- [x] 12.3 46개 map 헤더(줄 범위·`source_sha256`)를 `ast.json`에서 기계 생성한다.
- [x] 12.4 `ExitObserver.workingSet`의 callee 표에 남아 있던 원시 Python dict 출력을
      실제 호출부 표로 바꾼다.
- [ ] 12.5 `make gate`.

## 13. 증거 감사 2차 (2026-08-05, 다섯 번째 독립 리뷰)

리뷰가 코드에서는 결함을 찾지 못했고, 증거에서 다섯 건을 찾았다. 뿌리는 하나다 —
개정 5의 코드 변경 뒤 `ast.json`은 재생성했는데 `.md` prose·헤더는 그 *전에* 쓴 것이라
19개 map이 자기 AST와 어긋났다. 순서를 고정한다: 코드 → AST → prose → 헤더 → 측정.

- [x] 13.1 F1: `releaseReJudgedQuarantineTx` map을 개정 4 형태로 다시 쓴다.
      개정 3의 `reJudging bool` 시그니처와 2개 분기를 적고 있었고, 지금은 3개 가드다.
- [x] 13.2 F2: `ExitObserver.record` map의 `ReJudging` → `ReJudgingVersion`.
- [x] 13.3 F3: 19개 헤더를 `ast.json`에서 다시 생성한다. 대조 결과 불일치 0.
- [x] 13.4 F4: B3의 커버 테스트가 틀렸다 — 측정 대상 테스트 목록을 map이 이미 적어 둔
      것에서만 뽑았기 때문에, 아직 아무 map에도 없던 새 테스트가 빠졌다. 이 change가 쓴
      테스트를 전부 후보에 넣어 다시 측정한다.
- [x] 13.5 F5는 결함이 아니다 — `issues.md` I9. 새 함수는 `modified existing function`이
      아니다.
- [ ] 13.6 `make gate`.
