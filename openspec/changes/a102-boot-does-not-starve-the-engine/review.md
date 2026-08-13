# a102 리뷰 기록 — 적대 2왕복 + gstack 7패스

역할 분담(tasks.md §0): Manager(Fable)가 설계·발주·판정, T1/T2(Opus)가 구현,
각 Teammate에 전담 적대 리뷰어(A1/A2, Opus) 1명씩, 마지막에 gstack /review.
이 문서는 그 왕복의 판정 기록이다. 각 지적의 산출물 증거는 커밋 본문과
`analysis/` 번들에 있다.

## 1. A1 — T1(겹1, internal/reconcile) 적대 리뷰

대상: `1c76a580` (겹1 1판). 판정: **수정 요구** → T1이 `831cbc08`으로 반영 →
재검증 **MERGE-OK**.

| # | 지적 | 처분 |
|---|---|---|
| F1 | **설계 전제 붕괴** — Report를 소비하는 곳이 없다. `engine.go`가 `_, rerr :=`로 버려서 §1의 rate-limit 계수는 아무 데도 닿지 않는다 | 설계 정정 **D5b** (obs 한 줄) — T2 범위로 이관, 3.4b에서 구현 |
| N1 | 생존 뮤테이션 — 예산 경계 연산자를 바꿔도 초록 | 경계 산술 테스트(30s=2×15s)로 사멸 |
| N2 | 생존 뮤테이션 — `clk.Sleep` 오류 삼킴이 초록 | 취소 후 브로커 호출 수(`calls==1`) 단언으로 사멸 |
| F3 | `ratelimit.go` 취소 wrap이 `%v` — **이 change의 존재 이유와 같은 결함**을 새 코드에 재범 | `%w` + `errors.Is(err, context.Canceled)` 고정 |
| F4/F7 | 경계·대기 중 게이트 미고정 | 경계 테스트 + 대기 중 `CheckEntry` 단언 |
| F5 | flatten(liquidate.go:596)도 같은 Collect를 쓴다 — 429 정체가 도달하지만 구분 없음 | **선언된 생략** (design 범위 밖 — 비상 경로 인내는 별도 승인 필요) |
| F6 | 백오프 노브가 spec의 SHALL NOT(하한) 아래로 내려갈 수 있다 | `withDefaults` 하한 클램프 — 단, 포섭되는 죽은 분기를 만들지 않도록 `< Default` 한 분기로 |
| F8 | 예산 < 백오프 1회일 때 "kept being refused"는 거짓말 | 메시지 분기 + 정직성 테스트 |
| F14 | 겹1 단독 배포는 무보호 창을 0s→5m로 넓힌다 | design **배포 주의** 절 신설 |

## 2. A2 — T2(겹2, enginelock + cmd/tossctl) 적대 리뷰

대상: `6cd643ca` (겹2 1판). 판정: **수정 요구** → T2가 `9daa052c`로 반영 →
재검증 **MERGE-OK + 잔여 2** → §3.9b `9e184687`로 마감.

| # | 지적 | 처분 |
|---|---|---|
| F1 | **시체 마커의 가짜 ready** — 크래시는 Release를 안 부르고, 신선도만 보면 죽은 엔진의 ready_at이 "보호가 서 있다"로 읽힌다(02:03의 다음 장면을 실행 재현) | 설계 정정 **D4b** — ready는 마커 PID가 산 엔진 집합에 있을 때만; 열거 실패는 not-yet(모름≠없음) |
| N1–N4 | **생존 뮤테이션 4** — ready no-op·readiness 상수 0화·cap/poll 맞바꿈·형태 보존 `ready=nil` 전부 초록. "단위는 완벽한데 배선은 0%" | 설계 정정 **D5c** — 배선 자체를 실행 테스트로. stubRuntime이 ready를 잡고 마커 파일로 단언 |
| F3/F4 | **파일 경주 3종 실측** — ready_at 소거 139/3000, 찢어진 읽기 3617/12259(O_TRUNC 창), Release 뒤 refresh 부활(결정적) | 설계 정정 **D4c** — write를 뮤텍스 안, tmp+rename 원자화, live 플래그 |
| F5 | 비동기화가 연 부팅·버튼 이중 spawn 창 | 설계 정정 **D7b** — `soakSpawnGate` 직렬화 |
| F6 | 형태 테스트 관통 — `go func()` 문자열을 다 지키고도 콘솔을 120s 막는 판본 | 설계 정정 **D7c** — 영원히 안 돌아오는 start로 즉시 반환을 실행 단언 |
| F7 | obs 이벤트 값 충돌 가능 | 값 유일성 전수 테스트 |
| F8 | 콘솔 정상 종료가 "실패:"로 인쇄 | abandoned는 note (문장 완결은 §3.9b에서) |
| F9 | ctx nil 방어(전반)·Runtime.Run 취소 계약(후반) | 전반은 `recoverThenReady`에 방어, 후반은 **선언된 생략**(internal/app/engine 무변경이 이 change의 못) |
| N5 | 재검증에서 홀로 생존 — `engineRuntime` 본문의 ready 통과가 문자열 검사 하나 | §3.9b: `engineRecoverySequence` seam으로 조립된 Recover를 **실제 실행** — N5·N5b(본문 `ready=nil` 재대입) 사멸, 원장 20/20 |

## 3. gstack /review (§5.3) — 7패스

브랜치 전체(main 대비, 코드 +3,058/-86) 대상. 패스: Manager CRITICAL pass ·
전문 리뷰어 4(testing/maintainability/security/performance, Opus) · Claude 적대(Opus) ·
Codex 적대 · Codex 구조화. **A1·A2가 이미 훑은 코드에서 7패스가 새 결함을 더 찾았다.**

### 3.1 수렴 지도 (독립 패스가 같은 지점을 짚은 것)

| 지점 | 짚은 패스 | 처분 |
|---|---|---|
| **PID 재사용이 D4b를 뚫는다** (컨테이너 recreate: 신선한 전임자 마커 + 재생성된 PID namespace의 결정적 배정 + Hold 전 몇 초 창) | Codex 구조화 **P1** · Codex 적대 Critical · Claude 적대 F10 | **D4b-2** — 프로세스 인스턴스 토큰(boot_id+starttime ticks) 정확 일치. §3.9c ① |
| cap 120s가 스로틀 없는 5회 판독 아침(~128s)에도 걸리고, 초과는 fail-open | Claude 적대 F1 · performance · security · Codex 적대 | **D6-2** — cap을 겹1 예산(5m)에서 유도. §3.9c ② |
| nil ctx가 `clk.Sleep`에서 역참조 — 분리 goroutine 패닉 = 콘솔 다운 | security · testing · Claude 적대 F4 | 정규화 + 테스트. §3.9c ③ |
| cap 초과 노트가 실제 대기한 한도가 아니라 패키지 상수를 포맷 | maintainability · Claude 적대 F5 | `engineReadyNote(verdict, limit)`. §3.9c ④ |
| 비-429 오류 wrap의 `%v` 재평탄화 (stableSnapshot·replay·resolution) | Manager pass · maintainability · Claude 적대 F3 | 범위 밖 항목 정밀화(선언 유지) — replay 재시도는 취소·인수 의미론이 걸린 별도 설계 |

### 3.2 testing 전문가의 CRITICAL — 미고정 불변식 3

A2가 실측으로 재현·수정 확인한 것 중 **회귀 테스트로 스위트에 남지 않은** 것:

1. Ready↔refresh 동시 경주 — 커밋된 두 동시성 테스트 모두 refresh **전에**
   Ready를 부른다. D4c의 "write를 뮤텍스 안"을 되돌려도 초록. → §3.9c ⑤
2. Release 후 Ready 부활 — `!h.live` 가드 삭제가 초록. → §3.9c ⑥
3. staging cleanup 6곳 전부 도달 불가 — `TestAFailedWriteLeavesNoDebris`는
   이름과 달리 성공 경로만 밟는다. → §3.9c ⑦

"통과는 증거가 아니다"의 세 번째 재발 형태다: 결함의 *재현*은 있었는데
그 재현이 *회귀 고정*으로 스위트에 이식되지 않았다.

### 3.3 반박·기각 (근거와 함께)

- **Claude 적대 F8** (Hold와 경주하는 Release가 마커를 영구 고아화) — **반박**:
  호출자는 Hold가 반환한 뒤에야 핸들을 얻는데 `live=true`는 반환 전에 설정된다.
  외부에서 그 창에 도달할 수 없다.
- **Claude 적대 F2** (absent 단락이 재시작 루프와 경합) — spec이 고정한 오늘의
  동작이다("엔진이 없으면 기다리지 않는다" SHALL). 대안(죽은 엔진 대기·돌던
  서베이 개입)은 a101 계약 위반.
- **performance CRITICAL** (5분 백오프 = 보호 부재 창) — 새 결함이 아니라
  선언된 교환(design 배포 주의): 이전은 즉사 후 **무기한** 다운, 이후는 유한
  5m 후 성공 또는 오늘과 같은 fail-closed. 대안(복구와 병행 exit 루프)은
  미대사 상태에 강제 실행이라 더 위험.
- **maintainability의 중복 nil 가드 삭제 제안** — 기각: testing 전문가와 같은
  편에 선다. goroutine 문맥의 심층방어 가드는 유지하고 직접 테스트를 붙인다(§3.9c ⑨).
- 죽은 조건문·live/once 이중 장치·도달 불가 time.Now 폴백·소스 문자열 취약성 —
  전부 사실이나, 검증 완료된 동시성 코드의 재개봉 비용 > 정리 이득. 기록만.

### 3.4 선언된 생략 추가 (design 범위 밖 절에 등재)

429 재시도의 전체 재수집(cursor 재개는 execgw 계약 변경) · 대기 중 관측과
대시보드 Ready 표기 · 취소 중 부분 백오프 미계상 · pgrep 실행 무제한 대기
(a102 이전 표면) · spawn 게이트의 프로세스-로컬성(교차 프로세스 배타는
soakproc flock) · 콘솔 stderr 콜백의 미조인 goroutine(운영은 os.Stderr).

### 3.5 Fix-First 정산

- 수정 12항목(코드 4·테스트 5·문서 3) → T2 §3.9c 라운드로 발주. 결과는 §5에.
- 기각·반박 5건 — §3.3에 근거.
- 선언된 생략 6건 — §3.4, design.md에 등재.

## 4. 이 change가 드러낸 패턴 (다음 change를 위한 기록)

1. **단위 100% / 배선 0%.** T1·T2 모두 완벽히 테스트된 단위를 만들었고, 그
   단위를 운영 코드에 꿰는 배선은 어느 테스트도 닿지 않았다. A1·A2 모두
   정확히 그 틈을 생존 뮤테이션으로 뚫었다(A1 N1·N2, A2 N1~N5). 처방은 D5c:
   배선을 실행하는 테스트가 배선 주장의 유일한 증거다.
2. **재현은 고정이 아니다.** A2가 실측 재현한 경주 셋 중 둘은 회귀 테스트로
   이식되지 않아 gstack testing 패스가 재발견했다(§3.2). 재현 하네스는
   반드시 스위트로 옮겨야 증거가 된다.
3. **형태 검사는 배선의 모양만 약속한다.** 소스 문자열 테스트는 A2 F6·N5에서
   두 번 관통됐다. 실행 단언이 있는 곳에서만 보조로 쓴다.
4. **리뷰 층은 겹칠수록 새 것을 찾았다.** A2 2왕복이 끝난 코드에서 gstack
   7패스가 P1(PID 재사용) 하나와 미고정 불변식 셋을 더 찾았고, 그 P1은
   3모델이 독립 수렴했다. 단일 리뷰어의 MERGE-OK는 종결이 아니다.

## 5. §3.9c 결과

(T2 라운드 완료 후 기입)
