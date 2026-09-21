# Function Logic Map tools

기존 Go 함수의 내부 로직을 바꾸기 전에 다음 증거 묶음을 만든다.

```bash
python3 tools/logic-map/scaffold_analysis.py \
  --change <change-id> --file internal/pkg/file.go --func Receiver.Method

go run ./tools/logic-map \
  --file internal/pkg/file.go --func Receiver.Method \
  > openspec/changes/<change-id>/analysis/function-logic/<target>/ast.json

python3 tools/logic-map/risk_pattern_report.py internal/pkg/file.go \
  --output openspec/changes/<change-id>/analysis/function-logic/<target>/risk-pattern-report.md

python3 tools/logic-map/check_analysis.py --change <change-id>
```

`ast.json`은 현재 source SHA-256을 포함한 구조 증거다. Markdown scaffold는 사람이 CodeGraph 호출 관계,
불변식, 실패 경로와 테스트를 채우는 분석 문서다. `TODO`, 빈/오래된 AST, source·함수 불일치,
수정된 기존 함수의 누락은 gate에서 거절된다. CI의 `SDD_BASE_REF`는 persisted
`base-commit.txt`와 같은 commit으로 resolve되어야 한다.
ast-grep 발견은 자동 결함 판정이 아니다.

## 요구 git 버전

`check_analysis.py` 는 blob 을 `git cat-file --batch -Z` 로 읽는다. `-Z` 는 입력과 출력을
**둘 다** NUL 로 끊어 개행이 든 경로와 tree 의 raw NUL 을 견딘다. 그 옵션은 **git 2.42**
부터다(정본 값은 `check_analysis.GIT_BATCH_MINIMUM`).

더 낮은 git 에서는 blob 읽기가 실패하고, 게이트는 그것을 **결함으로** 멈추며 git 이 한
말을 그대로 전한다. 착지 기록을 읽는 자리가 모든 change 에 있으므로 **모든 change** 가 멈춘다
(문서만 고친 change 도). 예:

```text
[logic-map] cannot derive modified Go functions: cannot read blobs at 1a2b3c4d5e6f: `git cat-file --batch -Z` failed (rc 129: error: unknown switch `Z') — `-Z` needs git 2.42 or newer
```

(`at` 뒤는 게이트가 시작할 때 푼 `HEAD` 의 sha 앞 12자리다 — 2026-09-20 `-Z` 를 모르는 git 을 흉내 내어
실제로 찍어 본 줄에서 sha 만 바꿨다.)

버전 조언(`needs git 2.42`)은 git 이 **옵션을 모른다**(rc 129)고 할 때만 붙는다 — 저장소가 아닌
루트(rc 128) 같은 다른 실패에는 git 의 말만 나간다. 영수증은 git 자신의 릴리스 노트
`RelNotes/2.42.0.txt` 다.

예전(2026-09-18 까지)에는 실패를 "파일 없음"과 같게 다뤄서 **사유를 오진했고**(저자의 증거를
탓했다), 가드 하나는 그 혼동 때문에 **편집 전 커밋을 착지로 기록했다**. 2026-09-19 gstack 리뷰가
찾았고 a122 task 7.5.1 이 고쳤다 — "git 이 못 돌았다"와 "그 커밋에 그 파일이 없다"는 이제 다른 답이다.

## 착지 지점 — `landed-commit.txt`

`check_analysis.py` 의 5단계는 `base-commit.txt` 부터 **워킹트리**까지를 비교한다. 그
창 안에서 바뀐 기존 Go 함수는 전부 증거를 요구받으므로, base 뒤에 남이 고친 함수까지
이 change 의 몫이 된다. Go 작업이 끝난 change 는 창의 끝을 착지 커밋으로 좁힐 수 있다.

```bash
python3 tools/logic-map/check_analysis.py --change <change-id> --record-landing
```

`openspec/changes/<change-id>/landed-commit.txt` 에 40자리 커밋 id 하나를 쓴다. 기록은
**커밋해야** 효력이 있다 — 게이트는 워킹트리가 아니라 HEAD 에서 읽는다(안 그러면 untracked
파일로 통과한 뒤 지울 수 있다).

**값은 저자가 고르지 않는다.** 도구가 이 change 의 `revision: current` 번들로 계산하고,
게이트는 적힌 값이 자기가 계산하는 값과 같은지 확인한다. 유효 조건을 통과하는 커밋은
대개 여럿이라(실측: 착지를 얻는 76건 중 65건, 최대 516개) 조건만으로는 값이 안 정해진다.

받는 착지는 다음 **여덟**을 전부 만족하는 **가장 낮은** 커밋이다: base 의 자손 · 고정
번들이 있음 · 그 소스 해시가 전부 맞음 · 그 번들을 역사에 넣은 **보통(비병합) 커밋이 있음** ·
번들이 역사에 들어온 커밋 이후 · 판정이 디스크에서 읽은 번들 바이트를 그 커밋이 들고 있음 ·
고정 소스 중 하나 이상이 base 와 다름 · 그 뒤에 이 change 자신의 Go 작업이 더 없음.

기록은 **덮어쓰이지 않는다.** 증거를 갱신해 기록이 낡으면: 번들을 갱신 → `landed-commit.txt`
를 지우는 커밋 → 다시 기록. 거절 문장이 이 길을 같이 말한다.

증거를 **빌리는** change, a063 **이관 예외**, 고정 번들이 **0** 인 change 는 기록을 받지
않는다. 5단계는 그때 워킹트리를 대상으로 삼고 명령을 권하지 않으며, 왜 못 좁히는지를 적는다.

**판정은 역사 하나 위에서, 한 번 읽은 바이트로 선다.** 명령(5단계 · `--record-landing`)은 시작할 때 `HEAD` 를
sha 로 한 번 풀고, 증거 디렉터리를 한 번 읽고, 그 뒤의 역사 읽기는 전부 그 sha다.

**판정을 내놓기(기록을 쓰기) 직전에 판정이 읽은 것을 전부 다시 읽는다.** 대조 집합은 손으로 고르지 않는다 —
`check_analysis.py` 가 **파이썬으로** 디스크를 읽는 자리가 깔때기 넷(파일 · 디렉터리 목록 · 트리 순회 · 종류)
뿐이고 그 깔때기가 읽은 결과를 원장에 적으므로, 재확인의 집합은 그 원장이다(a112 한 판 실측: 경로 1,739).
**범위는 거기까지다 (2026-09-22 정정).** 자식 프로세스가 읽는 것은 원장에 안 남는다 — 바뀐 함수를 정하는
`git diff`, base 소스를 꺼내는 `git show`, 워킹트리 Go 바이트를 읽는 `go run ./tools/logic-map`,
그리고 `execution_baseline.py` 의 읽기가 그렇다. 판정 도중 그것들을 건드리면 재확인이 통과한다.
전 판본의 "디스크를 읽는 자리가 깔때기 넷뿐 · 재확인의 집합은 **구조적으로** 판정의 집합" 은 거짓이었다 —
그 열거의 모집단이 이 모듈의 AST 였고 자식 프로세스는 모집단에 없었다. 실패도 적는다 —
없음 · 정규 파일 아님 · 권한 · 너무 큼. 하나라도 달라졌으면 판정 대신
`<무엇> changed while this change was being judged — …; run it again` 을 낸다. `HEAD` 는 그 재확인의 **앞뒤로**
묻는다(다시 읽는 동안에도 역사는 설 수 있다). 이 워크트리는 병행 세션이 같이 쓰므로 그때는 다시 돌리면 된다.
창 줄 끝의 `judged at HEAD <sha>` 가 그 판정이 어느 역사의 것인지 말한다.
`--record-landing` 은 쓰기 직전에 **거절 사유 전부**를 다시 묻는다 — 순회 도중 워킹트리가 더러워지면
(추적 Go 파일 편집) 기록을 만들지 않는다. 도구는 그 기록을 **덮어쓰지 않으므로**(지우려면 위의 "지우는
커밋" 을 거쳐야 한다) 잘못 쓰인 기록은 사람 손이 들고, 그 편집이 든 분기는 창 밖에 남는다.

**읽기는 조용히 건너뛰지 않는다.** 번들 파일이 정규 파일이 아니거나(FIFO · 장치 · 소켓) 폴더이거나
16 MiB를 넘거나 UTF-8이 아니면 **이름 댄 판정 줄**이다. 목록을 못 여는 번들도 그렇다. 조용히 넘기면 그 파일의
표가 열거형 호출 감사에서 빠지는데, 목록과 열기는 다른 syscall이라 "저장소에 그런 파일이 없다"는 여는 순간에
대한 진술이 아니다(게이트가 여는 순간에만 FIFO로 바꿨다 되돌리는 것으로 감사를 끌 수 있었다 — 타이밍이고
그 판의 하네스는 안 남겼다).
**남아 있는 건너뛰기는 하나가 아니라 넷이다 (2026-09-22 정정).** 사라진 파일 · 증거 디렉터리 안에서
디렉터리가 아닌 항목 · `_verdict` 가 이름으로 거르는 항목 · 트리 순회(`Path.rglob`)가 못 읽는 하위 트리에서
삼키는 `OSError`. 앞의 것은 원장에 남아 되돌려 놓으면 재확인이 잡는다. 뒤의 셋은 그렇지 않다.
필수 파일이 그런 것이면 그 대상의 `could not be read` · `is not UTF-8 text` 다.

## 이 게이트가 **안** 하는 것 (2026-09-22)

- **CI 는 이 게이트를 안 돈다.** 생산 호출자는 `tools/gate.sh:321` 하나다. CI(`.github/workflows/ci.yml`)
  는 이 도구의 **시험**은 돌리지만 `check_analysis.py --change <id>` 자체는 안 돈다(:94 가 그렇게 적어 뒀다).
  그러니까 이 판정은 사람이 `make gate` 를 돌릴 때만 선다.
- **재확인 뒤에도 창은 열려 있다.** 재확인은 판정을 내놓기 직전에 서고, 그 뒤 `gate.sh` 가 다음 줄로
  넘어갈 때까지의 시간은 아무도 안 본다. 재확인은 "판정이 읽은 것이 판정하는 동안 안 바뀌었다" 이지
  "게이트가 끝날 때까지 안 바뀐다" 가 아니다.
- **비용.** 7.5.2.3 의 원장 때문에 파일 열기가 늘었다 — a112 2,407 → 3,854(+60%) · a066 1,203 → 2,246(+87%),
  2026-09-20 실측(계측 시점의 값이라 ±1 로 움직인다). git 호출은 20,249 → 20,365.
- **새 거절이 오늘 거절하는 정상 입력은 0 이다.** 정규 파일 아님 · 목록 실패 · 16 MiB 초과 · UTF-8 아님의
  네 거절이 오늘 저장소의 번들 파일 **12,411** 중 거절하는 것은 없다(`analysis/harness/7524_census.py`).

## 바뀐 기존 함수는 diff **문법**으로 센다 (2026-09-22)

`required` 는 `git diff --unified=0 <base> [<target>] -- '*.go'` 를 읽어 정한다. `--unified=0` 이면
문맥 줄이 없어 본문은 전부 `-`·`+`·`\` 로 시작하므로, 열 0 의 `diff --git ` 과 `@@` 는 모호하지 않지만
`--- `·`+++ ` 는 **모호하다** — 지워진 소스 줄 `-- x` 가 `--- x` 로, 더한 소스 줄 `++ x` 가 `+++ x` 로 나온다.
그래서 파일 이름은 **첫 훅 앞에서만** 읽는다. 상태 없이 읽던 판본은 파일 중간에서 이름을 바꿔, 그 파일의
요구를 통째로 지우거나(`/dev/null` 모양) 편집 전 논리의 지도를 요구하게 만들었다(`revision: base` 모양).
이 저장소의 추적 `*.go` 에 그 모양의 줄은 **111 줄 / 10 파일** 있다(전부 raw string 안의 SQL 주석).

## a063 execution-baseline adoption exception

`execution_baseline.py`는 일반적인 baseline 재설정 도구가 아니다. a063의 고정된
planning/execution commit 쌍만 검증하며, `execution-baseline.json`이 없으면 기존
`base-commit.txt` 정책을 그대로 쓴다. 레코드가 있으면 검증 worktree는 detached HEAD여야
하고, schema는 JSON 정수 `1`(boolean 불가), 정확한 키 집합, SHA-256 digest, change-local
ledger와 서로 다른 adversarial/gstack review를 모두 가져야 한다. ledger도 정확한 schema,
키 집합 및 inherited-history debt 문구를 요구한다. 잘못된 레코드는 P로 fallback하지 않는다.

고정 P는 `da80ce31b6a1ab5d443016768f970a82bab102db`, E는
`e65e394bf84b3c6e4559a219e816af96d341d75d`이다. 일반 change의 `SDD_BASE_REF`는 persisted
P와 같아야 하며, 유효한 a063 record가 선택한 경우에만 E와 같아야 한다. `SDD_BASE_REF`는
effective base를 선택하지 못한다. 이 결과는 `execution-baseline adoption exception`일 뿐,
inherited history의 pre-edit compliance를 증명하거나 historical debt를 완료/면제로 바꾸지 않는다.

검증은 S와 H의 tracked Go path/blob/Git executable mode 및 실제 worktree bytes/mode를 직접
대조한다. 따라서 `core.filemode=false` 상태의 chmod, symlink, source substitution도 거절한다.
Unified diff가 losslessly 나타낼 수 없는 탭·개행 Go 파일명도 Function Logic Map inventory
생성 전에 거절한다.

초안만 만들 때는 다음을 쓴다. 기존 출력은 덮어쓰지 않으며, review digest/approval은 생성하지
않는다.

```bash
python3 tools/logic-map/execution_baseline.py \
  --change a063-align-attestation-renewal-profile --source <full-SHA1>
```

## External doctor interpreter for adoption worktrees

The adoption source guard intentionally rejects a repository-local `.sdd/.venv`.
For an adoption worktree, keep that environment outside the checkout and use the
exact repository pin:

```bash
uv venv /tmp/tossos-sdd-venv
uv pip install --python /tmp/tossos-sdd-venv/bin/python -r tools/sdd/requirements.txt
SDD_PYTHON=/tmp/tossos-sdd-venv/bin/python make sdd-doctor
```

`SDD_PYTHON` affects only the doctor's `typedb-driver` probe. An explicitly supplied
path must be absolute, executable, and outside the checkout both lexically and after
resolution; invalid selections and dependency failures do not fall back to a local
venv. The adoption validator itself remains unchanged, as does the advisory
`tools/sdd-history/refresh_indexes.py` interpreter behavior.
