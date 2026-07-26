# TossOS Full SDD

정본 절차는 `docs/WORKFLOW.md`, 에이전트 공통 계약은 `.claude/CLAUDE.md`다.

## Evidence stack

| 질문 | 도구 |
|---|---|
| 무엇을 만들어야 하는가 | OpenSpec |
| 현재 누가 누구를 호출하는가 | CodeGraph + HEAD |
| 관련 문맥 후보는 무엇인가 | CodeGraphContext |
| 함수 안에서 어느 분기가 실행되는가 | Go AST + Function Logic Map |
| 위험 패턴 후보가 있는가 | ast-grep |
| 과거에 무엇을 배웠는가 | file memory + GBrain |
| 출시 가능한가 | gstack + make gate |

## Daily commands

```bash
make sdd-doctor
make sdd-sync
make sdd-check
make gate CHANGE=<change-id>
```

GBrain 명령은 TossOS 전용 데이터 홈을 보장하는
`python3 tools/sdd/gbrain_project.py <command...>` 형식으로 실행한다.
보조 그래프와 기억은 advisory다. 현재 HEAD·OpenSpec·실행 테스트보다 높은 권위를 갖지 않는다.
단, CodeGraph는 hard evidence이므로 `make sdd-sync`가 기록한 worktree fingerprint가
`make sdd-check` 시점과 다르면 gate가 실패한다.
