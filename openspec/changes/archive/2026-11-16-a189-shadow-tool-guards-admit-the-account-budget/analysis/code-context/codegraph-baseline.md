# CodeGraph baseline (a189, 1단계)

- 도구: codegraph **1.6.0**. 색인: main 체크아웃 `/mnt/D/Axipient/workspace/TossOS`(HEAD `7d1e3163a` = 이 change 의 base, 작업 트리 깨끗함). 이 worktree 에는 색인이 없어 같은 커밋의 main 색인을 읽기만 함(`codegraph sync` 안 함).
- 원문: `codegraph-1.6.0.txt`(status · callers 9건 · affected 2회).

| 질의 | 결과 |
|---|---|
| `callers a112MBUSReferenceAllowed` | 1 — `a112MBUSFindForbiddenRefs`(a112_mbus_static_test.go:423) |
| `callers a112MBUSFindForbiddenRefs` | 4 — `a112MBUSRepositoryForbiddenRefs`(394) + 픽스처 시험 3개(193·204·220) |
| `callers a112MBUSRepositoryForbiddenRefs` | 1 — `TestA112MBUSStaticNoProductCallerOrForbiddenPath`(170) |
| `callers a112MBUSSurfaceName` | 1 — `a112MBUSFindForbiddenRefs` |
| `callers nonTestSources` | 13 — a128 5·a140 3·a087 5(같은 이름의 함수 셋을 이름으로 합쳐 셈; a087 은 이 change 무관) |
| `callers` 시험 함수 4개 | 0(시험 실행기만 부름) |
| `affected <가드 3파일> --filter '*_test.go'` | 3 — 세 가드 파일 자신 |
| `affected <가드 3파일>`(기본 판별식, 비교용) | 1 — `auth-helper/tests/test_cli.py`(.claude/CLAUDE.md 하네스 주석의 Go 판별식 결함 그대로) |

결론: 고칠 함수 5개는 모두 시험 파일 안에서 닫힘. 운영 코드 호출자 0. 판정이 바뀌는 것은 그 다섯 시험과 `a112MBUSReferenceAllowed` 를 거치는 픽스처 시험 3개(선택지 A 에서 판정 불변이어야 함 — 세 픽스처 경로는 `internal/product/fixture.go`·`internal/official/ordinary_reader.go`·`internal/other/official_fixture.go` 라 더할 경로와 다름).
