# a133 review

## Pre-Edit Gate (2026-10-13, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a133-an-alert-carries-no-account / 1.2
- 대상 심볼(패키지.함수): obs.(*Ntfy).Publish · obs.renderFields (+ 새 leaf obs.redactForTransport — internal/obs/a133_transport_redact.go)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/)
  — Publish 생산 호출 7(CodeGraph 6 + HEAD grep deliverOne) · renderFields 1(notificationFor) · notificationFor 3 · impact renderFields 5 ·
  affected --filter 1127
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md (R1~R6)
- 기존 동작 파악 근거: base dbb0681c HEAD 직접 읽기 · AST 번들 2(analysis/function-logic/) · 운영 outbox 실측(analysis/verify/outbox-census.md, 값 미열람) ·
  기존 시험 TestNtfyPublishesToTheTopic · TestNtfySendsTheBearerToken · TestNtfyReportsARefusal · TestNtfyWithoutATopicIsNotConfigured(obs_test.go) ·
  TestTheNotificationBodyCarriesNoToken(a074) · a124 전달기 판정 시험군 · a092 일반 등급 이관 시험군
- Function Logic Map / Branch Test Map: analysis/function-logic/ 2 번들(편집 전)
- upstream 상속 테스트 영향: no — internal/obs 는 TossOS 신설 패키지. 관련 패키지 전체 go test 로 회귀 확인
- 실패 테스트 선행 작성: yes — analysis/red/ (1.1)
- 설정·DB·journal 변경과 rollback: 없음 — 원장 스키마 · outbox 저장 내용 · 설정 · 기록 형식 무변경(design D4). 롤백 = 이미지 핀 되돌림
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문 · 청산 · 대사 · 체결 판정 무편집. Publish 의 분기 · 반환 불변(전달 판정 입력 불변),
  치환은 전송 goroutine 안의 순수 문자열 연산으로 알림기 뮤텍스 밖 — 손절 · 청산 경로가 기다리지 않음(design D5)
```

## proposal-freeze 적대 리뷰 (2026-10-13, 분리된 컨텍스트 — code-reviewer 서브에이전트, 저장소 무편집 · 네트워크 0 · 운영 데이터 미열람)

판정: **APPROVE-WITH-FIXES** — P0 0 · P1 1 · P2 5 · P3 7. 단일 경계(D2)와 저장 시점 비치환(D4)은 HEAD 대조로 성립 확인
(생산 전송 1 · 호출 7 · Tags 생산 0 · 재무장 덮어쓰기 · 재전송은 Title/Body 만 · payload 생산 읽기 0 · parkAlert 우회 · 엔진 등록이 buildGateway 보다 앞 ·
attempt id 접두 없는 hex). RED 시험이 못 잡는 변이 셋(F-2 · F-3 · F-4)을 지적.
절차 메모: 리뷰 중에 RED 시험 4 파일이 작업 트리에 있었고 생산 파일 편집은 0 이었다. GREEN 은 결과 수신 직전에 착수했고 처분은 아래 표 그대로 GREEN 커밋 전에 반영.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| F-1 | P1 | spec 이 등록 무관 「원문 · 숫자열 없음」을 단정 — 등록 전 · 계좌 전환 · 중첩 값은 깨짐 | **수용** — spec 문구를 「등록한 계좌의 원문 · 숫자열, 마스크 · 유도 id 는 언제나」로, 잔재 셋을 spec 에 명명(D4 ①) |
| F-2 | P2 | 구조 핀 ① 이 이름 모양만 봄(별칭 import · 타입 별칭 · obs 안 다른 http 전송 놓침) | **수용(부분)** — 별칭 import 해석 · `obs.Notification` 타입 별칭/정의 타입 금지 · ⑤ obs 안 `net/http` import 는 ntfy.go 하나. go/types `Implements` 는 x/tools 가 직접 의존이 아니고 저장소 전체 소스 타입 검사가 무거워 기각. spec 시나리오를 핀이 증명하는 범위로 좁힘 |
| F-3 | P2 | 새 `Notification` 문자열 칸이 경계를 우회해도 green | **수용** — `TestA133TheBoundaryCoversEveryTextField`(리플렉션: 문자열 칸 목록 고정 + 각 칸 센티널 제거) |
| F-4 | P2 | 좁힌 필드 버림(`account` 만) 변이가 살아남음 — `account_ref` 가 등록 값이라 치환이 가림 | **수용** — 계좌 키 필드 다섯(`account` · `account_ref` · `account_no` · `masked_account` · `account_seq`) 모두 등록 안 된 값, 본문 줄의 키를 `redact.IsAccountKey` 로 판정. 변이 M9 추가 |
| F-5 | P2 | httpapi SSE 스냅숏이 마스크 · `pos-` 를 밖으로 푸시 — 알림 census 밖이라 명시 필요 | **수용(명시)** — D1 에 범위 밖 사유. Manager 후속 후보로 완료 보고 |
| F-6 | P2 | D4 ① 은 읽기 전용 기동 등록으로 닫을 수 있음(원장 편집 불요) | **수용(기록)** — D4 에 「기각한 다른 안」으로. Manager 결정(잔재 + 운영 절차) 유지 |
| F-7 | P3 | 「전송 goroutine 안에서만」 부정확 — P1 은 Notify 호출자에서 돎 | **수용** — D5 를 a092 구조(RecordOnly · NormalRelay)가 근거라고 정정 |
| F-8 | P3 | 분기 불변 논증에 빈 문자열 · ASCII · 길이 사실 누락 | **수용** — D5 에 세 문장 |
| F-9 | P3 | BTM 에 시험 이름 없는 행 · renderFields 분기 번호 이동 | **수용** — 1.3 에서 이름 인용 · 옛/새 ast.json 정렬 |
| F-10 | P3 | (a) 비계좌 8자리+ 숫자 보존 (b) 호출자 Tags 불변 (c) obs 판정 시험에 모드 승격 없음 | **수용** — (a)(b) `TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags`. (c) obs 의 `newNotifier` 는 AccountRef 가 없어 승격 경로가 없음 — 승격은 엔진 시험 `TestA133TheDelivererStillLatchesOnRefusal` 이 `ModeEntryBlocked` 로 잼 |
| F-11 | P3 | `TOSSCTL_NTFY_TOPIC` 은 운영자 임의 문자열 | **수용** — D1 잔재 |
| F-12 | P3 | census 는 attestation 계좌 하나만 앎 | **수용** — outbox-census.md 에 모집단 한계 |
| F-13 | P3 | 청산 · 손절 비지연 시험이 D6 에 없음 | **수용** — D6 에 a092 · a124 · a098 시험 이름, 2.2 배터리에 포함 |

### Manager 결정 (2026-10-13)
- D4 ①(계좌 전환 뒤 옛 계좌 PENDING 행): **잔재로 수용**, 코드로 막지 않음(YAGNI). 운영 절차 「계좌 자격 증명을 바꾸기 전에 alert_outbox PENDING 0 확인(아니면 먼저 전달 · 정리)」 — design 배포 절 · tasks 게이트 뒤 절차 4.
