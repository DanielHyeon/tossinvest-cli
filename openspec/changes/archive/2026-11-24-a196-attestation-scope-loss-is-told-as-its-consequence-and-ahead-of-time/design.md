# a196 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 감독 증거의 기한 = `ProofAt + (ExpiresAt − IssuedAt)`. `ProofAt` 은 기존 `binding` 의 값 — attestation 파일에 **실린** 증거의 endpoint 별 최신 중 가장 이른 것. `ProofAt` 이 0(증거 없음)이거나 `ExpiresAt ≤ IssuedAt` 이면 감독 증거의 기한은 없다 — 만료만 센다. **Manager 결정 2026-11-23**(Q1): 기한은 「실린 증거」 기준이다. 발급자는 endpoint 마다 입력 순서상 첫 유효 증거 하나만 싣으므로(acceptSupervised `seen`@335) 검증 기록에 다른 시장의 더 새 증거가 있으면 실제 범위 상실은 더 늦을 수 있다 — 경보가 이른 쪽이지 놓치는 쪽이 아니므로 받아들이고 발급자는 고치지 않는다(§6). 기한 정각은 이미 「지난 것」이다(`age >= validity` 제외, acceptSupervised B6 if@332) | 파일 안의 값만 쓴다. 엔진이 `internal/soak` 를 import 하지 않고, 30 일이라는 숫자를 복사하지도 않는다 |
| D2 | 「수용 기간 = attestation 유효 기간」이라는 전제는 시험이 고정한다: 발급 함수(`soak.BuildAttestation` 계열)에 나이 V−1ns · 정확히 V · V+1ns(V = 유효 기간)의 감독 증거를 넣어 첫째만 실리고 둘째·셋째는 빠짐을 보고, 실린 파일로 계산한 기한이 「그 증거가 빠지기 시작하는 시각」과 같음을 단언한다. **Manager 결정 2026-11-23**(Q5): 자리는 `internal/soak` 외부 시험(`soak_test` — 기존 픽스처 `threeCleanDays`·`criteria`·`supervisedProofs(at)`) + engine 쪽 계산 시험(같은 경계값). 두 패키지 사이 import 순환은 어느 방향에도 없다(tasks 1.4 실측) | 「우연이 지키는 안전」을 못 박는다 — 발급자가 두 값을 갈라 쓰기 시작하면 이 시험이 먼저 깨진다 |
| D3 | 날수의 기준 = `min(만료, 감독 증거의 기한)` 까지 남은 시간의 하루 단위 올림(기존 `daysLeftCeil` — **양수에만 부른다**, 함수 무편집). D-7~D-3 일반, D-2·D-1 critical. **Manager 결정 2026-11-23**(Q4): 남은 시간이 0 이하(기한은 지났는데 파일은 아직 판정 통과)면 날수 단계가 아니라 새 단계 `proof_overdue`(critical)다 — 신원·key 는 거절 단계와 같은 방식(단계 + 기한 시각 + KST 날짜)으로 하루 한 번. `attestationVerdict.critical` 은 healthy·notice 만 false 이고 나머지는 default(AST B3 case@78 → return@79 true)라 새 단계를 **편집 없이** critical 로 센다 | 갱신이 성공하면 그 구간은 최대 6 시간이고 다음 갱신이 범위 상실로 바꾼다. AST: 갱신이 실패하면 파일을 쓰지 않으므로(runSoakAttest — BuildAttestation 오류에서 return, `attest.Save` 미도달) 그 구간은 그 파일의 만료까지 이어질 수 있다(`Verify` 는 감독 증거 나이를 보지 않음) — 1.6 리뷰 정정. 그 구간의 처리는 Manager 결정 Q4 로 `proof_overdue`(위) |
| D4 | 날수 단계의 신원 = **기한의 시각**(파일의 발급·만료 시각이 아님). 만료가 기한이면 종전과 같은 동작이다(갱신이 만료를 바꾸므로 새로 센다). 감독 증거가 기한이면 갱신을 넘겨 같은 단계가 유지된다 | 감독 증거의 기한은 실린 증거가 같은 한 갱신으로 바뀌지 않는다. 실린 증거가 바뀌면(다른 시장의 증거로 대체, 새 감독 검증) 기한이 바뀌고 새로 센다(**Manager 결정 2026-11-23**, Q1) — 파일 신원으로 세면 6 시간마다 같은 통지가 다시 나간다 |
| D5 | 거절 단계의 신원 = 단계 + (범위 상실이면 정렬된 누락 endpoint 집합) + KST 날짜. 발급 신원을 뺀다. outbox key 도 같은 신원으로 만든다. `proof_overdue` 는 단계 + 기한 시각 + KST 날짜(**Manager 결정 2026-11-23**, Q4). **에피소드(Manager 결정 2026-11-23, gstack review G12)**: 거절 단계와 `proof_overdue` 의 신원·key 에 거절 에피소드의 시작 시각을 더한다 — 직전 판정의 거절 신원(단계 + 누락 집합 또는 기한)과 다르면(직전이 거절이 아니었던 경우·첫 판정 포함) 이번 판정 시각에 새 에피소드. 같은 에피소드의 재발급은 같은 key(하루 한 번), 날짜가 바뀌면 새 key, 회복 뒤 재발·A→B→A·재시작은 새 에피소드 = 새 key. 이유: 에피소드가 없으면 회복 뒤 같은 날 재발이 옛 outbox 행의 재알림 창(전달·확인 뒤 1시간) 안에서 정산된 행으로 삼켜지고 감시가 「보냄」으로 적어 그날 침묵함(journal/outbox.go claimOwed) 「보냄」은 프로세스 메모리뿐이라 재시작 뒤 현재 단계를 한 번 더 알릴 수 있다 — a135 D2 의 허용을 정본에 적는다(**Manager 결정 2026-11-23**, Q6) | 정본의 의도 「하루 한 번」. 2026-11-18~23 의 21 건이 5~6 건이 된다 |
| D6 | 타입은 `broker.credential_expiring` 그대로. 제목·본문만 고친다 | 새 타입은 `internal/obs` 열거·등급표·정본 문구·a102/a134 유일성 시험을 건드린다(조사 인용). 운영자가 먼저 읽는 것은 제목이다 |
| D7 | 문장(확정 문구는 구현에서, 뜻은 고정). **Manager 결정 2026-11-23**(Q3): 재시작이 거절될 상태의 결과 문장은 **제목과 본문의 첫 문장 둘 다**에 둔다(기기의 제목 표시는 저장소로 검증 불가 — 제목은 `Title` 헤더 UTF-8 원바이트). 범위 상실 제목 「지금 엔진을 재시작하면 기동이 거부된다 — 주문·취소 검증이 attestation 에 없다」. 본문: 첫 문장에 같은 결과 / 빠진 endpoint 이름 / 재시작이 거부되면 자동 손절·익절이 돌지 않는다 / 지금 도는 엔진은 계속 돈다 / 할 일(**Manager 결정 2026-11-23, gstack review G10** — 감독 검증은 엔진과 같은 실행 잠금(`cmd/tossctl/verify.go:440-449` `enginelock.Acquire`)을 잡으므로): 「감독하 실주문 검증을 다시 해야 한다. 이 검증은 엔진과 같은 잠금을 써서 엔진을 멈춰야 실행되고, 멈춘 동안에는 자동 손절이 돌지 않는다. 보유 종목의 시장이 닫힌 시간에 엔진을 멈추고, 터미널에서 `tossctl verify run`(주문을 받는 시장을 `--market` 으로) → `tossctl soak attest` → 엔진 재시작 순서로 하라」 — 범위 상실 POST 갈래·감독 증거 날수 통지(「기한 전에」)·`proof_overdue` 공통. `--resume`/`--redo` 는 문장에 쓰지 않음. 날수 통지: 「감독하 실주문 검증이 N 일 뒤(<시각>) 만료된다. 다른 유효한 감독 증거가 없으면 그 뒤 첫 갱신부터 엔진 재시작이 거부된다」 + 같은 할 일(Manager 결정 2026-11-23 R2-2 — Q1 과 정합). `proof_overdue`(Q4): 「감독하 실주문 검증의 기한이 지났다(<시각>). 다른 유효한 감독 증거가 없으면 그 뒤 첫 갱신부터 엔진 재시작이 거부된다 — 갱신이 실패하고 있으면 이 파일의 만료(<시각>)에 거부된다. 지금 감독 검증을 다시 하라」 | 2026-11-23 사고에서 빠져 있던 세 가지(결과의 단정·원인·할 일) |
| D8 | 계좌 정보·파일 경로·원문 오류 금지, 알림 꺼진 엔진의 일반 등급 강등, 판정 결함 회수·알림기 패닉 비회수 — 전부 기존 규칙 그대로 | 정본 그대로 |

## 2. 편집 자리

- `judgeAttestation`: Verify 통과 뒤 `DaysLeft` 계산의 기준 시각을 새 함수(`restartDeadline(verdict) (time.Time, kind)`)의 값으로 바꾸고, verdict 에 기한의 종류·시각을 싣는다. 남은 시간 ≤ 0 이면 `proof_overdue`(`daysLeftCeil` 을 부르기 전에 가름).
- `stageKey` / `prepare`: 신원 계산(D4·D5), `proof_overdue` 의 키.
- `eventFor`: 제목·본문 첫 문장·key, `proof_overdue` 문구(switch 에 case 하나).
- 새 단계 상수(`stageProofOverdue`)는 const 블록 — 함수 본문 아님.
- 무편집 확인(1단계 AST): `attestationVerdict.critical`(default 가 새 단계를 흡수), `daysLeftCeil`(양수에만 호출), `checkOnce`, `binding`, 헬퍼 `a135Attestation`.
- 편집하는 기존 함수 = production 4(`judgeAttestation`·`attestationWatch.prepare`·`stageKey`·`attestationWatch.eventFor`) + 시험 2(`TestA135StageBoundaries`·`TestA135SameStageOnANewAttestationIsAnnouncedAgain`) — 번들 6(`analysis/function-logic/`). 새 함수(기한 계산·문장 조립)는 같은 파일의 새 함수 또는 새 파일.

## 3. 침묵·거절하는 정상 입력

- 감독 증거가 실리지 않은 attestation(게이트가 조회만 요구하는 구성이 있다면): 감독 증거의 기한 없음 → 종전 동작.
- 게이트가 꺼진 엔진: attestation 을 읽지 않는다(정본 그대로).

## 4. 운영 절차 (코드 밖, 사람 승인)

1. 엔진 생존 감시 타이머 설치: `deploy/systemd/tossos-heartbeat.{service,timer}` 를 사용자 systemd 에 설치·활성(`systemctl --user enable --now tossos-heartbeat.timer`). 엔진이 죽으면 엔진과 무관하게 1 분 안에 알림이 간다. 2026-11-23 미설치 확인.
2. 감독하 실주문 검증의 재실행과 엔진 재기동(2026-11-23 사고 복구) — 이 change 와 독립.

## 5. 시험 전략 (RED 먼저)

새 시험 파일 `internal/app/engine/a196_attestation_deadline_internal_test.go`(a135 시험의 하네스·가짜 시계 재사용):

1. 만료 30 일 뒤·감독 증거 기한 7 일 뒤 → 일반 등급 1, 본문에 감독 증거 기한과 `verify run`. 8 일 뒤(7 일 1 초) → 0.
2. 같은 날수 단계 안에서 같은 감독 증거를 실은 새 발급 파일로 바뀌어도 추가 알림 0. 날수가 하나 줄면(D-6) 1(r2 R2-1 — 억제 단위는 KST 날짜가 아니라 「같은 기한의 같은 날수 단계」).
3. 감독 증거 기한 2 일·1 일 → critical 각 1. 기한이 지났고 파일은 아직 통과 → `proof_overdue` critical 1, 같은 날 재판정·같은 증거 재발급 → 0, 다음 날 → 1(이틀 연속 날마다 1 — Q4). 기한 정각(남은 0) → `proof_overdue`.
4. 새 감독 증거로 기한이 30 일 뒤 → 0, 그 뒤 새 기한으로 다시 셈.
5. 감독 증거 없음 → 만료만 센다(a135 시험과 같은 결과).
6. 범위 상실: 첫 판정 critical 1, 같은 날 같은 누락으로 재발급 2 회 → 0, 다음 날 → 1. 누락 집합이 바뀌면(하나 더 빠짐) 같은 날 1.
7. 범위 상실 제목**과 본문의 첫 문장**에 「재시작」「거부」가 있고 조건 접속(「…하면 …할 수 있다」류) 없이 단정이며, 본문에 빠진 endpoint·손절·`tossctl verify run`(Q3).
8. 제목·본문·필드·key 에 계좌 원문·숫자열·마스크 꼴 0(기존 누출 시험 형식).
9. 알림 꺼진 엔진 → 일반 등급 기록만.
10. D2 의 전제 고정 시험: `internal/soak` 외부 시험에서 나이 V−1ns 실림 · 정확히 V 빠짐 · V+1ns 빠짐, 실린 파일의 기한 식 = 빠지기 시작하는 시각. engine 쪽은 같은 경계값으로 기한 계산을 단언.
11. 기존 a135 시험: 두 시험 함수만 번들과 함께 고친다(헬퍼 `a135Attestation` 무편집 — Q5). `TestA135StageBoundaries` 는 표를 「기한」 기준으로(행 설명에 「기한 = 만료 − 1h」 주석 — 픽스처 증거가 발급 1시간 전), `TestA135SameStageOnANewAttestationIsAnnouncedAgain` 은 새 규칙으로 다시 쓴다(같은 감독 증거의 재발급 → 다시 안 알림, 기한이 바뀐 새 attestation → 다시 알림). 나머지 a135 시험은 그대로 통과해야 한다.
12. 「통과 이유가 바뀐 시험」 — 새 시험이 그 사실을 직접 단언한다:
    - `TestA135StageBoundaries`「1초 = D-1」(:82): 기한(만료 − 1h)은 이미 지남 — 종전 계산은 `daysLeftCeil(−59m59s)=1` 로 우연 통과, a196 에서는 `proof_overdue` 로 바뀜(표 수정 대상).
    - `TestA135SameStageOnANewAttestationIsAnnouncedAgain` 후반(:491-493): 발급만 +1분 → D1 식의 유효 기간이 1분 줄어 기한이 바뀌어 통과 — 발급 신원 재설정의 증명이 아님(재작성 대상).
    - `TestA135RenewalResetsAndScopeLossIsImmediate` 정상 갱신(:262-268): 새 픽스처가 감독 증거까지 바꿔 기한이 바뀜 — 「같은 증거를 유지한 갱신은 침묵」을 재지 않음(무편집, §5-2 가 그 사실을 단언).
13. 보강(Manager 결정 2026-11-23 R2-3): (i) outbox Key 문자열을 직접 비교한다 — 가짜 notifier 는 중복 억제를 하지 않으므로 건수가 아니라 Key(재시작을 흉내 낸 새 감시가 같은 거절 상태에서 같은 Key 를 내는가, 누락 집합이 다르면 다른 Key 인가). (ii) KST 날짜만 바뀌는 시각 쌍(같은 UTC 날짜, KST 23:59 → 00:01). (iii) scope_lost → expired 전이(같은 날 단계가 바뀌면 두 알림이 각각, Key 가 다름). (iv) endpoint 마다 시각이 다른 감독 증거(가장 이른 것이 목록의 첫 원소가 아니게 — binding 을 우회한 기한 계산을 잡는다).

Scenario 이름 셋(「만료 8일 전」·「만료 7일 전부터 3일 전까지」·「만료 2일 전과 1일 전」)은 정본 이름을 유지 — MODIFIED 는 기존 Scenario 이름을 뺄 수 없다(strict validate). 내용은 기한 기준(Manager 승인 2026-11-23).

변이(전부 CAUGHT, 12 — Manager 결정 2026-11-23 Q5 로 6 → 12): ① 기한을 만료로만 계산 · ② `min` 을 `max` 로 · ③ 날수 신원에 발급 시각 되돌리기 · ④ 거절 key 에 발급 신원 되돌리기 · ⑤ 누락 집합을 key 에서 빼기 · ⑥ D-2 경계 한 칸 밀기 · ⑦ 수용 경계 `age >= validity` → `>`(soak — D2 전제 시험이 죽여야 함) · ⑧ 기한을 「가장 늦은 증거」로 계산 · ⑨ 0 이하 잔여를 imminent 로 접기 · ⑩ KST 날짜를 UTC 로 · ⑪ 거절 단계 key 에서 KST 날짜 삭제(다음 날 재알림이 죽어야 함 — R2-3 으로 재정의) · ⑫ 누락 집합 정렬 삭제(`MissingEndpoints` 가 이미 정렬하면 동등 변이 — 실제로 돌려 생존을 확인한 뒤 원장에 사유와 함께).

선택 시험: `./internal/app/engine -run 'A196|A135|Attestation'`, D2 시험이 놓인 패키지의 해당 시험.

## 6. 잔여

- 타입 태그 `broker.credential_expiring` 은 남는다(D6). 운영자가 태그로 걸러 읽는다면 여전히 「자격 만료」로 보인다.
- 엔진이 꺼져 있으면 이 감시도 돌지 않는다 — §4-1 의 타이머가 그 구간을 맡는다.
- 두 시장 기록이 함께 있으면 기한이 실제 상실보다 이르게 나올 수 있다(발급자가 endpoint 마다 입력 순서상 첫 유효 증거 하나만 싣는다 — 순서는 `sort.Slice` 구현에 기댄다). 발급자가 endpoint 별 최신 증거를 싣게 하는 것은 후속 후보(a063 영역)(Manager 결정 2026-11-23 Q1).
- 제목 표시 미실측 — ntfy 서버·기기가 한글 `Title` 헤더를 어떻게 보여 주는지 저장소로 검증 불가(ntfy.go:112 주석은 「ASCII-fold」). 그래서 결과 문장을 본문 첫 문장에도 둔다(Q3). 실제 표시는 Manager 가 사용자에게 확인.
- 기본값보다 긴 `--validity` 로 손수 발급한 파일은 감시의 기한이 기본값 갱신 타이머가 증거를 빼는 시각보다 늦다 — 그 경우만 늦은 경보(scope_lost 사후 알림) 쪽(gstack review 적대 F2, 2026-11-23; 주석에 기록, 처리는 Manager 결정 대기).
- 재검증의 정확한 플래그 조합(기존 검증 기록 위에서 주문·취소 단계를 다시 재는 방법)은 미확인 — 운영 절차 문서의 몫(G10).
- 후속 후보(범위 밖, gstack review): 알림에 실린 증거의 시장 이름(G5), `IssuedAt` 이 0 인 파일의 명시 가드(G7), 콘솔 attestation 패널의 기한 표시(G11 — 지금은 만료만).
- 재시작 뒤 현재 단계 1회 재알림(Q6) — 거절 단계는 재시작이 새 에피소드라 새 key 로 한 번 더(G12) — 알림 수 예상(하루 1건)은 재시작이 없을 때의 값.
