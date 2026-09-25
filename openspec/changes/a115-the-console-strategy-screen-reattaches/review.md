# a115 리뷰 기록

## §0 proposal-freeze (2026-09-26)

위험 등급: 경량(콘솔 화면 — 주문·위험·원장 경로 무관, 읽기 전용 projection 소비) + 독립 적대 보이스 1.
보이스: Manager 셀프 4관점 + **독립 적대 Eng 리뷰어 1**(Opus 서브에이전트, 별도 컨텍스트, 읽기 전용).
`autoplan` 은 대화형 파이프라인이라 자율 세션에서 독립 서브에이전트로 대체했다(a114 와 같은 대체 경로).

### 셀프 4관점 (Manager)

- **CEO/Product** — 범위: a109 A2 P1-1 한 건의 이행. 합본 안 함(tasks 0.1). 화면 구조·문구 무변경, UI 마찰 0.
- **Eng** — httpapi 원형(a109 D4)·a114 선례 재사용. 판정 동치는 옮겨 적지 않고 테스트로 고정.
- **Security** — 읽기 전용 projection. client 권한 불변(같은 descriptor 검증·loopback). 주문·토글 경로 접촉 0.
- **QA** — 진짜 `strategyprojectionrpc.Start` 로 늦은 기동·재시작·펌프, 소비자 판정은 단위 시험 + 콘솔 실측 두 절반.

### 독립 적대 Eng 리뷰 — P0 0 · P1 6 · P2 10

AST 번들 5벌의 source_sha256 이 HEAD·base `8688f74f` 모두와 일치함을 리뷰어가 확인했다.

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P1-1 | 판정 두 벌 — `httpapi.StrategyRuntimeAbsent` 가 이미 「유일한 판정」(⛔ 주석)인데 design 이 콘솔에 `strategyRuntimeWired` 를 새로 만들었다. 구조적 인터페이스뿐이면 메서드 이름 변경이 조용한 회귀가 된다 | 수용 → 판정·presence 를 `internal/strategyprojection` 으로 이동, httpapi 는 alias·위임, 콘솔은 한 벌을 사용, `var _` 컴파일 결속 + 동치 시험(design D2 재작성) |
| P1-2 | 「시도 대상일 때만 wake」 펌프는 a109 G2 가 기각한 `failed` 게이트의 재도입 — live 로 보이는 죽은 자리가 안 드러나 첫 렌더가 「읽지 못했다」를 그린다 | 수용 → 무조건 wake(비용은 wake 안 rate limit·single-flight 가 간격당 1회로 고정, AST wake B1) |
| P1-3 | spec 의 「구성」에 운영상 대응물이 없다 — 엔진은 깨끗이 종료하면 descriptor 를 지우므로 가장 흔한 다운 모양에서 시나리오 1 이 충족 불가 | 수용 → 시나리오 1 WHEN 을 「descriptor 잔존 + 연결 불수락」으로 좁히고, descriptor 부재 → dormant 선언된 한계 시나리오·재시작 시나리오·요청 경로 dial 금지 SHALL NOT 추가. 양의 신호 대안은 기각 사유와 함께 design 에 기록 |
| P1-4 | design 의 재사용·펌프 논거가 wrapper 분기(attach·attempt·observe·Read·wake)에 기대는데 AST 번들이 없다 | 수용 → 인용 전용 번들 5벌 생성(`analysis/function-logic/cmd-tossctl--strategyruntimeattachment.*`) |
| P1-5 | evidence-reconciliation 의 「B38–B42(:397–407)」가 AST(B33–B37, :398–405)와 모순 | 수용 → 정정 |
| P1-6 | review.md 없이 tasks 0.4 가 [x] — 거짓 완료 표시 | 수용 → 이 기록과 같은 커밋에서 체크 유지(산출물 동시 착지) |
| P2-1 | 답한 실패(503·decode 거절)도 탈착 — 깜빡임 병을 콘솔이 물려받는다 | 기록 → issues R4(표면 밖) |
| P2-2 | 비부재 stat 오류 nil 접힘 · http-api spec 「반쪽 잔재」 문장과 코드 불일치 | 선언된 접힘으로 design 에 기록 · issues R5 |
| P2-3 | 재시도 경고 출력 대상 미명시(2880줄 위험) · 부팅 문구 「dormant로 뜬다」가 a115 뒤 거짓 | 수용 → design(io.Discard·문구 교체 + 시험) |
| P2-4 | nil wrapper 가 interface 에 담기면 presence 질문이 패닉 | 수용 → design 「반환은 언제나 non-nil」 |
| P2-5 | 간격 변수 공유(a109 시험이 30s 고정)·interval≤0 ticker 패닉·시험 goroutine 누수 | 수용 → design(콘솔 전용 변수·가드·정리) |
| P2-6 | 과도 창(발행 전·재시작 사이)이 시나리오에 없음 · 요청 경로 dial 금지 SHALL NOT 부재 | 수용 → spec 추가(P1-3 과 함께) |
| P2-7 | 규칙 13 실측이 dormant 절반만 | 수용 → tasks 1.6 두 절반 |
| P2-8 | issues 의 tasks 참조 오기 · proposal 낡은 좌표 · codegraph-baseline 의 `MarketScheduleReader` implements 오탐 | 수용 → 셋 다 정정 |
| P2-9 | 「건강한 자리는 깨우지 않는다」 문장이 부정확(소비자 presence 질문은 매 렌더 wake) | 수용 → design 문장 교체(P1-2 재작성에 포함) |
| P2-10 | console seam 은 `Read` 하나여야 한다(static_test:1010) — presence 를 seam 에 넣으면 깨진다 | 수용 → design 구현자 주의 |

리뷰어가 공격했으나 무너지지 않은 것: 기존 시험 회귀 없음(스텁·raw client 는 presence 없음 → 판정 불변) ·
presence→Read TOCTOU 안전 · single-flight/rate limit · 취소 비판정 · design 분기 주장과 AST 일치 ·
engineDir 해석 동일.

**판정: PASS with P1 — P1 여섯 전부 반영 완료(이 커밋), freeze 성립.** 구현(1.x)은 별도 Teammate.

## 콘솔 실측 (규칙 13, tasks 1.6 — 2026-09-26)

하네스 `analysis/console-measure/measure.sh <bin> <clean|dead> 18475`. 바이너리 둘: a115(`d4d667f4` 트리의 `git archive`
빌드)와 base(`8688f74f` 같은 방식). 격리 config(`mktemp -d /tmp/a115-console-XXXX`, 0700), 엔진 없음. `dead` 는 같은
config 에 잔재 descriptor(`.strategy-runtime-read/endpoint.json` 0600, 필드 유효) + 죽은 socket(`runtime.sock` 0600 —
bind 뒤 close, listener 없음). 요청은 GET 셋뿐(세션 링크 303 · `/strategy-runtime` 200 · `/settings/strategy` 200) —
버튼·폼·POST 없음. 매 판 엔진 프로세스 없음·config 에 새 descriptor 없음 확인(`dead` 잔재는 콘솔이 지우지 않음 — 조회 전용).

| 바이너리 / config | 전략 화면 안내 줄 | 시장 코드(본문 출현 수) | 설정 요약 | 콘솔 stderr(전략) |
| --- | --- | --- | --- | --- |
| a115 / clean | 「runtime endpoint 미기동 — … dormant … truth」 | NOT_CONFIGURED ×10 | 「… — dormant 미배선」 | 없음 |
| a115 / **dead** | **「runtime projection을 읽지 못했다. …」** | **RUNTIME_UNAVAILABLE ×8** | **「읽지 못함 — …」** | 「…지금 연결할 수 없다 (socket has no listener). 전략 화면은 붙기 전까지 도달 불가를 표시하고, 엔진이 돌아오면 콘솔 재시작 없이 다시 붙는다.」 |
| base / clean | 「runtime endpoint 미기동 …」 | NOT_CONFIGURED ×10 | 「… — dormant 미배선」 | 없음 |
| base / dead | 「runtime endpoint 미기동 …」 — **오귀속** | NOT_CONFIGURED ×10 | 「… — dormant 미배선」 | 「…연결할 수 없다 (…). 전략 화면은 dormant로 뜬다.」 |

판정: 두 절반 다 spec 대로다 — 깨끗한 config 는 미구성 그대로(오귀속 없음), 잔재 descriptor + 죽은 socket 은 도달 불가.
base 는 같은 잔재를 미구성으로 접었다(병의 재현 = 대조군). 측정은 curl 본문 대조다 — 템플릿·JS 무변경이라 브라우저 콘솔
오류 판정은 not-applicable(서버 렌더 문구만 바뀐 값). 엔진 기동 후 회복 절반은 실엔진 없이 못 재므로 시험
(`TestTheConsoleStrategyScreenShowsADeadDescriptorAsUnreachable` 등 진짜 `strategyprojectionrpc.Start`)이 진다.

## 검증 기록 (tasks 1.5 — 2026-09-26, 워킹트리 = `d4d667f4` + 병행 세션의 비-Go 문서뿐)

- `go test -race -count=1 ./cmd/tossctl/ ./internal/console/ ./internal/httpapi/ ./internal/strategyprojection/` →
  ok 176.4s · ok 503.0s · ok 1.4s · ok 4.5s
- a115·a109·a114 대상 시험 `-race -count=5` → ok(cmd/tossctl 7.8s · internal/console 1.9s)
- `make vet` rc 0 · `make lint` rc 0(`go vet ./...` + `go vet -tags tossos_testseams ./...`)
- `make test`(`go test -timeout 30m ./...`) rc 0 → **ok 99 · no test files 9 · FAIL 0**
- `check_analysis.py --change a115-…` rc 0(required 4, evidence complete)

## §1 구현 후 리뷰 (2026-09-26 — 900d7582..d92b4414)

구성: 독립 적대 Eng 서브에이전트 1(Opus, 별도 컨텍스트, 저장소 읽기 전용 — 실험은 `git archive` 사본
`/mnt/D/tmp-a115-review-*` 에서만, 전후 `git status --short -- cmd internal` 공백 확인, 사본 삭제). gstack `review`
pre-landing 관점의 대체(a114 와 같은 자율 세션 대체 경로). 리뷰어 실측: 네 패키지 `-race -count=3` 통과, 새 변이 19개
(생존 4 · 등가 2), `check_analysis` rc 0·AST 해시 HEAD 일치, 사본 cmd/tossctl 전체 실행의 FAIL 7 은 사본에 없는
`Dockerfile`·`tools/engine-autostart.sh` 탓(a115 무관).

**판정: PASS — P0 0 · P1 0 · P2 5.** 코드 결함 없음. 다섯 모두 수용했고 반영은 전부 시험·문서다(생산 Go 편집 0 →
FLM 재추출 불요, `check_analysis` rc 0 재확인).

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P2-1 | design 「깨끗한 정지 뒤 dormant」는 정지 뒤 부팅한 콘솔만 참 — live 로 붙어 있던 콘솔은 wrapper 가 격하하지 않아 엔진 복귀까지 도달 불가. 같은 디스크가 이력 따라 두 값 | 수용(기록) → design 정정 문단 · issues R6(선언된 한계, wrapper 편집은 표면 밖) · 핀 `TestAnAttachedConsoleShowsACleanStopAsUnreachable`. spec 시나리오 3 은 부팅 시점이라 무변경(Requirement 수정 없음) |
| P2-2 | 「부팅 해석은 lastTry 를 안 찍는다」 무핀 — R2b 생존 | 수용 → `TestTheConsoleBootLeavesTheFirstWakeFree`(간격 1h, 첫 presence 질문이 즉시 붙임) · K19 CAUGHT |
| P2-3 | 배선의 `if engineDir != ""` 자리 무핀 — 밖으로 나가면 cwd 상대 descriptor 를 두드리고 진짜 미배선이 사라짐, R17 생존 | 수용 → `TestRunConsoleNeverDialsTheStrategyProjectionItself` 가 게이트 본문 안의 호출을 센다 · K20 CAUGHT |
| P2-4 | 부팅 live 플래그 무핀 — R6 생존(거짓 「다시 붙었다」, 첫 렌더 전 사망 시 탈착 로그 소실) | 수용 → 재시작 시험에 live 부팅 `attached==true` 단언 · K21 CAUGHT |
| P2-5 | 절반 틱 not-applicable 선언 — R1 생존 | 수용 → AST 구조 핀 `TestTheConsoleStrategyPumpTicksAtHalfTheInterval` · K22 CAUGHT |

리뷰어가 공격했으나 무너지지 않은 것: 시나리오 넷(부팅 sentinel 결정적·부재 조용한 dormant·늦은 기동·렌더 없는 재시작)
· typed-nil(반환 언제나 non-nil, engineDir=="" 는 진짜 nil interface) · 요청 경로(잠금 없이 presence, wake 는 goroutine만,
Dial 의 200ms probe 는 ctx 무시하지만 한정) · 동시성(-race ×3, 펌프 종료, 정리 순서) · stat 오류 문구의 진실성(ENOTDIR →
dormant+경고, 경로 고치면 콘솔 재시작 없이 live — 실측) · httpapi alias·위임 회귀 없음 · static_test seam(`Read` 하나) 유지.
잔존(선언 유지): EACCES/ENOTDIR 에서 「runtime endpoint 미기동」 표기는 freeze 리뷰 P2-2 의 선언된 접힘.

재검증(3판 뒤): `go test -race -count=3` a115 대상 cmd/tossctl ok · 하네스 대조군 green · K19–K22 CAUGHT.
