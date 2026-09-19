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
말을 그대로 전한다 — 예: `cannot read blobs at … : \`git cat-file --batch -Z\` failed (rc 129:
error: unknown switch \`Z') — \`-Z\` needs git 2.42 or newer`. 영수증은 git 자신의 릴리스 노트
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
