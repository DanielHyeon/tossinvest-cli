# a112 5.5 인계 — 다음 세션이 먼저 읽는 문서

작성 2026-09-01, 갱신 2026-09-02. 대상은 이 change 를 이어받는 다음 세션과, 아래
남은 셋을 정해야 하는 사람이다.

한 줄(2026-09-04 갱신): 5 절에서 남은 것은 **5.2.2 · 5.6.2** 이고, 8 절에 **8.7.2** 가
새로 열려 있다. **5.1.2.2 와 8.7.1 이 한 로트로 닫혔다** — 서명된 4-가족 활성화가 서고
네 전략군 레인이 조정자 앞에 선다. 사람 결정 (7) 은 이 로트로 종결됐다(사람이 선택지
1번을 골랐다: 매니페스트 + 5.1.2.2 한 로트, 5.2.2 는 그다음).
그전 한 줄(2026-09-03): 5 절에서 남은 것은 **5.1.2.2 · 5.2.2 · 5.6.2** 였다.
그전 한 줄(같은 날): 5.1.2.2 · 5.2.2 · 5.3.3 · 5.6.2 였다.
그전 한 줄(같은 날): 5.1.2.2 · 5.2 · 5.3.3 · 5.6.2 였다. 그전: 5.1.2 · 5.2 · 5.3.3 · 5.6.2.
그전 한 줄: **5.5 의 backstop 굳히기는 적대 리뷰 13 라운드 끝에 APPROVE(P0/P1/P2 = 0)로
끝났고, 2026-09-02 에 세 커밋으로 랜딩했다.** 사람이 정해야 하는 것 다섯 중 셋
(커밋 단위·a117 스텁·CI 배선)은 답을 받아 처리했고, 둘이 남았다.

상세는 `review.md` 의 `## 5.5 현재 상태` 블록(fix 절들보다 **앞**에 있다)과
`analysis/function-logic/internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority/branch-test-map.md`
의 B3 절에 있다. fix3~fix13 열한 절은 **방법 기록**이지 현재 상태가 아니다 —
뒷 절이 앞 절을 정정하므로 diff 해서 상태를 재구성하지 말 것.

---

## 1. 랜딩한 것 (2026-09-02)

세 커밋으로 나눠 올렸다. 사람이 "3 커밋 분할"을 골랐다.

| 커밋 | 무엇 |
|---|---|
| `14b76963` | 신규 시험·seam 5 + 수정 시험 4. 행동(시험 함수 넷/하위 여섯) · 구조 단언 하나 · 필드 완전성 32+8 |
| `3183e22b` | `tools/logic-map/{role_check,check_analysis,test_role_check,test_check_analysis}.py` — 좌표 역할 대조와 열거 강제 |
| `80d3931d` | 병행 세션 스텁 a117 → a119 renumber + PM 레코드 (아래 2 절 (4)) |
| `364190c8` | `review.md` · `tasks.md` · a112 분석 번들 54 · 이 문서 |
| `38b17426` | CI 배선 — `sdd-check-ci` 타깃 · `sdd-checks` job · 갈라짐을 막는 시험 (아래 2 절 (2)) |
| `4724c3d0` | 5.1.1 — `internal/strategyworker`: 여덟 FamilyWorker · 폐포 증명 · 구조 단언 |
| `b647dc91` | 5.1.1 을 5.1.1(타입, `[x]`) 과 5.1.2(배선, `[ ]`) 로 나눔 |
| `4c4fe685` | 5.1 닫힘(runtime policy) + 5.3 을 5.3.1(`[x]`)/5.3.2(`[ ]`) 로 나눔 — `policy.go`·`lane.go` |
| `5389ff1e` | `origin/main`(52fb9bb2 WTS 카탈로그) 병합 — 병합 뒤 fingerprint 가 stale 이 되어 `make sdd-sync` 재실행 필요했다 |
| `472ce405` | 5.3.2 랜딩(단일 비행·카덴스·마감 시한) + 5.3 의 두 번째 분할 → 5.3.3(`[ ]`, durable latch) — `bounded.go` |
| `89a626c3` | 5.7 랜딩 — 여덟·둘 리허설, **경합 검출기 배선**(`make test-race`, gate 9/11, CI), `Step` 에 오류 반환 추가 |
| `70f0487d` | 5.6.1 랜딩 — 엔진 빈칸 일곱을 재서 채우고, "무엇이 엔진을 세울 수 있는가"를 패키지 전체 열거로 얼렸다. **생산 코드 변경 0** |
| `19976352` | 5.1.2.1 랜딩 — 여덟 레인이 생산 런타임 안에 서고, 전략군의 사이클이 `*Context` 클로저가 아니게 됐다. 여덟은 여전히 DORMANT 이므로 **생산 동작 변화 0** |
| `a5561ecf` | 5.2.1 랜딩 — 원격 권한 수집이 공유 assembly mutex 밖으로 나갔다(single-flight 파도). 두 시장이 한 파도를 타고, 기다리는 쪽이 `ctx` 를 본다. `internal/app/engine` 의 동시성 시험이 처음으로 `-race` 아래 돈다 |
| `(이 로트, 3 커밋)` | **8.7.1 + 5.1.2.2 랜딩 (2026-09-04)** — `internal/strategyrouter` 에 서명된 4-가족 활성화(`strategy-family-activation-<MARKET>.json`, ed25519 · digest 핀 · 정규 직렬화 등식 · 정확히 네 서술자)가 서고, `FamilyWorker` 의 `desired`/`effective` **필드가 없어져** 활성화의 함수가 됐다. 네 레인이 `coordinateMarketProposals` 의 조정자 `Submit` **앞**에 서고, 레인은 제안 수집 **앞**에서 세워진다(재시작 첫 주기의 durable latch 때문). 생산에는 서명 매니페스트가 하나도 없으므로 **생산 동작 변화 0** |
| 그전 커밋 | 5.3.3 랜딩 — 레인 잠금이 원장(schema **v32**)에 남아 재시작을 견딘다. 복구는 **엄격히 더 큰 서명 활성화 세대**가 있어야 하고 그 판정은 SQLite 트리거가 한다. 여덟은 여전히 DORMANT 라 **오늘 쓰이는 행은 0** |

랜딩 직전 실측: `make lint` PASS · `make test` 98 ok/0 FAIL · `make test-seams`
99 ok/0 FAIL · python 77 OK · `check_analysis --change a112` evidence complete ·
`openspec validate a112 --strict` valid · `gofmt -l internal/ tools/` 무출력.

## 2. 사람이 정해야 하는 것 — 리뷰로는 못 정한다

(2)·(3)·(4) 는 2026-09-02 에 답을 받아 닫았다. (1)·(5) 가 남아 있다.

### (1) `dispatchHandoff` 위조를 연 채로 5.5 를 내보낼 것인가 ★ 가장 무거움

13 라운드가 굳힌 것은 **backstop 이지 seam 이 아니다.** `dispatchHandoff` 는 패키지
사설 구조체의 메서드라, 엔진 안의 아무 함수나 자기가 고른 entries 로 수신자를 만들어
봉투를 주조할 수 있다. 4 차 리뷰가 그런 우회 넷을 컴파일해 보였다.

그것들이 주문으로 안 이어지는 이유는 경계가 아니라 `strategy_account_first_leg_authority.go`
의 재유도·identity 대조(`:217`, `:221`–`:222`, `:223`–`:225`)다. 그 다섯 줄이 유일한
방어라는 뜻이다. 진짜 해법(`Admit` 이 조정자만 만들 수 있는 봉인 값을 받는 것)은
이 task 의 파일 소유 밖이라 **6.2 의 차단 선결조건**으로 걸어 두었다.

**결정할 것:** 이 상태로 5.5 를 완료로 볼지, 아니면 6.2 전에 seam 을 먼저 닫을지.

### (2) CI 를 배선할 것인가 — **닫혔다 (2026-09-02), 절반은 배선하고 절반은 이름만 적었다**

배선했다. `.github/workflows/ci.yml` 에 `sdd-checks` job 이 생겼고 `make sdd-check-ci`
하나를 부른다. 그 타깃은 `sdd-check` 에서 **러너로 옮길 수 없는 둘만 뺀** 부분집합이다.

옮길 수 없다는 것은 측정한 것이다 — 외부 도구를 지운 PATH 와 depth-1 클론에서 각각
exit 1: `sdd-doctor` 는 로컬 설치 CLI(rtk·openspec·codegraph·gbrain)를 보고,
`check_index_freshness.py` 는 codegraph 실행 파일과 gitignore 된
`.sdd/index-state.json` 을 본다. 나머지는 같은 환경에서 전부 exit 0 이었다.

이제 CI 가 도는 것: 파이썬 시험 180개(그중 `tools/logic-map` 77개가 열거표 감사 —
`check_analysis.py`·`role_check.py`) + `go test ./tools/logic-map` + 에이전트 부트스트랩
동기화 + 기억 원장 + PM 번호 계약 + `compileall`.

목록이 두 곳으로 갈라지는 것은 시험이 막는다
(`tools/sdd/test_ci_runs_portable_sdd_checks.py`, 뮤테이션 5종으로 반증 확인).
`sdd-check` 에 옮길 수 있는 검사를 직접 한 줄 더하면 실패한다.

**안 배선한 것 — `check_analysis.py --change <id>` 자체.** 그 검사는 번들을 유도 당시
소스에 묶으므로 change 하나의 **완료** 게이트(`make gate` 5/10)에서만 참이다. 측정:
활성 change 31개 중 통과 1개(a112). 나머지 30개는 **AST 소스 해시 stale 15 · 넓어진
수정 집합의 FLM 누락 11 · base-commit 누락/무효 4**. 저장소 전체로 켜면 첫날부터
빨갛고, PR 이 건드린 change 만 골라 켜도 a113~a115 를 잡는 세션이 남의 빚으로
막힌다. 이것은 (5) 의 구조적 빈틈과 같은 뿌리다.

### (3) 커밋·푸시 시점 — **닫혔다 (2026-09-02)**

세 커밋으로 나눴다(1 절). 푸시는 아직 안 했다.

### (4) 외부 a117 스텁의 주인 — **닫혔다 (2026-09-02), 다만 절반만**

`a117-codex-session-handoff-and-gbrain-startup` → `a119-…` 로 renumber 했다.
번호만 바꾸면 `active change has no Story` 가 남으므로 PM 레코드도 함께 만들었다:
`docs/pm/portfolio/stories/STORY-TOS-a119.yaml`(FEAT-TOS-001), `_registry.yaml`,
`FEAT-TOS-001.yaml`, 그리고 재생성한 `docs/pm/generated/` 3 파일.
Story 의 acceptance 는 **그 세션의 proposal.md 에서 그대로 옮겨 적은 것**이고,
`found_by` 에 누가 왜 만들었는지 적었다. 내용의 주인은 그 세션이다.

`make sdd-check` 는 이제 통과한다. **남은 것:** 그 스텁은 `specs/` 델타가 없어
`openspec validate --strict` 를 통과하지 못한다(`make validate --all` 이 1 건
실패). 그 세션이 델타를 채워야 한다. 그리고 그 세션이 스스로 다른 번호로
renumber 하면 디렉터리가 둘로 갈릴 수 있다.

### (6) `refreshOnly` 가 중앙 무결성 판정보다 앞이어야 하는가 (5.6.1 이 올린 것)

`runMarket` 의 판정 순서는 `refreshOnly`(813:4) → `isCentralStrategyIntegrity`
(816:4)다. 그래서 **오늘 생산이 실제로 돌리는 유일한 구성**(두 worker 다
`Effective=false, RefreshesAuthority=true`)에서는 중앙 무결성 오류조차 삼켜진다.

두 권위가 갈린다. `design.md:198` 의 고장표는 "journal/Gateway/fence/owner
integrity fault → 모든 신규 entry fail-closed" 라 하고, 같은 절의 "lane context 와
safety context 를 분리한다"와 spec 의 "lane worker 가 safety loop 를 취소해서는
안 된다 (MUST NOT)"는 반대쪽을 가리킨다.

순서를 뒤집으면 전략 평가 하나가 `Run` 을 반환시키고 Runtime 이 **모든 loop 를
취소한다** — fill/exit/reconcile 포함. **엔진이 서면 손절을 놓는 주체가 없다.**
그래서 5.6.1 은 순서를 바꾸지 않고 값으로 고정했다. 진입만 닫는 수단은 이미 있다
(`execgw.EntryGate.Block`; reconcile·alert·filldetect·flatten 이 전부 그것을 쓴다).

**결정할 것:** fail-closed 의 수단을 프로세스 정지로 볼지, EntryGate 로 볼지.
후자라면 5.6.2 가 그 배선을 가져간다.

### (7) 여덟을 켤 권한을 어디에 둘 것인가 (5.1.2.1 이 올린 것)

5.1.2 의 나머지 절반(여덟이 진입의 **관문**이 되는 것)은 이 로트에서 할 수 없었다.
동결 골든이 여덟 서술자를 전부 `effective: OFF` 로 얼렸고, 스펙은 legacy 3-family
승인을 4-family activation 으로 승격하는 것을 MUST NOT 으로 금지한다. 잠든
`FamilyWorker.Run` 은 아무것도 보기 전에 DORMANT 를 돌려주므로, 오늘 여덟을
관문으로 세우면 **생산 진입이 0** 이 된다.

그래서 관문 교체는 5.1.2.2 로 나갔고, 그 태스크는 **서명된 활성화 매니페스트**를
전제한다. 매니페스트가 없으면 5.1.2.2 는 영영 열리지 않는다.

**2026-09-03 추가:** 이 결정에 **5.2.2 도 걸린다.** 시장 준비 상태에서 단일 제안
접기를 들어내면 소유자 범위가 둘인 시장이 오늘은 안 내던 주문을 내기 시작한다. 즉
"토글 OFF 는 upstream 동작과 동일" 을 지키려면 그 편집도 같은 매니페스트 뒤여야 한다.
5.2 의 나머지 절(원격 수집을 잠금 밖으로)은 이 결정과 무관하므로 5.2.1 로 갈라 랜딩했다.

**결정할 것:** 8 절(서명 매니페스트)을 5.1.2.2 보다 먼저 세울지, 아니면 5.3.3 을
먼저 하고 8 절과 5.1.2.2·5.2.2 를 한 로트로 묶을지. 후자를 고르면 여덟은 그때까지
"서 있지만 아무것도 결정하지 않는" 상태로 남는다.

### (5) 닫지 않고 이름만 적은 둘을 받아들일 것인가

- `revision: base` 번들 **15 개**가 소스 해시에 묶이지 않는다(13 개는 이 change 의
  `base-commit.txt` 와도 불일치). "133/133 이 열거한다"는 **표 모양의 참**이지 좌표가
  옳다는 뜻이 아니다. 독립 `go/parser` 재유도로 확인된 것은 118 개다.
- `internal/strategyflow` 에 exported 표면 census 가 없다. 이제 봉인된 `Result` 를
  만드는 함수가 태그 아래 **둘**이다(`AcceptedResultForAuthorityTest`,
  `ResultWithRestatedStopProvenanceForTest`). `strategyhandoff` 가 4 차 전에 있던 상태와 같다.

## 3. 다음 세션이 할 일 — 이 순서

1. ~~위 (3)·(4) 에 대한 사람의 답을 먼저 받는다.~~ **끝났다** — 3 커밋 + a119 renumber.
2. ~~커밋 → `make sdd-sync` 재실행 → `make sdd-check`.~~ **끝났다.** fingerprint 는
   sync 시점 HEAD 를 기록하므로, 이 뒤로 커밋이 붙으면 다시 `make sdd-sync` 를 돌린다.
3. 그다음 열려 있는 태스크로 간다. 5 절에서 남은 것은 **5.2.2 · 5.6.2**, 8 절에
   **8.7.2**, 그다음이 6 절이다.
   - **8.7.1 + 5.1.2.2 가 닫혔다 (2026-09-04).** 5.2.2 를 여는 사람이 볼 것 넷:
     (a) 관문은 이미 조정 앞에 서 있으므로 5.2.2 가 옮길 것은 관문이 아니라
     **상한**이다 — `strategyhandoff.Capacity = 1` · `deliverable = 1` · `Single()`
     의 서명을 함께 바꿔야 하고(그 파일의 주석이 그 로트를 지목한다),
     `buildProductionStrategyMarketWorker` 의 `p.dispatchHandoff().Single()` 한 줄이
     시장의 준비 상태를 정한다; (b) 상한을 들어내면 소유자 범위가 둘인 시장이
     **거래를 시작한다** — 진입 발행이 느는 방향이라 사람 승인이 필요하고, 지금은
     활성화가 그 승인의 자리다(활성화된 시장에서만 관문이 판정한다); (c) 기존 시장
     단위 경로를 **지우지 않았다** — 활성화가 없을 때의 갈래로 남아 있고 그것이
     토글 OFF = upstream 을 지킨다. 지우는 것은 5.2.2 와 6 절이다; (d) 다섯 digest 중
     둘(위험 번들 · ProtectionReady)은 `buildProductionStrategyMarketWorker` 에서
     결속하므로 그 함수를 편집할 때 그 두 줄을 같이 본다.
   - **활성화 매니페스트를 실제로 만드는 절차는 아직 없다.** 8.7.2 가 운영 자세
     (없거나 만료·폐기됐을 때의 rollback)를 가져가고, 서명 도구·키 관리·배포 순서는
     그 태스크를 여는 사람이 정한다. 오늘은 파일이 없으므로 여덟이 전부 OFF 다.
   - **5.3.3 이 닫혔다 (2026-09-03).** 레인 잠금이 원장에 남는다(schema v32, append-only
     두 테이블). 레인은 열린 채로 태어나지 않고 **기록에서** 태어나며, 복구는 성공한
     사이클이 아니라 **엄격히 더 큰 서명 활성화 세대**(`scheduler.Activation.Generation()`,
     ed25519 매니페스트)를 요구한다. 그 판정은 Go 가 아니라 SQLite 트리거가 한다 —
     첫 판본은 판정이 둘이었고 반증이 "각자가 상대의 시험을 통과시킨다"를 보여 줘서
     Go 쪽을 지웠다. **이 로트가 없앤 운영 수단 하나**: 잠긴 레인을 재시작으로 열 수
     없다. 여는 것은 서명 매니페스트뿐이다. 여덟이 DORMANT 라 오늘 쓰이는 행은 0 이고
     그것도 시험이 값으로 확인한다.
   - **원장 스키마가 31 → 32 로 올랐다.** 이 마이그레이션이 돈 DB 를 구버전 바이너리로
     되돌리면 엔진이 `ErrSchemaTooNew` 로 **거절**한다(조용히 오해하지 않는다). a112 는
     8.6 기준 배포 BLOCKED 이고 병합·배포 순서는 사람이 정한다. main 과 합칠 때
     SchemaVersion 을 먼저 대조할 것.
   - **5.2 는 둘로 갈렸고 5.2.1 이 닫혔다 (2026-09-03).**
     `refreshPairedStrategyEntryProductionAssembly` 가 더 이상 공유 잠금을 들고
     원격을 타지 않는다. 잠금은 상태 전이만 지키고(`joinStrategyRefreshWave` /
     `publishStrategyRefreshWave`), 파도는 잠금 밖에서 돌며, 기다리는 시장은
     채널에서 기다려 자기 주기가 취소되면 빠져나온다. **1초 창의 의미는 그대로**
     (파도의 시작 시각 기준) — 합치기는 창이 아니라 파도가 한다. 5.2.2 를 여는
     사람이 볼 것 셋: (a) 시장 준비 상태의 단일 제안 접기는
     `buildProductionStrategyMarketWorker` 의 `p.dispatchHandoff().Single()` 한
     줄이고, 그것을 들어내면 소유자 범위가 둘인 시장이 **거래를 시작한다** — 진입
     발행이 느는 방향이라 네-가족 OFF 동안은 토글 불변식에 어긋난다; (b) 그래서
     5.2.2 는 5.1.2.2 와 **같은** 사람 결정 (7) 에 걸린다; (c) `internal/app/engine`
     을 통째로 `-race` 에 넣으면 게이트가 14분 46초 늘어난다(측정) — 그래서
     `RACE_ENGINE_TESTS` 는 이름 목록이고, 그 목록의 완전성은 python 가드가 잰다.
   - **5.1.2 는 둘로 갈렸고 5.1.2.1 이 닫혔다 (2026-09-03).** 여덟 레인이
     `Context.productionStrategyLanes` 아래 프로세스 수명으로 서고,
     `runProductionStrategyMarketCycle` 이 새로 고침 뒤·공유 mutex 밖에서 이
     시장의 넷을 돌린다. 레인 안에서 도는 것은 `strategyFamilyLaneStep` 하나이고
     인자가 `*strategyworker.Lane` 뿐이다. **여덟은 전부 DORMANT 라 생산 동작은
     그대로다.** 5.1.2.2 를 여는 사람이 볼 것 셋: (a) 레인이 보는 것은 조정을
     **지난** 제안이다 — 관문을 옮기려면 `coordinateMarketProposals` 의 제출
     루프 앞으로 레인을 옮겨야 한다; (b) 레인 고장은 아직 감독자 fault 스트림에
     안 간다 — 보내는 순간 5.6.1 이 찾은 `cap(faults)==2==시장 수` 등식이 깨진다;
     (c) 켜진 worker 는 `strategyworker` 밖에서 만들 수 없다(`newWorker` 비공개).
     그 문을 여는 것이 곧 매니페스트 작업이다.
   - **5.1.1 은 닫혔다 (2026-09-02, `4724c3d0`).** `internal/strategyworker` 에 여덟
     `FamilyWorker` 와 폐포 증명이 섰다 — 5.5 가 미룬 "worker 의 `Cycle` 이 broker
     mutation 에 못 닿는다"가 **새 타입에 대해서는** 시험이 되었다. 상세는
     tasks.md 5.1.1 본문과 review.md 의 2026-09-02 절.
   - **닫힌 것은 절반이고, 나머지 절반은 새 태스크 5.1.2 다.** 원래 5.1.1 문장의
     "rather than" 앞이 5.1.1, 뒤가 5.1.2 다. 나누면서 빠진 의무는 없다 — 원문의
     모든 절이 둘 중 하나의 소유이고 5.1.2 는 열려 있다.
   - **5.1 은 닫혔고 5.3 은 세 조각이 되었다 (2026-09-02).** 5.3.1(고장 카운터·
     backoff·entry latch)과 5.3.2(단일 비행·카덴스·마감 시한)가 랜딩했고,
     5.3.3(**durable** latch 와 recovery conditions)만 열려 있다. 5.3.3 을 여기서
     만들 수 없는 이유는 `internal/strategyworker` 가 **쓰기 능력이 없다는 것이
     증명된** 패키지라서다 — 지속 기록에는 쓰는 쪽이 필요하고, 그것을 여기 넣으면
     5.1.1 이 세운 성질이 사라진다. `design.md:223` 이 그 일을
     `internal/app/engine` 에 준다. 오늘의 엔진도 지속 latch 가 없으므로 5.3.3 은
     **옮겨 적기가 아니라 새 동작**이고, 영수증은 현재 코드가 아니라 서명 매니페스트다.
   - **5.3.2 가 남긴 사람 결정 하나: deadline 이 abnormal 인가.** 엔진(`:897`)은
     그렇다 하고 `design.md:198` 은 아니라 한다. 생산 임계값이 1 이라 오늘은 결과가
     같아서 엔진을 따랐다. **임계값을 1 보다 올리기 전에** 사람이 정해야 한다.
   - **5.7 이 가져갈 빈칸이 측정으로 일곱 개 나왔다.** 엔진 스위트가 한 번도 돌리지
     않는 블록이다(잠금 실패·재시작 대기 실패·사이클 건너뛰기·마감시한과 취소의 경합).
     좌표는 review.md 의 5.3.2 절과 두 FLM 번들의 branch-test-map 에 있었다.
     **일곱은 2026-09-03 에 5.6.1 이 전부 채웠다** — 전부 `count=1` 이고 각 번들의
     branch-test-map 이 어느 시험이 채웠는지를 적는다.
   - **5.7 은 닫혔다 (2026-09-02).** 여덟 레인·두 조정자를 함께 세우는 리허설과,
     그보다 중요한 것 하나: **이 저장소는 `-race` 를 한 번도 돌린 적이 없었다.**
     이제 `make test-race`(일곱 패키지, 11.9초)를 게이트 9/11 단계와 CI 가 돌고,
     배선이 빠지면 `tools/sdd/test_race_detector_actually_runs.py` 가 실패한다.
     **남은 34개 동시성 패키지는 여전히 검출기 밖이다**(`internal/journal`,
     `internal/app/engine` 포함; 전체 `-race` 는 10분에 끊겼다). 새 change 가
     동시성 패키지를 만들면 `RACE_PACKAGES` 에 더할 것.
   - **5.7 이 5.3.2 의 구멍을 찾아 고쳤다.** `Step` 이 오류를 못 돌려줘서 설계
     고장표의 "보통 오류" 줄에 닿는 입력이 존재하지 않았다 — 즉 `FailureThreshold`
     가 죽은 값이었다. 이제 `Step` 은 `(Cycle, error)` 다. 5.1.2 가 실제 사이클을
     꽂을 때 이 서명을 쓴다.
   - **`design.md:255` 의 "dispatch handoff 를 spy 로 막아라"는 엔진에서만 할 수 있다.**
     `strategyhandoff` 의 importer census 가 그 seam 을 들여올 수 있는 패키지를
     엔진 하나로 못 박는다. 5.1.2/5.2 를 여는 사람이 그 spy 를 세운다.
   - **아직 참인 것:** 생산은 `StrategyMarketWorker` 를 돌리고 그 `Cycle` 은
     `*Context` 클로저라 Journal/Gateway/Guardian 을 들고 있다. AST 산출물이
     `c.Journal.CurrentPositionCampaignCAS`(`:446`)와 `fresh.dispatch.dispatch`(`:453`)를
     열거한다. 그 교체가 5.1.2 이고, 5.2 는 같은 변경의 감독 쪽이다. 그전에 handoff 를
     spy 로 막는 것이 설계 순서다(`design.md:255`).
   - **여덟은 지금 아무도 안 부른다.** 골든 대조 시험이 그것들을 동결된
     `descriptors` 에 묶어 두므로 조용히 계약에서 어긋나지는 않지만, 5.1.2 전까지
     생산 호출자는 0 이다.
   - **5.1 은 닫혔다 (2026-09-02).** 남아 있던 넷째(서버 소유 versioned worker
     runtime policy)가 `strategyworker.RuntimePolicy` 로 섰다. 소비자 없는
     매니페스트가 되지 않도록 5.3 의 고장 절반(5.3.1)을 같은 로트에 붙였다 —
     레인이 임계값과 backoff 두 값을 실제로 쓴다.
   - **정책의 여섯 값은 고른 수가 아니라 엔진에서 읽은 수다.** cadence 의 영수증은
     상수 이름이 아니라 `PollInterval` 대입 세 자리이고, queue depth 의 영수증은
     **생산이 그 값을 한 번도 지정하지 않는다는 것**이다. 두 시험 모두 엔진
     디렉터리의 모든 비시험 파일을 훑는다(파일 하나를 고르면 완전성이 고른
     사람의 것이 된다). 임계값 1 은 완화 거부다 — 오늘 엔진은 첫 실패에 잠근다.
   - **5.3 도 둘로 갈랐다.** 5.3.1(카운터·backoff·latch)은 닫혔고, 5.3.2 가
     single-flight cadence·monotonic deadline·**durable** latch/recovery 를 갖는다.
     이 로트의 latch 는 프로세스 메모리에 있어 재시작을 못 넘긴다.
   - **5.6 은 둘로 갈렸고 5.6.1 이 닫혔다 (2026-09-03).** 5.6.1 은 오늘 서 있는
     런타임의 고장 범위를 재는 계기이고, 5.6.2 는 교체 뒤 같은 세 절을 여덟 lane
     위에서 다시 증명한다. 재서 나온 것 셋을 5.1.2 를 여는 사람이 반드시 볼 것:
     (a) 전략 고장이 엔진을 세우는 경로는 넷뿐이고 넷 다 **감독자 자신의 장부가
     깨진 경우**다 — 평가 실패는 아니다; (b) 오늘 생산이 도는 구성은
     `refreshOnly` 갈래라 사이클 오류가 잠금 없이 삼켜진다 — 교체하면 그 성질이
     사라진다; (c) **fault 스트림 용량 2 = 시장 수 2** 라는 균형이 어디에도 적혀
     있지 않았고, 여덟으로 늘리면서 그것을 함께 옮기지 않으면 **세 번째 lane 의
     잠금이 엔진과 safety loop 를 함께 세운다.** 지금은
     `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch` 가 막는다.
   - **`StrategyCentralIntegrityFailure` 는 생산 호출자가 0 이다** —
     `a112_central_integrity_census_test.go` 가 패키지 전체 열거로 확인한다.
     새 자리를 여는 것은 "엔진을 세울 새 이유를 만드는 것"이고 그 시험이 실패한다.
4. **6.2 를 여는 사람은 tasks.md 6.2 본문을 먼저 읽는다.** 아래 4 절이 그 요약이다.

## 4. 6.2 를 여는 사람에게

6.2 는 `strategy_account_first_leg_authority.go` 의 다섯 줄을 지우거나 바꾸게 된다.
**세 가지가 그 자리를 잡고 있고, 셋 다 초록으로 유지해야 한다.**

| 무엇을 증명 | 어디 |
|---|---|
| 이 가드가 돌고 거절한다 | `strategy_first_leg_identity_backstop_test.go` — 시험 함수 넷(하위 여섯) |
| 이 가드가 봉인된 identity 를 **그대로** 비교한다 | `strategy_first_leg_backstop_shape_test.go` — 구조 단언 하나 |
| 그 identity 가 **모든 필드**를 담는다 | `strategyflow/execution_terms_identity_fields_test.go` (32) + `weeklyvaluelane/execution_policy_identity_fields_test.go` (8) |

**구조 단언은 6.2 첫날에 정당하게 걸린다.** 그것이 `proposal.entries[0].authority` 를
못 박고 있는데, 6.2 의 일이 한 종목에 네 가족을 받는 것이라 `entries[0]` 이 정당하게
틀려진다. 그때:

- **기대 문자열만 고쳐서 통과시키지 말 것.** 실패 메시지에 그 이유를 적어 두었다.
- 여러 항목 중 하나를 고르게 바꾼다면 **그 선택이 `accepted` 에 기대면 안 된다.**
  accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와 비교하면 가드가 **자기 참조**가
  되고, 행동 시험도 census 도 전부 초록으로 남는다(12 차 리뷰가 재현).
- 오늘 그 편집이 무해한 이유는 `:217` 의 `len(entries) != 1` 이 항목을 하나로 묶기
  때문인데, **6.2 가 바로 그것을 바꾼다.**

## 5. 검증 명령 (그대로 복사)

```
make lint                                   # PASS
make test                                   # 98 ok / 0 FAIL
make test-seams                             # 99 ok / 0 FAIL
python3 -m unittest discover -s tools/logic-map -p 'test_*.py'   # 77 OK
python3 tools/logic-map/check_analysis.py --change a112-run-four-strategy-families-independently
openspec validate a112-run-four-strategy-families-independently --strict
$(go env GOROOT)/bin/gofmt -l internal/ tools/    # 출력 없음
make sdd-sync && make sdd-check              # PASS (a119 renumber 뒤)
make sdd-check-ci                            # CI 가 도는 부분집합 — 로컬 도구 없이도 PASS
```

`make gate CHANGE=…` 는 **not-applicable** 이다 — change **완료** 게이트라 미완료
태스크가 있으면 2/10 에서 멈춘다. 진행 중 로트의 판정에 쓰지 말 것.

## 6. 함정 (여기서 실제로 밟은 것들)

1. **`go test -overlay` 는 소스를 읽는 시험에 안 보인다.** 구조 단언은
   `parser.ParseFile` 로 디스크에서 읽으므로 overlay 뮤테이션이 전부 GREEN 으로 나온다.
   그것은 "가드가 강하다"가 아니라 **"안 쟀다"** 이다. 파일을 진짜로 바꾸고 해시
   백업에서 복원할 것. `git checkout` 은 커밋 안 된 이 로트 전체를 지우므로 금지.
2. **뮤테이션 범위에 `_test.go` 를 넣지 말 것.** 패키지 전체 리네임을 돌렸더니 시험
   파일까지 함께 바뀌어, 뮤테이션이 자기를 잡아야 할 시험을 **수리**하고 초록을 냈다.
3. **커밋하면 `make sdd-sync` 를 다시 돌린다.** fingerprint 는 sync 시점 HEAD 를
   기록한다.
4. **`rtk` 가 출력을 요약한다.** 정확한 개수를 세려면 `rtk proxy <cmd>`.
   `rg` 는 PATH 에 없다(`grep -rn` 사용), `gofmt` 는 `$(go env GOROOT)/bin/gofmt`.
5. **`make sdd-check` 는 이제 통과해야 한다.** 실패 줄이 생기면 그것은 이 로트나
   병행 세션의 문제다. `make validate --all` 은 a119 스텁의 델타 부재로 1 건 실패가
   남아 있는데, 그것은 그 세션의 것이다.
