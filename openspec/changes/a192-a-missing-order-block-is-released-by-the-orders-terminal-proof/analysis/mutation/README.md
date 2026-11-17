# 변이 원장 (a192 tasks 4.1)

하네스: `mutate.py`(이 디렉터리 — 저장소 경로를 파일 위치에서 유도, 옮긴 자리에서 실행함). 규율: 시작 때 HEAD = base 단언, 대상 소스 sha256 을
끝에 다시 대조(원본 무변경), 변이마다 pid 를 이름에 넣은 **새 사본**(`go.mod`·`go.sum`·`internal/`·`cmd/`)에서 돌리고 사본을 지움(원복에
git checkout 을 쓰지 않음), 무변이 대조군 m0 이 모든 (패키지, -run) 조합에서 GREEN 이어야 진행, `GOFLAGS=-trimpath -buildvcs=false` + 전용
GOCACHE, 한 번에 한 판. 「닿음」은 두 겹: 구조(바꿀 문장이 사본에 정확히 한 번) + 행동(각 시험 안의 양성 대조 — 예: 쓰기 결과 계수
`entered`, 해제 시도 횟수 `calls`, 끼워 넣은 쓰기가 실제로 일어났음, 무변이 대조 입력이 무장함). **빌드 실패는 CAUGHT 로 세지 않음(NOT_REACHED).**

| 실행 | 원문 | 결과 | 비고 |
|---|---|---|---|
| 1 | `mutation-run-1.txt` | CAUGHT 25 · SURVIVED 4 · NOT_REACHED 0 | **오분류 있음** — m10·m12·m24 는 컴파일 실패였는데 CAUGHT 로 셈(코드 리뷰 1 #2 지적). 살아남은 m2·m17·m22·m25 |
| 2 | `mutation-run-2.txt` | CAUGHT 29 · SURVIVED 0 | 빌드 실패 분류 고침, 잘못 쓴 변이 3 수정, 살아남은 4 에 시험 추가(아래) |
| 3 | `mutation-run-3.txt` | CAUGHT 30 · NOT_REACHED 2 | 리뷰 1 수정 뒤 소스. m11·m12 의 문장이 콜백 개편으로 사라짐(닿지 않음) → 패턴 고쳐 4 에서 돌림 |
| 4 | `mutation-run-4-m11-m12.txt` | CAUGHT 2 | m11(시한 120s 초과 — tx 안 공개 읽기의 자기 대기), m12 |
| **5(최종)** | `mutation-run-5-final.txt` | **CAUGHT 33 · SURVIVED 0 · NOT_REACHED 0** | 최종 소스(리뷰 2 수정 포함), m1~m33 전부 |
| Manager 표본 1(시험 추가 전) | `manager-sample-run-1-before-test.txt` | CAUGHT 10 · SURVIVED 2 | Manager 독립 하네스, g1~g12. 살아남은 g8(해제 UPDATE 의 `released_at IS NULL` 삭제)·g9(`RowsAffected == 1` 삭제) — tx 안 재판정의 C2(`state.Active()`)가 먼저 막아 장부 쪽 고정이 혼자 일한 적이 없었음 |
| 6 | `mutation-run-6-manager-sample.txt` | CAUGHT 12 · SURVIVED 0 | g1~g12 를 m34~m45 로 옮김(문장 그대로) + 시험 `TestA192TheReleaseNeverOverwritesARowAlreadyReleased` 추가 뒤. 무변이 대조군 5조합 GREEN, 원본 sha256 무변경 |
| 7 | `mutation-run-7-g8-g9-detail.txt` | CAUGHT 2 | m41(g8)·m42(g9) 의 단언 문장까지 남긴 재실행: 둘 다 `released=true err=<nil>` 단언(시험 271행)에서 죽음. g8 은 그 전에 행을 덮어썼지만 시험은 첫 단언에서 멈춰 「행 그대로」 단언까지 가지 않음 |

## 살아남았던 것과 그 실측 사유(1회차) — 「동등 변이」로 넘기지 않고 시험으로 못 박음

| 변이 | 왜 살았나(실측) | 못 박은 시험 |
|---|---|---|
| m2 재기동 결속 | 변이 코드가 `updateOrderTerminalBindings` 의 「결속 0 이면 반환」 뒤에 있어 **행동이 닿지 않음**(구조 대조만 통과). 변이를 갈래 전(`begin…`)으로 옮기고, 재기동 시험을 「차단 중 재기동 → 다시 missing 관측 → 종결」 순서로 강화 | `TestA192ARestartedTrackerNeverReleasesThisWay` |
| m17 Terminal nil 가드 | 결속은 `Terminal != nil` 일 때만 생기므로 가드가 겹쳐 있음(우연이 지키는 안전) | `TestA192CandidateRefusesWithoutTerminalOrJournal`(결속을 직접 넣고 Terminal·Journal 이 없을 때 판정하지 않음) |
| m22 disputed 배관 | 깨끗한 갈래의 분쟁은 외부 포지션뿐인데, 그것이 갈래 전 갱신에서 이미 실격 흔적을 세워 C3 가 먼저 막음 | `TestA192CandidatePassesTheDisputedSymbolThrough`(흔적 없는 결속 + 분쟁 집합 → 무장 안 함, 대조: 분쟁 없으면 무장) |
| m25 계좌 배관 | 시험의 비교 계좌가 늘 추적기 계좌와 같았음 | `TestA192AForeignComparisonDoesNotDropTheBinding`(다른 계좌 비교 뒤에도 결속 유지 → 제 계좌 비교로 해제) |

## 변이 목록(최종 45 — m34~m45 는 Manager 표본 g1~g12)

m1 삽입 증명 무시 · m2 재기동 뒤 결속 · m3 워터마크 비교 뒤집기 · m4 무장 없이 해제 · m5 digest 무효화 삭제 · m6 C3 · m7 C5 · m8 C7 fail_closed ·
m9 C8 엄격→이상 · m10 C10 · m11 tx 안 공개 읽기 · m12 tx 재판정 생략 · m13 전부→하나라도 · m14 빈 표본 · m15 해제 원인 · m16 종결 판정을 크레딧 앞으로 ·
m17 Terminal nil 가드 · m18 로그에 주문번호 · m19 갈래 전 갱신 삭제 · m20 재시도 삽입 결속 · m21 Restore 결속 지우기 삭제 · m22~m26 배관(분쟁 집합·요청
워터마크·결속 워터마크·계좌·해제 tx 결속 복사) · m27 RunOnce 로그 호출 · m28 새 원인 허용 목록 · m29 buildGateway 배선 · m30 종목 없는 체결 · m31 1차 사유 ·
m32 tx 사유 · m33 사라진 블록의 결속 버리기. 문장과 시험 대응은 `mutate.py` 의 표와 각 실행 원문.
