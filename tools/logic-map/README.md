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
