# a166 로트 「time namespace 가 다른 참여자」 — 1단계 조사 + 계약 개정안

- 작성: 2026-11-25, 구현 팀메이트. 기준 `origin/main` `6ccc809d3`, 브랜치 `feat/a166-time-namespace`.
- 범위: 조사와 개정안까지. **Go 제품 코드 · 시험 파일 편집 0.** design.md · spec delta · tasks.md 본문 편집 0
  (수정 문안은 `contract-delta.md`). 운영 호스트 · 컨테이너 · 브로커 접근 0.
- 입력: `../host-binding-evidence-2026-11-25.md` 의 B1.

## 0. 사용자 결정(기록)

2026-11-25 사용자 결정 — **B1 처분 = 1번 코드 수정**: 「boot 식별자가 같고 각 참여자의 시간 오프셋이 0 이면
time namespace 신원이 달라도 같은 시계로 받는다」. 배치 변경(모든 참여자를 한 time namespace 로)은 고르지 않았다.
조사 결론: 아래 §4 · §5 근거로 이 방향은 **현재 지원 전제 안에서 안전하게 성립한다**(새로 열리는 구멍 0, §5 표).

## 1. 결론 먼저

1. 지금 코드는 time namespace 신원이 다르면 **같은 boot · offset 0 이라도** hint 복원을 거절한다. 실제 커널 time
   namespace 로 재현했다(§4, `run-current-head.log`). 거절 지점은 `clock.RemainingBootHint` 의 분기 B3/B4 이고,
   그 앞 단계(`OpenObservation`, `Begin`, 사건 기록)는 통과한다.
2. 거절은 일시적이지 않다. **만료된 hint 도** 다른 namespace 에서 읽으면 계속 거절된다(S6). namespace 검사가
   경과 계산보다 앞에 있어서 그 slot 을 지우는 경로(만료 삭제 · prune)가 모두 같은 오류로 막힌다 — 다음 재부팅
   (boot_id 변경)까지. 컨테이너를 다시 만들 때마다 time namespace 가 새로 생기면(운영 실측: 컨테이너마다 다름)
   **참여자가 엔진 하나뿐이어도** 배포 뒤 그 poll 종류가 영구 실패한다. spec 의 「유효한 폴링 hint와 프로세스
   재시작」 시나리오가 컨테이너 재생성에서는 성립하지 않는다(재생성 시 새 namespace 여부는 운영에서 안 봄 — S3 은
   모사).
3. 교차 namespace 에서 `Finish(429)` 가 실패하면 pending 이 남고, 다음 `Begin` 이 그 slot 을 **영구 poison** 한다
   (S2 → S2b 실측). 운영 흐름에서는 `PollRemaining` 이 먼저 막아 이 경로까지 가기 어렵다(§3.3).
4. 신원 비교를 빼도(권장안 가) 지원 전제 안에서 새로 열리는 구멍은 0이다. 가드 없이 새로 통과하는 정상 밖 입력은
   둘 — 「손으로 고친 저장 값」(원래도 가드 아님: 같은 namespace 값으로 위조하면 지금도 통과)과 「같은 boot_id 를
   보이는 서로 다른 실제 boot」(신뢰된 procfs · 단일 호스트 전제 밖, 호스트 ns 참여자끼리는 지금도 통과) — 이다. 오프셋 비영 namespace 는 `Snapshot` 이 실제 커널에서
   거절한다(S5). namespace inode 는 재사용된다(실측: `4026534476`, `4026534477` 이 연속 실험에서 다른 namespace 에
   다시 붙음) — 지금의 신원 비교는 「같은 시계」를 보증하지도 못했다.
5. 권장안(가): `RemainingBootHint` 한 함수의 4개 조건에서 `Namespace` 항만 뺀다. 저장 형식 · 버전 · 골든 무변.
   overlay 로 잰 결과 기존 시험 중 **3개 함수만** 단언이 뒤집히고 나머지 선택 패키지는 그대로 통과한다(§6).

## 2. `Namespace` 신원 비교가 들어간 이유(인용)

찾은 범위: a166 `proposal.md` · `design.md` · `tasks.md` · `review.md` · `issues.md` · `specs/`, `analysis/` 전체 md
(`rg -i 'namespace|boottime'`), `poll-clock-lot/` 의 pre-edit · review · manager-verification, 시험 이름, git 이력
(`bootstamp*.go` 3 리비전 — 처음 커밋 `acd23170f` 부터 비교 4곳 · 오프셋 0 검사 1곳이 그대로).

- `analysis/poll-clock-contract.md` 「복원 산술 계약」: 「동일 boot 및 동일 namespace이면 now.uptime >= saved.uptime일
  때만 차이를 빼며 … **동일 boot인데 namespace가 다르거나 uptime이 역행하면 오류다.**」
- 같은 문서 「첫 지원 범위」: 「실행 중 setns/unshare/time namespace 교체 없음, 현재 namespace와 children namespace가
  동일하며 monotonic/boottime offset 모두 0인 환경으로 제한한다. **실제 host/container는 운영 전 별도 확인한다.**」
- `design.md` 77행: 「같은 boot/time namespace에서는 주입된 공유 boottime 표본과 duration으로 남은 시간을 구한다.」
- `poll-clock-lot/review.md`: 「동일 boot에서는 namespace와 시간 순서를 검증한다.」 시험 이름
  `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor`, `TestBootZeroDurationDoesNotHideInvalidClock`
  (「zero masked mismatch」), `TestBootRemainingRejectsAnchorBeforeSameBootProvenance`(「conflicting same boot anchor」),
  변이 기록 `mutation-namespace.log`(최초 생존) → `mutation-namespace2.log`(보강 후 검출).

**무엇을 막으려 했는지(구체 위협)는 찾은 범위에 없음.** 문서는 신원 일치를 「같은 시계」의 전제로 적었을 뿐,
오프셋 0 검증만으로 부족한 반례를 적지 않았다. 「실제 host/container는 운영 전 별도 확인」으로 미뤄 둔 것이
2026-11-25 실측(B1)에서 드러났다.

## 3. `Namespace` 사용 자리 표

산출: `ast/*.json`(18 함수, `tools/logic-map`), 자리 열거 `census.py` → `census.txt`(AST 분기 · 호출 줄 + 함수 범위 안
`Namespace` 토큰 줄). 저장소 전체 `rg Namespace --type go` 는 위 함수들과 시험 외에 `strategyprojectionrpc` · `app/engine`
시험 2건(시험 이름의 「CrossNamespace」, 네트워크/유닉스 소켓 namespace — 무관)뿐이다.

### 3.1 (a) 한 프로세스 안의 앞뒤 표본 일관성 — 4 자리, 유지

| 함수 | 줄 | 내용 |
|---|---|---|
| `systemBootSource.Snapshot` | 140 · 144 | 현재 · 자식 time ns 읽기(`probe.namespace`) |
| 〃 | 145 (B6) | `current != children` 거절 — offsets 파일이 자식 ns 를 보여 주므로 현재 ns 의 오프셋임을 보장 |
| 〃 | 163 · 167 · 168 (B12) | BOOTTIME 표본 뒤 재읽기, 앞과 다르면 거절 |
| `systemBootProbe.namespace` | 58–67 | stat device/inode, 0 이면 거절 |

비교 자리로 세면 B6 · B12 두 분기(읽기 4회 포함). `Snapshot` 179행이 결과에 `Namespace: current` 를 넣는다.

### 3.2 (b) 서로 다른 프로세스가 남긴 값끼리 비교 — 4 자리(모두 한 함수), 바꿀 대상

| 함수 | 분기 | 비교 | 입력 출처 |
|---|---|---|---|
| `clock.RemainingBootHint` | B2 L67 | `anchor.Namespace != saved.Namespace`(같은 boot 일 때) | 저장 hint 의 stamp vs anchor — 다른 프로세스가 썼을 수 있음 |
| 〃 | B3 L70 | `now.Namespace != saved.Namespace` | 현재 프로세스 표본 vs 저장 stamp |
| 〃 | B4 L73 | `now.Namespace != anchor.Namespace` | 현재 표본 vs 저장 anchor |
| 〃 | B8 L86 | `now.Namespace != base.Namespace` | 현재 vs base(=saved 또는 anchor). **이미 죽은 분기**: B7 통과 뒤 now·base 는 같은 boot 이고 B3(anchor 없음) 또는 B4(anchor 있음)가 같은 조건을 먼저 거절한다 |

이 함수를 부르는 자리(AST 호출 줄): `pruneObservationSlots` L396, `ObservationStore.Finish` L531,
`ObservationStore.PollRemaining` L598, `validateObservationState` L230(stamp vs 자기 anchor), `validObservationStamp`
L165(`(0, st, st, nil)` — 자기 자신과 비교라 신원 차이가 생길 수 없음, 사실상 `validStamp`).
`validStamp`(L58)는 `Namespace.Device/Inode != 0` 만 본다 — 비교가 아니라 형식 검사.

비교가 없는 자리: `accountbudgetruntime.checkOffline` L287 은 `Snapshot()` 성공만 보고 값을 버린다.
`accountbudgetruntime.Open` L95 는 `clock.SystemBootSource()` 를 주입만 한다. `ObservationStore.sample` L373 · L377 은
표본 + 형식 검사. a087/a169 소비자(`ObservationSnapshot`, `NewObservationFence`)는 사건의 `Stamp` 를 읽지 않는다
(`rg '\.Stamp\b'` 생산 코드 0). `tools/a166-budget-state` 는 `ProvisionOffline` · `CheckOffline` 만 부른다.

### 3.3 (c) 디스크에 쓰거나 읽는 자리 — 형식 무변(권장안 가)

- 형식: `event.state`(`observationState`, `Version: 1`). `clock.BootStamp` 는 json 태그가 없어
  `{"BootID","Namespace":{"Device","Inode"},"Uptime"}` 로 직렬화된다. 들어가는 곳 3필드:
  `ObservationEvent.Stamp`(observation.go L49), `observationHint.Stamp`(L82), `observationHint.Anchor`(L84).
- 쓰기: `ObservationStore.write`(L279–360) — `Finish` L542 · L546, `pruneObservationSlots` L401–402,
  `Finish` L537–538, `PollRemaining` L603–604 가 만든 값.
- 읽기 · 검증: `ObservationStore.read`(L238–272, `DisallowUnknownFields` + 재직렬화 바이트 일치 = canonical 검사) →
  `validateObservationState` L181(사건 stamp 형식), L226(hint stamp · anchor 형식), L230(stamp vs anchor 비교, (b) 표의
  B2 경로).
- 골든 · 픽스처: `event.state` 를 담은 testdata 파일 0(`rg 'event.state|"Namespace"'` testdata 0). 시험은 구조체로 만든다
  (`observation_store_test.go` L45 · L367 · L752, 모두 `Namespace{1,1}` 고정). `offline_test.go` L39 · L50 은 provision
  직후 파일 바이트를 비교할 뿐 stamp 가 없다.
- 운영에 provision 된 실제 root 0(OFF). 권장안 가는 형식을 바꾸지 않으므로 버전 처리 불필요. 구 형식 값은 그대로 유효.

### 3.4 a166 census

`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 90 키 중 `internal/clock/*` · `observation_linux.go` 키 0.
권장안이 고칠 `RemainingBootHint` 는 census 밖. (census 안: `Transport.RoundTrip`, `CheckOffline`, `pollError` 등 — 안 고친다.)

## 4. 지금 무엇이 거부되는가 — 실제 time namespace 재현

방법: `run.sh` + `harness-main.go.txt`(제품 코드 무변경, 실제 `clock.SystemBootSource`, 네트워크 0). 새 time
namespace 는 비특권 `unshare --user --map-current-user --time`(이 호스트에서 동작; 커널 7.0.0-43). 임시 root 는 ext4
홈(`~/.cache/a166-tns-run.*`). **운영과의 차이**: 운영 컨테이너는 user namespace 가 호스트와 같지만 여기서는 새 user
namespace 가 같이 생긴다(uid 1000 매핑). 제품 코드는 user namespace 를 읽지 않는다(`rg 'ns/user'` 0). 원문:
`run-current-head.log`.

| 시나리오 | 쓰는 쪽 ns | 읽는 쪽 ns | `OpenObservation` | `PollRemaining` | 이어서 `Begin` / `Finish(429)` |
|---|---|---|---|---|---|
| S1 대조 | 호스트 | 호스트 | nil | 29s, nil | nil / nil |
| S2 실측 배치 모사 | 호스트 | 새 ns(offset 0) | nil | **ERR** `boot time observation unavailable` | nil / **ERR** |
| S2 뒤 | — | 호스트 | nil | 29s, nil | nil / nil |
| S2b 한 번 더 | — | 호스트 | nil | **ERR `observation poll slot poisoned`** | nil / nil |
| S3 재시작 모사 | 새 ns X | 새 ns Y | nil | **ERR** | nil / **ERR** |
| S4 역방향 | 새 ns | 호스트 | nil | **ERR** | nil / **ERR** |
| S6 만료 뒤(40s+) | 새 ns(소멸) | 호스트 | nil | **ERR**(만료돼도) | nil / **ERR** |
| S6 대조 | 호스트 | 호스트 | nil | 0, nil(만료 삭제) | nil / nil |

같은 boot(`ad865048..`), uptime 은 namespace 와 무관하게 같은 초로 찍혔다(S2: 136005 → 136006).

### 4.1 실패 순서(경로별, AST 인용)

1. **엔진 poll 사전 확인** `execgw.Retrier.pollContext` L65 → `Budget.PollRemaining` → `ObservationStore.PollRemaining`
   L594 표본 → L598 `RemainingBootHint` B3(hint 만 있음) 또는 B4(anchor 있음) → L600 `ErrUnavailable` 반환. 그 key 에
   hint slot 이 있을 때만. slot 이 없으면(L583–585) 0 반환으로 통과.
2. **transport** `Transport.RoundTrip` L302 `Begin` 성공 → L309 `PollRemaining` 같은 오류 → fence 가 있으면 L311
   `fence.latch` → L329 `Finish(NoResponse)`(표본 안 함, 성공) → 요청 안 보냄.
3. **429 기록** `Finish` L510 표본 → L514 진입 조건(`result == RateLimited` · `pending.Kind == Poll` · 정규화 뒤
   `hint != nil && *hint > 0`)을 모두 만족하고 그 key 에 기존 hint 가 있으면 L531 B3/B4 오류. 새 429 의 hint 가
   없거나 0 · invalid 이면 L531 을 건너뛰고 사건은 기록된다 → state 미기록, pending 잔존 →
   다음 `Begin` L440–461 이 그 pending 을 gap 으로 바꾸며 slot 을 `Poisoned` 로 만든다(S2b). poison 을 푸는 코드 0
   (`rg 'Poisoned = false'` 0).
4. **prune** — `Begin` 은 이전 poll pending 의 key 가 slot 에 없고 slot 이 256 개일 때(L440–450), `Finish` 는 새 유효
   poll hint 의 key 가 slot 에 없고 slot 이 256 개일 때(L514–518) → `pruneObservationSlots` L396 → 다른 ns 의 hint 가
   하나라도 있으면 오류. 운영 key 수(계좌 × 4 종류)로는 도달하기 어려움.
5. **anchor**: 다른 boot 의 hint 를 처음 복원한 ns 가 anchor 를 쓰면, 같은 boot 안에서 다른 ns 참여자는 B4 로 거절.
6. **거부 안 되는 것**: `OpenObservation`(read → validate 는 stamp 와 자기 anchor 만 비교; anchor 는 다른 boot 에서만
   생기므로 정상 흐름에서 B2 신원 차이 없음), 사건 기록(stamp 형식만), `checkOffline`(값 비교 없음), a087/a169 독자.

결론: 실패는 **hint 복원 · 429 hint 갱신 · 만료 삭제**에 모인다. offline check 와 사건 독자는 영향 없다.
design 73행은 「price 종류를 공유하는 보호·일반 가격 조회가 함께 보류될 수 있」다고 적는다 — 영구 오류가 보호 가격
조회에 닿을 수 있다(엔진 쪽 소비 경로는 이번에 추적 안 함, 운영 OFF 라 현재 영향 0).

## 5. 고치면(권장안 가) 새로 받아들이는 입력과 그것을 막는 가드

| 입력 | 지금 | 권장안 가 뒤 | 막는 가드(AST) | 근거 |
|---|---|---|---|---|
| 같은 boot, 다른 ns, 각자 offset 0 | 거절(B3/B4) | **통과(의도)** | — | 실측 S2/S3/S4/S6 overlay: 29s · 만료 0 정상 |
| offset 비영 ns 의 표본 | `Snapshot` 거절 | 같음 | `Snapshot` L152 `parseZeroOffsets` | 실측 S5(`--boottime 100`, `--monotonic 100` 둘 다 ERR). 오류 값이 같아 B6(L145)과 구분되지 않는다 — 같은 unshare 구성의 S2 에서 current==children 이 성립했으므로 L152 라는 것은 **추론**(리뷰 R1-K4) |
| 현재 ≠ `time_for_children` | `Snapshot` 거절 | 같음 | `Snapshot` B6 L145 | fake 시험 `TestBootSnapshotFailureCases`(실제 ns 실측 아님) |
| 표본 앞뒤에서 **관측된** ns 불일치 | 거절 | 같음 | `Snapshot` B12 L168 | fake 시험. A→B→A 같은 중간 교체는 앞뒤 비교로 못 잡는다 — 원 계약의 「실행 중 namespace 교체 · thread별 조작 없음」 전제(poll-clock-contract 13행)에 기댄다(리뷰 R1-C2) |
| 두 표본 사이 ns 교체(한 프로세스) | 거절(B3) | 통과 — 새 ns 도 offset 0 검사를 통과해야 표본이 생기므로 같은 시계 | `Snapshot` L152 | man time_namespaces(7). 원 계약은 실행 중 교체를 지원 밖으로 둔다 |
| 같은 boot_id 인데 실제로는 다른 boot(boot_id 가상화 · bind 위조 · VM 복제) + 비초기 ns 참여자 | 거절(ns 가 다르면) | **통과 — 가드 없음** | 없음(`procFS` 는 `/proc` 루트 statfs 만, L33–41) | **지원 전제 밖**(신뢰된 kernel procfs · 단일 호스트, poll-clock-contract 13행). 초기 ns inode 는 상수 4026531834 라 호스트 ns 참여자끼리는 지금도 통과 — 신원 비교는 원래 이 경우를 일부만 막았다. 지원은 범위 확장(리뷰 R1-C1, R1-K2) |
| 다른 boot | NeedsAnchor | 같음 | B7 L83 | 기존 시험 |
| 같은 boot 에서 uptime 역행 | 거절 | 같음 | B3 · B4 의 `Uptime <` 항 | 기존 시험(겹침 부분 유지) |
| 같은 boot 의 anchor 가 stamp 보다 앞섬 | 거절 | 같음 | B2 의 `Uptime <` 항 | `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` 앞 절반 |
| 손으로 고친 저장 값(ns 만 다르게) | 거절 | **통과** | 없음 | 원래도 가드 아님 — ns 를 같게 고치면 지금도 통과. 관리자 비변조 전제(`429-store-draft.md` 25행) |
| 손으로 고친 uptime/boot_id | 형식만 검사 | 같음 | `validStamp` 형식 | 위와 같음 |
| 다른 BootSource 로 만든 stamp | — | — | 생산 주입은 `accountbudgetruntime.Open` L95 의 `SystemBootSource()` 하나 | `rg OpenObserved` 생산 1곳 |
| 죽은 ns 의 inode 를 물려받은 새 ns | 같은 ns 로 오인해 통과 | 통과 | (신원 비교는 원래 이것을 못 가름) | 실측: inode 재사용 2건 |

**어느 가드에도 안 걸리게 되는 정상 밖 입력**: 2건, 둘 다 지원 전제 밖.
(1) 「ns 만 다르게 손으로 고친 저장 값」 — 지금도 같은 ns 로 위조하면 통과하므로 보호가 줄지 않는다.
(2) 「같은 boot_id 를 보이는 서로 다른 실제 boot」 — 신뢰된 procfs · 단일 호스트 전제 밖. 호스트 ns 참여자끼리는 지금도
통과한다. 지원 전제 안에서 새로 열리는 구멍은 0. 설계 중심 문제로 올리지 않되 잔여 위험으로 남긴다(review.md).

시계 전제의 출처: 로컬 `man 7 time_namespaces`(man-pages 6.7, 2023-10-31) — 「offsets, expressed with respect to the
initial time namespace」, 「after the first process has been created in or has entered the namespace, write(2)s on
this file fail with the error EACCES」, 「In a new time namespace created by unshare(2), the contents of the
timens_offsets file are inherited from the time namespace of the creating process.」 → offset 0 이면 그 ns 의
CLOCK_BOOTTIME 은 초기 ns 의 값과 같고, 멤버가 생긴 뒤 offset 은 바뀌지 않는다(`Snapshot` 을 부른 프로세스 자신이
멤버). 웹(man7.org) 채널은 이번에 안 봤음 — 로컬 man 페이지만. setns 가 다중 스레드 프로세스에서 time ns 를 거부하는지는
커널 소스를 안 봐서 확인 못 함(권장안의 안전성은 이것에 기대지 않음: 새 ns 도 offset 0 검사 통과가 필요).

## 6. 개정안

### 권장안 (가) — 프로세스 간 비교에서 신원 항만 뺀다

- 편집할 기존 함수: **`clock.RemainingBootHint` 1개**(B2 · B3 · B4 · B8 조건에서 `Namespace !=` 항 삭제,
  `Uptime <` 항 유지). diff: `overlay-option-ga.diff`(4줄).
- 유지: `Snapshot` 의 (a) 검사 전부, `validStamp` 의 ns 0 아님 검사, 저장 필드 `Namespace`(진단 · 형식 안정).
- 저장 형식 · `Version` · 골든: 무변. 구 형식 값 처분 불필요.
- 새 파일: 시험 1개(예: `internal/clock/bootstamp_crossns_test.go` — 양성 · 거절 사례), 선택적으로 실제 ns 재현 시험
  (§7 RED 6). 새 생산 파일 0.
- 통과 이유가 바뀌는 기존 시험 — **선택 실행에서 실패로 전환된 함수 3개**(overlay 실측,
  `selected-tests-baseline-vs-overlay.log`). 전체 저장소 시험의 전수가 아니다. 정적 근거: `RemainingBootHint` 를 직접
  부르는 시험은 `bootstamp_test.go` 에만 있고 accountbudget 시험의 fake 는 ns 가 모두 `{1,1}` 고정(리뷰 R1-K 정적 확인):
  1. `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` L49 — 「namespace drift accepted
     {Remaining:10s NeedsAnchor:true}」
  2. `TestBootZeroDurationDoesNotHideInvalidClock` L67 — 「zero masked mismatch <nil>」
  3. `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` L130 — 「conflicting same boot anchor accepted」
  나머지: `internal/clock` 다른 시험, `internal/accountbudget`(49.6s), `internal/accountbudgetruntime`(24.4s),
  `tools/a166-budget-state` 는 overlay 에서도 ok. `internal/execgw -run 'A166|EnginePoll'` 무태그 · `tossos_testseams`
  각각 대조군 · overlay 모두 ok — 원문 `selected-tests-execgw.log`. 1·3번 시험 안의 uptime 역행 단언은 보존한다.
- FLM/BTM 대상(gate ⑤ 는 시험 함수 본문 편집도 셈): 생산 1 + 위 시험 3 = **4 번들**. 시험 3개를 고치지 않고
  지우기만 해도 본문 변경이라 같다.
- 비례 원칙: 429 hint 복원은 엔진 폴링(보호 가격 조회 포함) 게이트를 바꾸므로 High-risk 쪽으로 취급 —
  변이 원장 · 적대 리뷰 적용 권장(Manager 결정 1).

### 대안 (나) — `BootStamp` 에서 `Namespace` 를 없애고 「offset 0 검증됨」을 계약으로만 둔다

- 이점: 비교 안 하는 필드가 남지 않는다. 단점: `BootStamp` 형식 · `validStamp` · `Snapshot` 결과 · `event.state`
  canonical 형식이 바뀐다 → `DisallowUnknownFields` 때문에 구 값은 읽기 실패, `Version` 2 와 처분 규칙 필요, 편집 함수
  ≥5 와 시험 다수. 운영 root 0 이라 이행 위험은 낮지만 diff 가 커진다. 「검증됨」 bool 을 저장해도 위조 가능하므로
  보호가 늘지 않는다. **비권장.**

### 대안 (다) — 배치 변경(모든 참여자를 한 time namespace)

사용자가 고르지 않음. 참고: Docker 가 컨테이너마다 새 ns 를 만드는 한 컨테이너 재생성마다 같은 문제가 다시 생긴다
(§1-2) — 배치만으로는 「재시작」 시나리오를 닫지 못한다.

## 7. 2단계 계획(승인 뒤)

RED(모두 먼저 실패 확인; 1–3 은 현재 HEAD 에서 RED, 4–8 은 현재도 GREEN 인 회귀 핀):

1. 같은 boot · 다른 ns · 같은 uptime 기준: saved(ns A) → now(ns B) 경과 차감(남은 시간 = duration − 경과).
2. 같은 boot · 다른 ns 의 만료 hint → `Remaining 0`(slot 삭제 가능).
3. 다른 boot anchor(ns C) 뒤 같은 boot 의 다른 ns(D) 재시작 → anchor 기준 차감.
4. 같은 boot · 다른 ns 라도 uptime 역행 → `ErrBootStamp`(B3/B4 의 남는 항).
5. 같은 boot anchor 가 stamp 보다 앞섬(ns 다름) → `ErrBootStamp`(B2 의 남는 항).
6. accountbudget 저장소 수준(**새 시험 파일**에 둔다 — 기존 `observation_store_test.go` 에 넣으면 인접 함수가 gate ⑤
   번들로 잡힐 수 있음, 리뷰 R1-K6): fake BootSource 두 개(ns 만 다름)로 `Finish(429)` → `PollRemaining`(다른 ns) 남은 시간,
   만료 뒤 slot 삭제, `Finish` 갱신이 poison 을 만들지 않음(S2b 반례).
7. `Snapshot` offset 비영 · current≠children · 표본 중 교체 거절(기존 시험 유지 확인).
8. (선택) 실제 커널 ns 시험: `unshare --user --map-current-user --time` 가 되는 환경에서만 돌고, 안 되면 Skip 이 아니라
   「검사 0」으로 보고하는 형태 — 기본 스위트에 넣을지는 Manager 결정 3.

변이(사본에서, 하네스 시작 sha 단언 · 무변이 대조군 GREEN 먼저):
- M1 B3 의 `Uptime <` 삭제 → RED 4 가 잡음. M2 B4 의 `Uptime <` 삭제 → RED 4(anchor 판). M3 B2 의 `Uptime <` 삭제 →
  RED 5. M4 `Namespace !=` 를 되살림(B3) → RED 1·2·6. M5 B7(다른 boot) 를 같은 boot 처럼 차감 → 기존 시험.
  M6 B8 삭제 → **동등 변이 예상**(§3.2 죽은 분기) — 지울지 둘지는 Manager 결정 2.
  M7 `Snapshot` L152 오프셋 검사 삭제 → 기존 `TestBootSnapshotFailureCases`.
- 선택 시험 패키지: `./internal/clock`, `./internal/accountbudget`, `./internal/accountbudgetruntime`,
  `./tools/a166-budget-state`, `./internal/execgw -run 'A166|EnginePoll'`(무태그 + `-tags=tossos_testseams`).
  CodeGraph 1.6.0 `affected --filter '*_test.go'`: `bootstamp.go` 1,294 · `observation_linux.go` 1,287 파일(파일 단위
  추이라 선택 근거로 못 씀 — 위 목록은 직접 호출자 `rg` 기준).
- `-race`: `RemainingBootHint` 는 순수 함수, 저장소 잠금 경로 무변 → 필요 없음. 단 RED 6 이 프로세스/고루틴을 쓰면
  `internal/accountbudget` 선택 race 1회.

## 8. 하지 않은 것 · 확인하지 못한 것

- 리뷰 근거(미검증 기억): setns(time)은 단일 스레드에서만 되고(EUSERS) 새 스레드는 부모의 time ns 를 물려받는다는 설명
  (R1-K5), `unix.ClockGettime` 은 vDSO 가 아닌 raw syscall(x/sys v0.46.0 zsyscall_linux.go:585–590, R1-C2) — 커널 소스
  인용은 2단계에서 필요하면 한다.
- 안 봄: 운영에서 같은 컨테이너 재시작 · 재생성 시 time ns 가 새로 생기는지(S3 은 모사), 운영 컨테이너 안
  `time_for_children` 값, 커널 소스(setns 단일 스레드 조건 · offsets 파일이 가리키는 ns), man7.org 웹 판본,
  엔진에서 보호 가격 조회가 `PollRemaining` 오류를 어떻게 소비하는지.
- 없음(찾아서 0): 신원 비교 도입의 구체 위협 서술, `event.state` 골든 · testdata, 생산 코드의 사건 `Stamp` 소비자,
  poison 해제 코드, census 의 clock/observation 키.
- 전체 시험 · 전체 게이트 · `make sdd-check` 미실행. `make sdd-sync` 미실행(CodeGraph 는 main 체크아웃의 1.6.0 색인을
  `6ccc809d3` 에서 읽음).
- B2 · B3 은 범위 밖. 얽힌 점: B3 의 호스트 도구(`a128`/`a140` go run, 사람 CLI)는 `ObservationOther` 종류라 poll slot
  을 만들지 않으므로 B1 의 실패 경로에 거의 닿지 않는다(§4.1-4 의 prune 만). B1 의 실제 피해자는 컨테이너 안 엔진이다.

## 9. Manager 결정 대기

1. **위험 등급** — 권장: High-risk 취급(변이 원장 · 적대 리뷰 유지). 근거: 429 hint 복원은 엔진 폴링(보호 가격 조회
   포함, design 73행) 의 통과 여부를 직접 바꾼다. 경량으로 두면 FLM 4 번들만 남고 변이 기록이 빠진다.
2. **B8(L86) 처리** — 권장: `Namespace` 항만 빼고 분기는 둔다(최소 diff). 근거: 이미 죽은 분기라 동등 변이(M6)로 기록하면
   되고, 지우면 편집 줄이 늘 뿐 행동은 같다. 대안: 지우고 동등 변이 기록을 없앰.
3. **실제 커널 ns 시험을 기본 스위트에 넣을지** — 권장: 넣지 않고 `analysis/time-namespace-lot/run.sh` 재현을 2단계 수용
   증거로 다시 돌린다. 근거: `unshare --user --time` 은 호스트 설정(비특권 userns)에 달려 CI · 컨테이너에서 「검사 0」이
   되기 쉽다. 대안: 태그 시험으로 넣고 불가 시 실패가 아니라 명시 보고.
4. **tasks.md 에 이 로트 줄을 둘지 · 4.2 와의 관계** — 권장: 새 task 줄(예: 4.2 앞 선행)을 Manager 가 추가하고, 4.2 는
   B1 수리 착지 + B2/B3 처분 뒤 다시 실측해서만 체크. 근거: 이 로트는 4.2 의 미충족 1건만 닫는다.
5. **계약 개정 리뷰 등급** — 권장: 이 로트의 2 보이스(적대 Eng 2, 그중 다른 계열 1) 기록을 「requirement 변경」 리뷰로
   인정하고, 2단계 본문 반영 diff 에 경량 확인 1회. 근거: 문안이 코드 · 실측과 맞는지는 이미 두 보이스가 대조했다.
6. **잔여 위험 수용** — 「같은 boot_id 를 보이는 다른 실제 boot」(지원 전제 밖)와 「앞뒤 비교로 못 잡는 중간 ns 교체」
   (원 계약 전제)를 issues.md 의 잔여 위험으로 옮길지. 권장: 옮긴다(새 가드는 만들지 않음).
