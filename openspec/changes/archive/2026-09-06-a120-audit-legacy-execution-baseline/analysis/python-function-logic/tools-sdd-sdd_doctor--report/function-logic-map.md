# Function Logic Map: `tools/sdd/sdd_doctor.py::report` (pre-edit)

## Evidence identity

- Current HEAD: `19021ada90135532adfaa70712fb199e42e1c95b`.
- Source is unchanged from HEAD (`git diff --quiet -- tools/sdd/sdd_doctor.py` exited 0).
- Source SHA-256: `eb6c4ecb4d1f80c17a1514bd93ba7683b37f40d5c8e508aec5c89e95a6a7c26d`.
- Exact Python AST locations and evidence-local IDs are in [ast.json](ast.json).

## Inputs and invariants

`root` defaults to the repository root. The function always reports the configured
CLI tools, skills, required files, optional service probes, and the TypeDB driver
probe. Before this amendment there is exactly one interpreter selection: node
`N004` constructs `root/.sdd/.venv/bin/python`.

## Control flow, calls, and outputs

| Node | Location | Condition / action | Observable result |
|---|---:|---|---|
| N003 | 92:5–92:85 | Builds status for every `TOOLS` command through `command_status`. | `tools` result group. |
| N004 | 93:5–93:60 | Constructs the repository-local venv interpreter path. | No environment override exists. |
| N005 / N015 | 94:5–104:72 / 94:8–94:27 | Tests whether the local path exists. | The only branch in `report`. |
| N016 | 95:9–102:10 | Calls `command_status` with the local interpreter and metadata version probe. | Existing driver result; version is reported but not compared with the pinned requirement. |
| N017 | 104:9–104:72 | Local interpreter missing. | Required probe fails with `run \`make sdd-infra\``. |
| N006 | 105:5–108:6 | Builds required-file status. | `files` result group. |
| N047 | 111:19–111:33 | Calls `skill_status`. | `skills` result group. |
| N086 / N087 | 115:31–116:61 | Calls `service_status` for shared TypeDB and Neo4j. | Advisory `services` group. |
| N007 | 109:5–118:6 | Returns all five report groups. | No write, checkout, order, or runtime action. |

## Later-edit boundary

The amendment may alter only interpreter selection and the TypeDB-driver status
within this function. It must retain the absent-variable local-path branch and
the rest of the returned report schema unless the approved contract is revised.
Invalid explicit `SDD_PYTHON` values must stay in the failing path; they must not
reach N004's local fallback behavior.
