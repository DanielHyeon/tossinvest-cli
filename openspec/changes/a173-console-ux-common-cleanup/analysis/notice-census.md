# 알림 상자 전수 분류표 (a173 3.2)

기준 b27a1ea87 의 `class="notice|danger"` 자리 전부(옛 줄 번호). 새 class 는 같은 문구의 현재 HEAD 줄에서 읽은 근사 매칭(문구가 바뀐 자리는 diff 참조). 규칙: `.danger`=차단(읽기 실패·거부), `.notice`=결과가 따르는 사실(승인 대기·구버전 바이너리·저장 결과·비가역), `.aside`=상자 없는 본문(반영 시점은 `muted effect-timing`), 행 단위 표시는 `status-pill warn`/`submetric bad|warn`. 별도로 `causes` 배너(`danger cause-banner`)가 화면당 최대 1개 새로 생김.

| 파일:옛줄 | 옛 | 새 | 문구 |
|---|---|---|---|
| templates.go:520 | notice | notice console-approval | 승인을 기다리는 검증 run이 있다. |
| templates.go:539 | notice | aside | 지금은 KR 정규장(평일 09:00–15:30 KST) 밖이다. |
| templates.go:554 | danger | (상자 class 없음) |  |
| templates.go:572 | notice | aside | 같은 바이너리를 같은 포트로 다시 실행한다. 초 뒤 이 페이지가 |
| templates.go:575 | notice | (상자 class 없음) |  |
| templates.go:590 | notice | notice | 이 콘솔은 설치된 바이너리보다 오래되었다. |
| templates.go:594 | notice | notice | 실행 중인 soak은 설치된 바이너리보다 오래되었다. |
| templates.go:605 | notice | (상자 class 없음) |  |
| templates.go:646 | notice | notice | 실행 중인 엔진은 설치된 바이너리보다 오래되었다. |
| templates.go:670 | notice | aside | 마지막 응답 (): |
| templates.go:686 | danger | (상자 class 없음) |  |
| templates.go:724 | notice | aside | 게이트가 열리지 않는 이유: |
| templates.go:727 | notice | aside | attestation 만료까지  남았다. 72시간 이내다. |
| templates.go:728 | notice | aside | renewal 진단:  · 마지막 시도  ( 전) |
| templates.go:737 | danger | (상자 class 없음) |  |
| templates.go:750 | notice | notice | () 단계는 새 프로세스를 기다린다 — 위의 [콘솔 재시작] 뒤 이어하기. |
| templates.go:753 | danger | danger | 계좌에 아직 살아 있는 객체: |
| templates.go:769 | notice | (상자 class 없음) |  |
| templates.go:773 | notice | notice | 이 프로세스는 이미 검증을 수행했다. 조건주문 존속은 등록한 프로세스가 |
| templates.go:792 | notice | notice | 재측정 단계 — . 판정이 pass인 단계는 |
| templates.go:797 | danger | danger | 여기서부터 실제 계좌에 요청이 나간다. 아래 목록이 이 실행이 보낼 수 있는 |
| templates.go:814 | notice | (상자 class 없음) |  |
| templates.go:824 | danger | (상자 class 없음) |  |
| templates.go:831 | notice | notice | 중단: |
| templates.go:833 | danger | danger | 계좌에 아직 살아 있는 객체: |
| templates.go:837 | notice | notice | 여기서 프로세스 경계다. 조건주문이 이 프로세스보다 오래 사는지가 |
| templates.go:858 | notice | notice | 이 기록에는 이미 검증이 있다. 시작하면 이어서 진행한다 — |
| templates.go:862 | notice | notice | () 단계가 새 프로세스를 기다리고 있었다. 이 콘솔이 그 새 프로세스다. |
| templates.go:865 | notice | notice | 이전 실행이 남긴 객체 건을 이 실행이 먼저 취소한다 — |
| templates.go:872 | notice | aside | 이어할 단계가 없다 — 모든 단계에 판정이 있어 이어하기는 아무것도 |
| templates.go:897 | notice | aside | deferred·refused 단계는 대상이 아니다. pass도 |
| templates.go:926 | danger | (상자 class 없음) |  |
| templates_exit_quarantine.go:21 | danger | danger | 해제는 수리가 아니라 재시도다. |
| templates_optimization.go:20 | notice | (상자 class 없음) |  |
| templates_optimization.go:21 | notice | (상자 class 없음) |  |
| templates_optimization.go:22 | danger | danger | lifecycle을 읽지 못했다: . 마지막 값을 현재값처럼 사용하지 않고 모든 변경을 닫았다. |
| templates_optimization.go:23 | notice | aside evidence-banner | 성과 근거  · 근거 기반 추천 candidate를 만들지 않는다. |
| templates_optimization.go:74 | danger | danger | legacy effective config를 읽지 못했다: |
| templates_optimization.go:75 | notice | aside | lifecycle command seam 미배선 · 조회만 가능하다. |
| templates_optimization.go:80 | danger | danger configuration-error | 설정 오류 · 읽기 전용 owner descriptor를 바로잡기 전에는 preset preview와 app |
| templates_optimization.go:128 | danger | danger | 보호 상태를 읽지 못했다: |
| templates_optimization.go:186 | notice | aside | 미승인 · passed 구조적 0 · verdict 비활성 |
| templates_optimization.go:205 | notice | aside | lane desired OFF · engine autostart OFF · scheduler desired  |
| templates_optimization.go:217 | notice | aside | 최근 30일 · 전체 시장 · 전체 lane · complete lineage · 읽기 전용 |
| templates_optimization.go:237 | notice | notice | 아직 적용되지 않았다. before/after, 적용 시점과 권한 불변을 확인한다. |
| templates_optimization.go:257 | danger | danger | 보호 또는 위험 계약에 영향을 줄 수 있어 3초 확인 대기가 적용된다. 문구나 사유를 직접 입력하지 않는다. |
| templates_optimization.go:258 | notice | notice | 초 남음승인 가능 |
| templates_optimization.go:276 | danger | danger | base v 승인 중 latest v을 확인했다. stale draft를 자동 retry하지 않았다. |
| templates_optimization.go:315 | notice | aside | performance.db 조회 seam 미배선 — 성과를 0으로 꾸미지 않는다. |
| templates_optimization.go:317 | danger | danger | 성과를 읽지 못했다. 값은 변경되지 않았다: |
| templates_optimization.go:334 | notice | aside | insufficient_sample · 추천 근거로 사용 불가 |
| templates_optimization.go:361 | notice | aside | 고정 조회 범위에 complete lineage 거래가 없다. 빈 결과는 성과 0이 아닙니다. |
| templates_optimization.go:367 | notice | aside | Attribution generation을 읽지 못했다. aggregate를 0으로 바꾸지 않는다: |
| templates_optimization.go:383 | notice | aside | 저장된 attribution generation에 표시할 row가 없다. 이는 손익 0이 아니다. |
| templates_optimization.go:393 | danger | danger | 브로커 보호가 약해질 수 있다. preview 뒤 최소 3초 동안 영향 범위를 확인한다. |
| templates_optimization.go:411 | danger | danger | 상태를 읽지 못해 닫힌 기본값을 표시한다. 신규 진입은 fail-closed다. |
| templates_optimization.go:412 | notice | aside | scheduler seam 미배선 — 닫힌 기본값만 표시한다. |
| templates_optimization.go:435 | danger | danger | runtime projection을 읽지 못했다. 두 시장 모두 신규 진입 OFF와 typed unknown |
| templates_optimization.go:436 | notice | aside | runtime endpoint 미기동 — KR과 US의 dormant OFF/UNOBSERVED/UNWIRE |
| templates_optimization.go:476 | danger | danger | 선택KR / KRWUS / USD |
| templates_optimization.go:478 | danger | (상자 class 없음) |  |
| templates_orders.go:78 | notice | notice | 페이지가 잘렸다. 브로커가 다음 페이지가 있다고 답했으므로 |
| templates_orders.go:91 | notice | aside | 주문 조회가 배선되지 않았다. 이 빌드의 콘솔에 주문 조회 seam이 |
| templates_orders.go:94 | notice | aside | 검증 중 — 데이터 없음. . |
| templates_orders.go:97 | notice | aside | 검증 중 — 갱신 보류. . |
| templates_orders.go:105 | notice | aside | 아직 읽지 않았다. |
| templates_orders.go:128 | notice | aside | 필터를 쓸 수 없다. 목록을 가져오지 못했으므로 걸러 낼 것이 없다. |
| templates_orders.go:185 | notice | aside | 발주 주체는 전부 "불명"이다. 원장을 읽지 못했으므로 어느 주문이 |
| templates_overview.go:86 | notice | aside | 엔진 원장에 포지션이 없는 보유가 있다 — 엔진이 진입한 포지션이 아니므로 손절·익절 라인도 없다. |
| templates_overview.go:107 | notice | aside | 엔진 상태 배선이 없다 (활성 마커 경로 미지정) — 정지로 읽지 않는다. |
| templates_overview.go:110 | notice | aside | 기동 거부 사유를 포함한 마지막 응답 (): |
| templates_overview.go:153 | danger | danger | 마지막 브로커 조회 실패: |
| templates_overview.go:156 | notice | aside | 실계좌 검증이 실행 중이다 (). 이 화면은 원래 |
| templates_overview.go:234 | danger | danger | 계좌에 아직 살아 있는 객체: |
| templates_overview.go:239 | notice | aside | 잔여물은 다음 검증의 노출 상한을 먼저 채운다 — 그것이 남아 있는 동안 |
| templates_overview.go:258 | danger | danger | 한도를 읽지 못했다: |
| templates_overview.go:268 | notice | (상자 class 없음) |  |
| templates_overview.go:290 | notice | (상자 class 없음) |  |
| templates_portfolio.go:23 | notice | aside | 원장 경로가 배선되지 않았다. 이 빌드의 콘솔은 엔진 journal을 읽도록 |
| templates_portfolio.go:26 | notice | aside | 엔진 미가동 — 원장 파일이 아직 없다. |
| templates_portfolio.go:30 | danger | danger | 콘솔 업데이트 필요. 원장 스키마 v는 이 콘솔이 이해하는 |
| templates_portfolio.go:34 | danger | danger | 엔진 기동 필요 — 마이그레이션 전이다. 원장 스키마 v에는 이 화면이 |
| templates_portfolio.go:38 | danger | danger | 원장을 읽을 수 없다. |
| templates_portfolio.go:49 | notice | aside | 브로커 조회가 배선되지 않았다. 수량·평단·현재가·평가손익은 표시되지 않고, |
| templates_portfolio.go:52 | notice | aside | 검증 중 — 데이터 없음. . 실계좌 검증이 rate limit을 쓰고 있으므로 |
| templates_portfolio.go:55 | notice | aside | 검증 중 — 갱신 보류. . 아래 브로커 수치는 캐시된 |
| templates_portfolio.go:65 | notice | aside | 브로커 보유를 아직 읽지 못했다. |
| templates_portfolio.go:67 | danger | danger | 브로커 조회 실패:  — 위 표는 실패 이전의 |
| templates_portfolio.go:104 | status-pill | (상자 class 없음) |  |
| templates_portfolio.go:107 | status-pill | (상자 class 없음) |  |
| templates_portfolio.go:112 | submetric | submetric warn | 편입 예약 · 실행 미확인 |
| templates_portfolio.go:136 | notice | status-pill warn | 편입 예약됨 · 아직 보호 미적용 |
| templates_portfolio.go:137 | notice | status-pill warn | 편입 예약됨 · 실행 상태 미확인 · 아직 보호 미적용 |
| templates_portfolio.go:152 | notice | (상자 class 없음) |  |
| templates_portfolio.go:155 | notice | status-pill warn | 원장 기록 · 실효 미확인 |
| templates_portfolio.go:181 | notice | aside | 관리 포지션 없음. 원장은 정상적으로 읽혔고, 엔진이 관리 중인 미청산 |
| templates_portfolio.go:195 | notice | aside | 관리 여부 불명 보유가 있다. 엔진 원장을 읽지 못했으므로 아래 |
| templates_portfolio.go:199 | notice | aside | 엔진 원장에 포지션이 없는 보유가 있다 — 엔진이 진입한 포지션이 |
| templates_portfolio.go:257 | notice | (상자 class 없음) |  |
| templates_portfolio.go:285 | notice | aside | 이 표의 한계. 성과 행(trade_outcomes)은 엔진이 관리하던 포지션이 |
| templates_portfolio.go:298 | notice | aside | 최근 기록만 표시한다 — 창을 넘는 오래된 판정은 생략되었다. |
| templates_position_policy.go:9 | notice | (상자 class 없음) |  |
| templates_position_policy.go:10 | danger | danger | 불러오기 실패: |
| templates_position_policy.go:11 | notice | aside | engine-owned PositionPolicyCommander가 배선되지 않아 조회만 가능하다. |
| templates_position_policy.go:12 | notice | aside | 판정 격리 상태를 읽지 못했다: . 격리된 포지션이 있어도 이 화면에 표시되지 않는다. |
| templates_position_policy.go:15 | notice | aside | exit 평가 근거를 확인할 수 없어 현재 보호선과 다음 exit 단계를 표시하지 않는다. |
| templates_position_policy.go:31 | notice | (문구 변경·이동 — diff 참조) | desired 알 수 없음 · |
| templates_position_policy.go:32 | danger | danger | desired validation 거부 · |
| templates_position_policy.go:33 | notice | (문구 변경·이동 — diff 참조) | effective 알 수 없음 · |
| templates_position_policy.go:42 | danger | danger | 지속성영구 차단일시 차단 |
| templates_position_policy.go:43 | notice | aside | 현재 runtime projection에 active 대사 차단이 없다. |
| templates_position_policy.go:59 | submetric | (상자 class 없음) | 대사 차단 ·  ·  ·  ·  · 영구 차단 |
| templates_position_policy.go:60 | submetric | (상자 class 없음) | active exit 있음 |
| templates_position_policy.go:61 | submetric | (상자 class 없음) | 판정 격리 v ·  ·  경과 · 유지 보호선 |
| templates_position_policy.go:62 | submetric | (상자 class 없음) | 이 포지션은 지금 exit 판정 대상이 아니다 — 손절도 평가되지 않는다 |
| templates_position_policy.go:74 | notice | aside | 관리할 현재 포지션 행이 없다. |
| templates_position_policy.go:91 | danger | danger | 보호 공백 위험. active exit가 있으면 409로 거부된다. |
| templates_position_policy.go:93 | notice | aside | 일반 정책 변경은 추가 대기나 입력 없이 이 preview 그대로 적용된다. |
| templates_settings.go:50 | notice | (상자 class 없음) |  |
| templates_settings.go:69 | notice | aside | — |
| templates_settings.go:75 | notice | aside | — |
| templates_settings.go:90 | notice | (상자 class 없음) |  |
| templates_settings.go:91 | danger | danger | 설정 파일을 읽을 수 없다: |
| templates_settings.go:92 | danger | danger | 현재 블록은 엔진이 거부한다. |
| templates_settings.go:127 | notice | muted effect-timing | 반영 시점 — |
| templates_settings.go:146 | notice | (상자 class 없음) |  |
| templates_settings.go:155 | notice | aside | 이 판정은 파일이 답하는 절에 한한다. Guardian 주입·게이트웨이 |
| templates_settings.go:159 | notice | notice | 게이트를 켜면 시작되는 것. 다음 엔진 기동에서 대사·exit 관측· |
| templates_settings.go:182 | notice | muted effect-timing | 반영 시점 — |
| templates_settings.go:193 | notice | (상자 class 없음) |  |
| templates_settings.go:195 | danger | danger | 엔진 자동 시작 설정을 읽을 수 없다: |
| templates_settings.go:213 | notice | aside | 마지막 엔진 기동 결과: |
| templates_settings.go:234 | notice | aside | 편입 규칙은 상시 탭으로 |
| templates_settings.go:243 | notice | (상자 class 없음) |  |
| templates_settings.go:244 | danger | danger | 한도를 읽을 수 없다: |
| templates_settings.go:259 | notice | (상자 class 없음) |  |
| templates_settings.go:301 | notice | (상자 class 없음) |  |
| templates_settings.go:342 | notice | muted effect-timing | 반영 시점 — |
| templates_settings.go:351 | notice | (상자 class 없음) |  |
| templates_settings.go:352 | danger | danger | 거래 정책을 읽을 수 없다: |
| templates_settings.go:376 | notice | muted effect-timing | 반영 시점 — |
| templates_settings.go:406 | notice | (상자 class 없음) |  |
| templates_settings.go:408 | danger | danger | 알림 설정을 읽을 수 없다: |
| templates_settings.go:417 | notice | (상자 class 없음) |  |
| templates_settings.go:483 | notice | muted effect-timing | 반영 시점 — |
| templates_settings.go:494 | notice | (상자 class 없음) |  |
| templates_settings.go:524 | notice | notice | 이 프로세스에서 서명 검증됨. |
| templates_signals.go:52 | notice | aside | 시장이 하나도 없다. 저장소는 읽혔는데 평가된 시장이 |
| templates_signals.go:56 | notice | (문구 변경·이동 — diff 참조) | 미측정 () — |
| templates_signals.go:65 | notice | (문구 변경·이동 — diff 참조) | 미측정 () — |
| templates_signals.go:86 | notice | (문구 변경·이동 — diff 참조) | 원천 상태 미측정 () — |
| templates_signals.go:96 | notice | aside | 이 스캔은 원천을 하나도 부르지 않았다. 일정상 아직 부를 |
| templates_signals.go:118 | notice | aside | 강등인데 빠진 원천 이름이 없다. 스캔이 결손을 보고하면서 |

총 153자리. 집계(옛→새 첫 class: 자리 수)

- danger → (상자 class 없음): 6
- danger → danger: 32
- notice → (문구 변경·이동 — diff 참조): 5
- notice → (상자 class 없음): 22
- notice → aside: 56
- notice → muted: 5
- notice → notice: 17
- notice → status-pill: 3
- status-pill → (상자 class 없음): 2
- submetric → (상자 class 없음): 4
- submetric → submetric: 1
