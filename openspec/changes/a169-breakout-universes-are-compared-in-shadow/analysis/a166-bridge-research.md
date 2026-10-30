# a169 재개 설계 조사 — 2026-10-31

기준: 로컬 `origin/main` cf0f0919e19bd384995b8bfb7fb8ee44a42031a7 (원격 새 조회 없음). `git show` 원문 손읽기·검색만; AST/시험/운영 미실행. a166 부분 착지와 a169 재개·운영 승인은 별개다. ai-memory CLI 검색은 결과 없음; 과거 로컬 memory의 리뷰 이력은 길잡이로만 사용했다.

## 구현 상태
- a169 `tasks.md:37-42`에서 2.1만 완료: `internal/official/stock_reads.go:30-37,85-96`은 `/stocks`의 securityType을 원문 전달, `internal/domain/models.go:426-429`가 필드를 노출한다. 이것은 유형별 분류/영수증/보고서 완료 증거가 아니다.
- 같은 tasks 2.2 이후 미완. 현재 `tools/a140-breakout-signal-shadow/main.go:75-81,95-96`에는 기존 --top과 단일 config-dir만 있고, `shadow.go:263-265`는 거래대금 단일 랭킹이다. U0/U1/U2, profile preflight/enginewatch, a169 cohort와 유형별 집계는 명세이지 현재 기능이 아니다. `main.go:139-156`에는 기존 a140 공통 factory 및 Auxiliary ctx가 이미 있으나 a169 전용 mode/필수 강제/감시는 없다. 환경상 OFF도 가능하므로 전체 참여자 required가 실제 보장된다는 뜻이 아니다.

## 25초의 단위와 상한
- design.md:79-82,134-143,191-212: U2 선정 구간은 정상일 09:30:30~09:30:55, 총 25초 벽시각. 두 랭킹 GET이 먼저 들어가고 `selected_at`은 랭킹 수신 시점이며 보충은 그 뒤 남는 약 24초다. 후보는 최대 200개, 각 미캐시 후보의 일봉 1 GET(25는 봉 개수이지 초가 아님). `ready_at ≤ D`로 전체 보충이 완료돼야 선정, 미완은 U2 전체 late. 과거 250ms 기준 ≤96 보충 GET 및 200개 약 50초는 현재 상한하에 재계산 필요(`issues.md:82-84`, design:193,212).
- 현재 `internal/accountbudget/budget_linux.go:27,612-639`는 Aux가 queue/wire 점유 시 즉시 거절하고 획득 뒤 350ms를 기다린다. 허용된 wire는 응답 처리·해제까지 직렬화된다. 25초 구간에서 350ms 간격만 적용한 관대한 상한은 시작 ≤72회, 그중 랭킹 2회를 빼면 보충 ≤70회다(양 끝에서 시작 가능한 이상화 상한; 네트워크/토큰/다른 호출 전부 0). 보장치 아님.
- 미래 a169가 현 `source.go:84-96`의 동기 `retrying` 패턴을 랭킹과 일봉에 그대로 재사용하고 `shadow.go:51-52`의 250ms 사전 pause에 이어 매번 350ms Aux 획득 지연을 순차로 지불한다는 **가정**에서는, 첫 요청도 0.60초 후이므로 25초 내 floor(25/0.60)=41회, 랭킹 2회를 뺀 보충 ≤39회다. 새 일봉 래퍼 미구현이라 조건부 수치다. 이전 요청의 경합, 실패 즉시 거절, OAuth/401/재시도, 다른 소비자, 저장·계산·네트워크와 스케줄링은 여유를 더 줄이며 성공 보장 최소치는 0이다. 기존 `now + ReadGap < D`만으로 응답/350ms를 포함한 ready_at 마감 준수는 보장되지 않는다; 남은 기한을 request context와 completion 확인에 연결해야 한다. 최대 200 GET 필요 시 선택 자체를 마감 내 완료할 수 없다.

## 관측과 중단의 실제 공백
- `internal/accountbudget/observation.go:13-18,40-79` 이벤트는 attempt·kind·boot stamp와 별도 event/gap cursor를 제공하고 프로세스/엔진/프로필 ID를 제공하지 않는다. `observation_linux.go:616-695` Snapshot은 cursor 범위/소속을 확인하고 event/gap overrun 및 busy/orphan pending을 노출한다. 과거 429를 시작 시 무시할지, 시작 전부터 중단할지 cursor 기준과 재시작/abort 영속 정책을 명세해야 한다. 429 이벤트, gap, overrun, orphan, snapshot 오류/세대 경합은 각각 폐기·재시도·중단 조건을 명시해야 한다. Busy pending은 진행 중이므로 곧바로 장애와 동일시할 수 없다; 확인 가능한 기한과 감시기 자체 고장 처분이 필요하다.
- `observed_linux.go:12-24`/`runtime_linux.go:85-102`의 required factory는 관측 store를 내부 budget에 결속하지만 위 조사에서 해당 store를 감시자가 빌려 Snapshot할 공개 factory accessor는 찾지 못했다. `OpenObservation`은 expected FileID와 budget을 요구한다(`observation_linux.go:100`); manifest의 event_lock을 스스로 실제 값에서 채워 신뢰를 만들지 말아야 한다. 최소한 같은 검증된 store를 읽는 경계 및 lifetime을 설계해야 한다. 기존 코드가 모든 참여 프로세스에서 빠짐없이 required/observed라는 증거는 없다.
- 응답의 429는 `transport_linux.go:329-345`에서 collector와 Finish에 기록한다. 첫 HTTP 결과는 이미 도착했고, Finish 실패는 callback/collector 경로이므로 폴링 Snapshot만으로 즉시 감지·기록 성공을 가정하면 안 된다. 첫 429 발견 뒤 이미 시작한 요청과 동시 취소 사이 경합이 있다. `design.md:435-445`의 '취소 뒤 새 GET 0, 최대 1'은 pause 검사만으로 원자적 전송 경계를 입증하지 못한다. 같은 observed wire/store에 대한 중단 판정과 전송 사이의 원자성 또는 명시적 경합 상한/실패 처분을 따로 증명해야 한다. 보장 못 하면 fail closed하고 한도 넘는 정확 주장 금지. 429 자체를 사전에 예방한다는 뜻도 아니다.
- `source.go:84-103`은 현재 비호가 429를 재시도하며 `106-124` 호가 429는 일시 정지한다. 따라서 a169 첫 429 전체 중단은 기존 a140 행동을 바꾸지 않는 a169 전용 처리로 구현·증명해야 한다. event는 종류별이지만 엔진과 도구를 구분할 식별자가 없어 엔진 전용 429 필터를 주장할 수 없다; 같은 그룹 전체 사건을 중단 조건으로 할지 명시 필요.

## 신원 및 결정
- a166 manifest의 root/queue/wire/event identity 및 scope/instance 결속(`runtime_linux.go:35-42,85-102`)은 예산 저장소 동일성이다. 실제 엔진이 살아 있고 그 저장소를 사용 중인지, 엔진 config-dir이 맞는지, 별도 키/토큰 프로필이 올바른지는 증명하지 않는다. a169 기존 D5 별도 키와 profile G1~G6 및 engine identity/liveness는 독립적으로 유지·재설계 필요(`design.md:358-415`, `issues.md:76-93`). 예산 공유를 이유로 별도 키 요구를 폐기하면 기존 사람 결정 변경이다.
- 기존 정책 보존하에 명세 작업 자체에 새 승인 불필요: 3/s aggregate+보호 우선, Aux 거절, 25초 마감, 미완 U2 late, D5 별도 키, 측정 보류. 그러나 25초 마감·선정 시각·부분 후보 허용·별도 키·첫 429 중단 범위 또는 관측 실패 시 완화 등 실질 정책 변경은 사용자 결정 필요. 실제 리허설/측정은 별도 사람·운영 승인과 a166 전체 재개 조건 검증 전 금지. 구체적 cursor 초기화 및 보수적 fail-closed 구현 선택은 기존 정책에 충돌하지 않으면 설계 상세로 확정 가능하되 취소 경합의 새 허용 범위는 승인 없이 가정하지 않는다.
