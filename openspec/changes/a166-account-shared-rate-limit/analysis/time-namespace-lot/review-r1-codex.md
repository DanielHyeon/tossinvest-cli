`aaad76d7e75368c7d3188217e1107c92e0ea1f42`, 요청 브랜치·깨끗한 작업 트리 확인. 읽기 전용 검토이며 시험·네트워크·docker·브로커·운영 접근은 하지 않았다. 아래 문서명은 `analysis/time-namespace-lot/` 기준이다.

**[P2] §5의 “새로 받아들이는 정상 밖 입력 1건”은 미지원 환경을 누락한다 — Q1·Q2·Q5**

근거: `investigation.md:156–177`, `contract-delta.md:14–19,61`. 원 계약 `analysis/poll-clock-contract.md:13`은 procfs bind 대체와 런타임 thread별 namespace 조작을 명시적으로 제외한다.

`Snapshot`은 `/proc`의 filesystem 종류만 검사하고, 읽은 boot UUID의 실제 부팅 계보를 인증하지 않는다(`internal/clock/bootstamp_linux.go:33–40,132–138,171–179`).

반례: 런타임이 서로 다른 실제 boot에 동일한 가상 boot UUID를 제공하고 저장 상태를 복원한다. saved는 `uptime=100s, ns=A`, now는 `131s, ns=B`, duration은 `30s`라면:

- 기존 코드: namespace 차이로 거절.
- 권장안: 역행도 UUID 차이도 없어 만료 처리.
- 양쪽 offset이 0이어도 서로 다른 실제 boot의 uptime 차감은 정당화되지 않는다.

**원 계약 밖 반례이며 권장안의 지원 범위 안 결함은 아니다.** 표에는 이를 “새로 통과 가능, 코드 가드 없음, trusted kernel procfs 전제로 제외”라고 추가해야 한다. 가상 boot ID가 실제 boot를 일대일로 식별하지 않는 환경까지 지원하는 것은 **범위 확장**이다.

**[P2] 표본 앞뒤 검사를 모든 스레드·중간 변경의 검증으로 과장한다 — Q1·Q2·Q5**

근거: `investigation.md:65–66,158–159,176–177`, `contract-delta.md:61`; `internal/clock/bootstamp_linux.go:18–20,140–168`.

`/proc/self`는 호출 스레드 전용 경로가 아니다. 검사 대상은 process leader 쪽 namespace 정보이고, BOOTTIME syscall은 호출 task의 `nsproxy->time_ns`를 사용한다. 로컬 Linux 헤더의 `include/linux/time_namespace.h:80–84`에서도 이 구분이 확인된다. 따라서 “새 ns도 offset 검사를 통과해야 한다”만으로 스레드별 조작 가능성을 배제한 논증은 불충분하다.

또한 앞뒤 값 비교는 중간의 `A → B → A` 변경 자체를 검출하지 않는다. 이것이 실제 지원 커널에서 실행 가능한지는 별도 문제이며, 이번 검토에서는 `timens_install` 구현을 확인하지 못했다. **실행 가능한 다중 스레드 공격으로 확정하지 않는다.**

수정 제안: 원 계약의 “실행 중 namespace 교체·thread별 조작 없음”을 그대로 전제하고, 문구를 “관측한 앞뒤 namespace 불일치를 거절한다”로 제한한다. 기존 가드가 모든 중간 변경을 잡는다고 쓰지 않는다.

참고로 현재 provider의 `unix.ClockGettime`은 설치된 `x/sys v0.46.0/unix/zsyscall_linux.go:585–590`에서 직접 `SYS_CLOCK_GETTIME`을 호출한다. **이 호출을 Go `time.Now`의 vDSO 경로와 혼동하면 안 된다.**

**[P2] §4.1의 429 실패 순서에 hint 갱신 진입 조건이 빠졌다 — Q3**

근거: `investigation.md:138–142`; `internal/accountbudget/observation_linux.go:489–491,506–531`.

“429이고 기존 hint가 있으면 L531에서 실패”는 조건이 부족하다. L531까지 가려면 다음이 모두 필요하다.

- `result == ObservationRateLimited`
- `pending.Kind == ObservationPoll`
- 정규화 후 `hint != nil && *hint > 0`
- 해당 slot에 기존 hint 존재

예컨대 기존 cross-namespace hint가 있어도 **새 429의 hint가 nil·0·invalid**이면 L531을 건너뛰고 사건을 기록할 수 있다. §4.1-3에 이 조건을 명시해야 한다.

prune도 단순히 “256개가 찼을 때”만으로 실행되지 않는다. `Begin`에서는 이전 poll pending의 key가 없을 때, `Finish`에서는 새 유효 poll hint의 key가 없을 때 추가로 진입한다(`observation_linux.go:440–450,514–518`).

**[P2] “시험 3개만 영향”과 execgw 통과 주장의 증거 범위가 다르다 — Q4**

근거: `investigation.md:189–197`; `selected-tests-baseline-vs-overlay.log:1–20`.

로그가 입증하는 것은 **선택한 네 패키지 실행에서 실패한 시험 함수 3개**다. `execgw -run 'A166|EnginePoll'`의 무태그·`tossos_testseams` 실행 명령과 결과는 해당 로그에 없다. 전체 기존 시험에서 “통과 이유가 바뀌는 함수가 정확히 3개뿐”이라는 전수 주장도 이 로그만으로는 성립하지 않는다.

세 실패는 소스와 일치한다.

- `internal/clock/bootstamp_test.go:48–49`
- 같은 파일 `66–67`
- 같은 파일 `129–130`

수정 제안: “선택 실행에서 실패 전환된 함수 3개”로 표현하고, execgw는 원 실행 로그를 추가하거나 미입증으로 표시한다. 첫 번째·세 번째 시험에 함께 있는 uptime 역행 단언은 보존해야 한다.

**기존 함수 수정 기준 FLM 4번들 계산은 맞다.** 생산 함수 1개와 위 시험 함수 3개를 수정한다는 계획에 한정된다. 새 시험 함수는 기존 함수 수정 번들에 포함되지 않는다(`docs/WORKFLOW.md:164–169,197–201`).

**[P3] AST 함수 수는 17개가 아니라 18개다 — Q3**

근거: `investigation.md:56`; `census.txt:1–71`.

census와 `ast/`에는 18개 함수가 있다. 18개 AST의 소스 SHA-256은 모두 현재 파일과 일치했다. `17 함수`를 `18 함수`로 고치면 된다.

나머지 판정:

- **Q1:** 신뢰된 동일 kernel boot, 실제 채취 namespace의 offset 0, 실행 중 namespace 조작 없음이라는 원 계약 안에서 산술 반례 **없음**. 중첩 namespace의 offset은 부모 대비 누적값이 아니라 initial namespace 기준이며, suspend는 BOOTTIME에 포함된다. 두 요소 모두 이 전제에서 비교를 깨지 않는다.
- **Q2:** 위 미지원 환경 누락을 제외하고, 지원 범위 안에서 가드 없이 잘못 수용되는 새 입력은 찾지 못했다.
- **Q3:** namespace 비교 4곳, B8의 선행 검사 중복, 저장 필드 3곳, read→validate 경로는 코드와 일치한다.
- **Q5:** 계약 개정에 원 지원 범위를 유지한다는 조건을 명시하면 신원 비교 제거 방향은 타당하다.

기존 AST·census를 현재 소스와 대조했으며 새 AST·CodeGraph·시험은 실행하지 않았다. ai-memory MCP는 없었고 CLI 조회는 네트워크 금지 범위 때문에 진행하지 않았다. `impeccable: N/A — backend/documentation-only`.

권장안 가 채택 가능