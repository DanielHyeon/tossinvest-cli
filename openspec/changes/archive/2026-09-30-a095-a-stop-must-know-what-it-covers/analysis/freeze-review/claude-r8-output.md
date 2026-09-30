# a095 8판 — 좁힌 재검증 Claude 독립 보이스 결과 (2026-09-29)

프롬프트 `claude-r8-prompt.md` (sha256 `59797248…ae2da7`, 실행 사본 일치). 트리 HEAD `103f1ded` export + a095 오버레이(diff 0),
스크래치패드 `r8-tree`(끝까지 존재). 약 40 도구 호출. `analysis/freeze-review/` 미열람. 판정: **VERDICT: REJECT**.

## 1~5 항목

| # | 항목 | 결과 | 근거 |
|---|---|---|---|
| 1 | r7b F1 | RESOLVED(N3 · N6 주의) | es-spec:58-62 「거부되지 않은」, 거부 등급 Q2(a) 잔류; 시나리오 :94 |
| 1 | r7b F2 + 저자 추가 | RESOLVED | es-spec:98, :102 두 전제 |
| 1 | r7a F1 / r7b F5(기록 + 순서 조건) | PARTIAL | proposal:205 · design:204-209 · tasks:165-171 문언은 같으나 tasks:179/186-193과 모순(N1) · 7.0 위치(N2) |
| 1 | r7b F4 | RESOLVED | es-spec:49-50 ↔ 시나리오 :109-111 |
| 1 | r7b F6 / r7a F5 | RESOLVED | exit-policy:28 · es-spec:119 · tasks:150-151 |
| 1 | r7a F2 | RESOLVED | es-spec:34-36 · design:107-108 |
| 1 | 거부 알림 블록 = 꺼짐 | RESOLVED | es-spec:44-45, :121-123; notifications.go:105-107 |
| 1 | r7b F8 | RESOLVED | design:102-105 · tasks:92-93 · 델타 머리말 8판 |
| 1 | 오류 계약 전파 | PARTIAL | 생성 맵 21개는 표본 전부 맞음; 손으로 쓴 evaluateladder 맵이 틀림(N5) |
| 2 | Q2(a)/(b)/(c)를 아직 정하는가 | 알림 켜진 엔진에서는 정하지 않음 | 알림 off는 결정 (2)에 따라 전부 non-critical(의도). `mergeAdoption` B3 → `judgeHoldings` B11 건너뜀 → B12 → `alertUnmanaged` B3 사유 — 정의가 거부를 옳게 뺀다. 주의 N3 · N6 · N7 · N8 |
| 3 | 공동 충족 | 충족 가능 | 로드 값 판정 · 거부 = 꺼짐 모두 `Enabled` 거짓(notifications.go:106)으로 성립. `resolution.Refused`는 거부와 topic 없음 둘 다에 세팅 — `Refused`를 꺼짐으로 읽는 구현은 거부 시나리오는 통과하나 es:97-99 · tasks 2.5a에서 실패해 잡힌다. 재시도 SHALL ↔ 시나리오 정합 |
| 4 | 오류 계약 영수증 | 21개 중 7개 표본 전부 코드와 일치 | judgeHoldings · 두 alertUnmanaged · adopt · checkExternalIncrease · resolveNotificationPublisher · Notify. RETHROWS 7개는 전부 `error`를 돌려주나 evaluateladder 문구는 거짓(N5) |
| 5 | 순서 조건 문구 | 네 자리에서 일관, 다른 자리와 모순 | N1 · N2 |

## 새 발견

| id | 등급 | 내용 | 근거 | 제안 |
|---|---|---|---|---|
| N1 | **P1** | 새 순서 조건(「a095 구현은 a092 『모든 보유자』 착지 이후 또는 같은 창, 깨지면 P1」)이 같은 change의 선후 관계 절과 모순 — 그 절은 a095가 a092와 **독립**이고 남은 것은 「순서 의존이 아니다」라고 적으며, 2판의 「a095는 a092 뒤」가 사용자 결정 (1) 「묶지 않는다」와 모순이었다고 적는다. 8판은 조건부 「a092 뒤」를 기록된 사용자 결정 없이 되살렸고, 전제가 깨질 때 무엇을 하는지 말하지 않는다(「P1이다」는 행동이 아니라 딱지). 스케줄러가 선후 관계 절을 읽으면 「독립」을 본다 | tasks.md:179, 186-193 vs tasks.md:165-167, proposal.md:205, design.md:207-209; 결정 (1) proposal.md:33 | 선후 관계 절을 고쳐 조건과 그것이 범위 묶기가 아니라 스케줄링 전제인 이유를 적고, 사용자 확인을 받거나 결정 (1)을 좁히는 권한을 기록; 전제가 깨질 때의 행동(구현 착수 안 함 / 7절 보류 등)을 명시 |
| N2 | P2 | 7.0은 「구현 착수 전」 확인인데 「7. 배포와 운영」에서 7.1 · 7.2 뒤에 있음 | tasks.md:159, 165 | 「0. 게이트 선행」으로 옮김 |
| N3 | P2 | 「거부된 엔진은 보호를 요청한 엔진」 근거가 한 모양에서 거짓 — `validate()`는 enabled 거짓 · include 없음이어도 `DefaultStopPct≠0`이면 돌아, 의도적으로 끈 블록에 범위 밖 pct가 남은 경우도 거부된다 | engine.go:157-163; design.md:154-155; mergeadoption 맵 결론 | 두 자리의 문장 정정, off-with-pct 거부 모양을 Q2(a)에 기록 |
| N4 | P2 | tasks 2.6a 「같은 경로로 normal이 되는 구현은 실패해야 한다」 — 경로는 관측 불가, Q2(a)=normal을 배제하는 것으로 읽힘 | tasks.md:79-81 | 관측 가능한 것만 단언: 사실 칸 · key가 「설정 거부」이고 「편입 꺼짐∧미지정」이 아니다, 등급 단언 없음 |
| N5 | P2 | 오류 계약 정정이 손으로 쓴 번들 3개에 닿지 않음 — evaluateladder 맵은 「브로커 · 원장 오류를 되던진다」인데 `EvaluateLadder`는 브로커 · 원장 호출이 없음(import fmt · math/big · strconv); 생성기 `ERROR_CONTRACTS` 항목은 BUNDLES에 없어 쓰이지 않음. review §3.21 영수증은 7개 전부 맞다고 적음 | evaluateladder function-logic-map.md:63; ladder.go:75-79, :307; render_bundles.py:51, 54-55 | 그 맵을 손으로 고치고 영수증에 손으로 쓴 번들을 적음 |
| N6 | P3 | 「이 둘의 보고는 … 이전과 같은 등급」(SHALL)이 거부 문장 바로 뒤에 있어 거부 경우까지 덮는 것으로 읽힘 | es-spec:60-61 | 거부 문장을 SHALL 뒤로 |
| N7 | P3 | 시나리오 「앞선 normal 보고 뒤의 시도 실패」는 연기 보고가 normal임을 전제 — Q2(c)=normal일 때만 성립 | es-spec:101-103; tasks:92 | 「(Q2(c)=normal일 때)」 |
| N8 | P3 | enabled 참 + include 지정 종목은 B5(:409가 :412보다 앞)로 가 critical — Q2(b)(B6)와 정합하나 일곱 칸 중 어느 칸인지 말하지 않음 | adoption.go:409-414; design.md:103 | enabled + include는 「편입 켜짐 — 시도 실패」라고 명시 |
| N9 | P3 | config 번들 두 맵의 커버리지 문구가 네 패키지를 이름 댐, 실제는 `coverage/r8-config.out` | mergeadoption 맵 「표의 유래」 | config 프로파일 명시 |

> **VERDICT: REJECT** — N1 (P1): 8판의 순서 조건이 같은 change의 선후 관계 절과 모순되고, 기록된 조정이나 전제가 깨질 때의
> 행동 없이 사용자 결정 (1)에 조건을 붙인다. 좁힌 범위의 나머지는 해소됐거나 P2/P3 기록만 필요하다.
