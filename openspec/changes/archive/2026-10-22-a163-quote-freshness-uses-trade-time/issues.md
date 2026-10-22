# a163 issues

## I1 — `stableObservationID` 는 `ExitObserver` 메서드가 아니다 (safe local · 2026-10-22 구현 로트)

design D7 1 · tasks 1.2 는 「수신자 `ExitObserver` 의 … `stableObservationID` … 열 개」 라 적었다. HEAD(`exitloop.go:1202`)에서
`stableObservationID` 는 수신자 없는 함수다(`render-hash-base.txt` 는 이름 · 파일만 적었고 값은 그 자유 함수의 렌더). 렌더 해시 시험
(`a163_judgement_pin_test.go`)은 아홉 메서드는 수신자 `ExitObserver`, `stableObservationID` 는 수신자 없음으로 고른다. 기대값 열 개는
base 와 같다(편집 전 GREEN 실측). 설계 의도(판정 함수 열 개 본문 불변) 그대로 — 문구 정정만.

## I2 — `SourceTimestamp` 이름은 편집 전에도 `internal/official` 에 넷 있다 (safe local · Manager 확인 요청)

design D7 2 · tasks 1.2 는 census 가 「편집 전 출현 0 = 허용 밖 0」 이라 적었다. 실측(편집 전, HEAD `1eab9b931`):
`internal/official/strict_quote_reads.go` 의 a112 L1c 엄격 리더 — `type StrictLastPrice` 필드 1 · `type StrictTopOfBook` 필드 1 ·
`Client.StrictLastPrice` 리터럴 1 · `Client.StrictOrderbookTop` 리터럴 1. 다른 타입(`StrictLastPrice`·`StrictTopOfBook`)의 같은 이름
필드이고 생산 판정 소비자는 0(`officialbars` 만 읽음 — proposal 비목표 「a112 잔여 (k)」).

처분(구현 로트): census 의 허용 자리에 이 넷을 **정확한 수**로 더했다 — 새 출현(판정 경로로의 누출)은 여전히 실패한다. census 를
타입 정보(go/types)로 `domain.Quote`·`observedQuote` 필드만 세게 바꾸는 안은 고르지 않았다(설계는 식별자 census 를 정했고, 이름 기준이
더 넓게 막는다). 설계 의도를 넓히거나 좁히지 않는 판단이나, design 의 사실 진술과 다르므로 Manager 확인을 요청한다.

## I3 — 편집 대상 기존 함수가 tasks 1.1 목록보다 다섯 많다 (safe local)

design D8 이 정한 `recordGateSettings` 끝 인자 추가는 호출자 `engine.NewContext`(`engine.go:472`)와 시험 함수 넷
(`a074_notification_audit_test.go` 의 `TestNotificationSettingsAreAudited` · `TestTheAuditTrailCarriesNoNotificationSecret` ·
`TestARefusedNotificationBlockIsAudited` · `TestExistingAuditEntriesKeepTheirOrder`)의 본문을 바꾼다. tasks 1.1 은 이 다섯을 적지 않았으나
게이트 ⑤(`check_analysis.py`)는 바뀐 기존 함수를 직접 계산하므로 다섯 모두 번들을 만들었다(`analysis/function-logic/`). 편집은 각각
호출 인자 하나(생산 `cfg.Engine.ExitSourceAgeReport`, 시험 `false`)이고 분기 · 반환 수 불변.

## I4 — 해독 불가 원문의 정화 (safe local · gstack 리뷰 R3)

design D4 「필드」 는 `source_timestamp` 를 「원문 앞 64 바이트」 라 적었다. 구현은 글자 경계에서 64 바이트 안으로 자르고 인쇄 불가 글자(개행 ·
제어문자 · 깨진 UTF-8)를 `?` 로 바꾼다 — 그 값이 알림 본문(`renderFields` → ntfy)에 그대로 들어가 줄을 끼워 넣거나 전송을 망치지 않게. 인쇄
가능한 원문은 그대로다. 해독도 64 바이트 초과면 하지 않고 unparseable(R10).

## I5 — 토글 ON 전 사람 결정: 재무장 진동 · 15 s 경계 진동 (잔여 · pre-flip 사람 결정, gstack 리뷰 R12 · R13)

- 시각 부재 ↔ 복귀가 번갈아 오면 design D4 의 재무장 규칙대로 그때마다 `source_missing` 이 다시 나간다. `NormalRelay` 는 Key 중복 제거가 없어
  여러 주기에 걸쳐 일반 등급 버퍼(64)를 채울 수 있고, 차면 capped · unmanaged 같은 다른 exit 일반 알림이 버려진다(주기당 8 상한은 한 주기 안만).
- 체결 간격이 15 s 근처인 종목은 `le_15s` ↔ `le_5m` 전이 줄을 체결마다 남길 수 있다(포지션당 하루 수천 줄 가능).
- 둘 다 토글 ON 일 때만이고 판정 · 제출과 무관. 실제 브로커가 시각을 간헐적으로 비우는지는 실측 0(H6). 사람이 토글을 켜기 전에 쿨다운 ·
  히스테리시스 여부를 정할 것 — 설계 변경이라 이 change 에서 하지 않음.

## I6 — 보고 줄 쓰기는 동기 (잔여 · pre-flip 사람 결정 — Manager 처분 2026-10-22, Codex 리뷰 R11 P1)

`SourceAgeLog` 쓰기는 청산 goroutine 안에서 동기다. 로그 싱크(stderr 파이프 · 파일)가 멈추면 그 goroutine 이 멈춰 **다음 주기의 손절 판정이 없다.**
이 위험 종류는 이미 있다 — 같은 goroutine 의 a090 `UnobservedLog`(`settleUnobserved`) · a129 `ExitOrderLog`(`endExitOrderCycle`) · a091 `ZeroFloorLog`
가 같은 엔진 로거에 같은 방식으로 쓴다. a163 은 토글 ON 일 때 그 쓰기 지점을 하나 늘린다(그 주기의 판정 · 제출 뒤). 비동기 · 유계 싱크로 바꾸는
것은 design D4 「생산 싱크」 의 변경이라 하지 않았다. 엔진 로거 전반의 비차단화는 별도 Story 후보.

## I7 — 쓰기 실패여도 래치가 전진 (잔여 · pre-flip 사람 결정 — Manager 처분 2026-10-22, Codex 리뷰 R14)

`obs.Logger` 는 쓰기 결과를 돌려주지 않으므로 ENOSPC 같은 실패에도 구간 래치가 전진해 그 전이 줄이 사라진다. 기존 a090 · a129 줄과 같은 성질.

## I8 — OFF 여도 업그레이드 첫 기동에 audit 행 하나 (수용, 적대 리뷰 P3-4)

`recordGateSettings` 의 `RecordChange` 는 이전 기록이 없으면 `"" → "false"` 행을 쓴다(a136 · a138 · a139 와 같은 선례). design D8 「OFF 의 정의」
(새 로그 줄 · 알림 · 래치 0)에는 audit 파일이 들어 있지 않다. 그 뒤 같은 값 기동은 0 행(시험 `TestA163AStartupAuditsTheLoadedReportToggle`).

## I9 — 변이 ⑧ 은 시한 초과로 잡힘 (기록)

어댑터가 `timestamp` 때문에 오류를 내면 e2e 의 가격 읽기가 Retrier 재시도 대기(가짜 시계)에 멈춰 시험이 시한(하네스 4m)으로 실패한다. 잡히기는
하나 실패 이유 문장이 시한 초과다. 생산에서는 그 주기 관측 실패(B1) → 두절 사다리 — 이 change 가 만들지 않는 경로.

## I10 — 환경: 빈 `/tmp/.git` (기록)

구현 로트 중 빈 `/tmp/.git`(2026-10-22 06:20 생성, 출처 미상)이 있어 이 worktree 의 Go VCS 스탬핑이 exit 128 로 실패했다
(`TestAssembleConsoleSystemUpdateUsesRealTossctlBinary` 한 건 — `GOFLAGS=-buildvcs=false` 로 원인 확인). Manager 가 지운 뒤 같은 시험을 보통
실행으로 다시 돌려 PASS.

## I11 — 후속 Story 후보 (tasks 5.1 — 이 change 에서 고치지 않음)

`analysis/quote-time-evidence.md` 의 소비자 표와 사람 결정 H3 에 따른 잔여:

- **C7** 진입 게이트의 시세 신선도는 받은 시각 기준(브로커 시각 미사용) — 진입 쪽 거름은 H6 분포 뒤 별도 결정(design D1 B 기각 사유).
- **C8 · C9** 재편입 · 자동 편입의 t0 기준가에 나이 검사 0.
- **C10** 비상 청산(flatten) 지정가에 나이 검사 0 — flatten-all 즉시성(안전 4)에 직접 닿으므로 「보고만」 외의 변경은 별도 Story · 적대 리뷰.
  (이 change 는 금지 census 를 모듈 전체로 넓혀 flatten 이 브로커 시각을 **읽기 시작하면** 시험이 실패하게 했다 — 변이 ㉗.)
- **C12** CLI 바닥글 시각(받은 시각 표시).
- **a112 잔여 (k)** `officialbars.PollQuoteL1` 호가 경로의 나이 상한 — 생산 호출자 0, 배선하는 change 의 몫.
- **H1 후속** `source_old` 알림의 임계 · 정규장 한정 — H6 분포(토글 ON 기동부터) 뒤. 전이 줄로는 구간 **내부** 분포를 얻지 못한다(design D4 한계).
- **freeze 리뷰 P3-7(기존 위험)** `apiPrice.Timestamp` 디코드 타입(string)이 바뀌면 `Prices` 전체가 디코드 오류 → 그 주기 판정 0. a163 은 이 표면을
  넓히지 않음(디코드 타입 불변).
- **보고 토글의 콘솔 스위치 부재** 사람 flip 은 config 편집 + 엔진 재기동(KST 05:00~09:00 창), 기동 audit 행이 남음. 콘솔 스위치가 필요하면 별도 Story.
- **I5 · I6 · I7** (위) 토글 ON 전 사람 결정(Manager 처분 2026-10-22 — 기본 OFF 에서는 셋 다 없음, flip 전 결정 필수) · 엔진 로거 비차단화.
