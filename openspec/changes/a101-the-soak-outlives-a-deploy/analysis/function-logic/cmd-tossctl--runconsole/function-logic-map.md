# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (source SHA-256 `b5f9cd6435d421c2…`, 분기 44 / return 18, L211-516 — **편집 후 재생성**(편집 전 41/17))
- Risk scan: `risk-pattern-report.md`
- 작성 사유: a101 tasks 2.3·3.1 — 콘솔 기동에 soak autostart 호출을 넣고 `RestartSoak` seam을
  감싼다. **기존 함수의 내부를 편집하므로 편집 전에 만든다.** 산출물이 attestation을 갱신하는
  프로세스이므로 인증 경로이고 High-risk다(§0-5).

## 이 함수가 하는 일

콘솔 프로세스의 **조립 전체**다. 경로를 풀고, 자격증명·기록·attestation의 위치를 정하고,
브로커 seam·엔진 control plane·업데이트 seam을 만들고, 마지막에
`console.ListenAndServe(ctx, console.Options{...})`에 전부 넘긴다. 분기 44개 중 대부분은
**「선택 seam이 nil이면 그 화면만 조회 전용으로 뜬다」**는 형태의 열화(degradation) 판정이고,
나머지는 초기 오류 반환이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root.configDir` | 비어 있거나 존재하는 디렉터리 | CLI 플래그 | 경로 해석 실패는 초기 return |
| `root.sessionFile` | 임의 | CLI 플래그 | — |
| `opts.port` 등 | 콘솔 플래그 | CLI | 검증 실패는 초기 return |
| 엔진 control plane descriptor | 있거나 없음 | `engineDir`의 파일 | **없으면 화면이 조회 전용으로 뜬다. 기동을 막지 않는다** |
| `engineBoot` seam | nil 가능 | `consoleEngineBootSeam(root)` | nil이면 autostart 판정이 "배선 없음" |
| `engineBootNote` | 문자열 | `runConfiguredEngineAutostart` | **비어 있지 않으면 stderr에 한 줄. 그 뒤 조립은 계속된다** |

**이 함수 전체를 관통하는 불변식 하나**: 선택 기능의 부재는 **출력 한 줄**이고 기동 실패가
아니다. 엔진 control plane에 붙지 못해도(`:359`, `:375`), 성과 DB가 없어도, 콘솔은 뜬다.
a101이 추가하는 판정은 **그 규칙을 따라야 한다** — 조회 전용 서베이를 못 세운 것이 운영자
화면을 없앨 이유가 될 수 없다.

## Branches and early returns

44개 분기의 개별 열거는 `branch-test-map.md`에 측정값과 함께 있다. 여기서는 **a101이 닿는
구간**만 적는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B29 (`:338`) | `engineBoot != nil` | `engineBootLoad` 대입 | — | seam이 없으면 load도 없다 |
| B30 (`:345`) | `engineBootNote != ""` | stderr 한 줄 | — | **a101의 삽입 지점 바로 위** |

a101은 **B30 블록 다음에 같은 형태의 블록을 놓았다.** 실제로 늘어난 분기는 **셋**이다 —
seam nil 검사 둘(`soakBoot != nil`, closure 안 같은 검사)과 출력 판정 하나(`note != ""`).
셋 다 nil 검사·출력이고, 그것들이 결정하는 내용은 전부 측정되는 함수 쪽에 있다
(`runConfiguredSoakAutostart` 100.0%, `rememberSoakApproval` 100.0%).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleEngineBootSeam(root)` (`:336`) | 엔진 autostart 승인 읽기 seam | nil 가능 | ast.json |
| `runConfiguredEngineAutostart` (`:341`) | 기동 시 엔진 자동 시작 판정 | **에러를 반환하지 않는다.** 결과는 사람이 읽는 한 줄 | `engineautostart.go:57` |
| `startEngine(root)` (`:343`) | 실제 기동 | 실패는 문자열로 돌아온다 | ast.json |
| `restartSoak(root, soakRecord, …)` (`:460`) | 대시보드의 soak 재시작 seam | `(string, error)` | `soakproc.go:142` |
| `console.ListenAndServe` (`:384`) | 조립된 Options로 서빙 | 이 함수의 최종 return | ast.json |

live binding: `RestartSoak` closure(`:459-461`)가 콘솔이 서베이를 세우는 **유일한 경로**다.
a101은 새 경로를 만들지 않고 이것을 재사용한다.

## State mutations and fallbacks

- 이 함수는 **프로세스 밖 상태를 거의 바꾸지 않는다.** 예외가 정확히 둘이다:
  `runConfiguredEngineAutostart`가 엔진 프로세스를 띄울 수 있고, `RestartSoak` closure가
  호출되면 서베이 프로세스를 띄운다. a101은 **두 번째를 기동 시점에도 부른다.**
- fallback은 전부 "그 화면만 죽인다" 형태다. 붙지 못한 control plane은 조회 전용 화면이 되고
  (`:359`, `:375`), 없는 seam은 nil로 남아 해당 버튼이 `NotImplemented`를 답한다.
- **config 파일 쓰기는 이 함수에 없다.** a101의 승인 영속은 `RestartSoak` closure 안에서
  일어나며, 그 closure는 서베이가 실제로 섰을 때만 실행되는 경로다.

## 편집이 건드려선 안 되는 것

1. **엔진 autostart보다 먼저 서베이를 세우지 않는다.** 둘은 같은 계좌의 rate budget을 쓰고,
   엔진의 기동 인터록이 attestation을 읽는 쪽이다. 순서를 뒤집으면 서베이의 첫 사이클이
   엔진 기동과 경합한다.
2. **서베이 기동 실패로 return하지 않는다.** 이 함수의 관통 불변식(선택 기능의 부재 = 출력
   한 줄)을 깨는 유일한 방법이 그것이다.
3. **`RestartSoak` closure의 반환 계약을 바꾸지 않는다.** 대시보드는 그 문자열을 그대로
   출력하고, 에러는 재시작 실패로 표시된다. 승인 기록 실패를 에러로 승격시키면 **이미 선
   서베이를 운영자가 다시 죽인다.**
4. **`soakRecord`·`root`의 해석을 다시 하지 않는다.** 프로필 해석은 이미 이 함수 위쪽에서
   한 번 일어났고, autostart는 같은 값을 받아 써야 한다(a060의 프로필 격리).

## Safety conclusion

- **Safe edit boundary**: B30 블록 **다음**에 출력 블록 추가, 그리고 재시작 closure를
  같은 시그니처로 감싸기. **편집 완료** — 기존 41개 분기 중 어느 것의 조건도 바뀌지 않았고,
  늘어난 3개는 전부 nil 검사와 출력이다.
- **High-risk impact**: **yes** — 이 배선의 산출물이 automation gate가 읽는 attestation을
  갱신하는 프로세스다. 다만 방향은 보수적이다: 이 편집이 잘못돼도 **서베이가 안 서는 것**이
  최악이고, 그것이 바로 지금의 동작이다.
