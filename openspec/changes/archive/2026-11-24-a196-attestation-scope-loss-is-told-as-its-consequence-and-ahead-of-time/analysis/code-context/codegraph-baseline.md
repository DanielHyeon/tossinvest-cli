# CodeGraph baseline (a196, 1단계)

- 도구: codegraph **1.6.0**(`codegraph --version`). 이 worktree 전용 색인을 새로 만듦: `codegraph init -y .` → 노드 55,950 · 엣지 206,351, 파일 3,237
  (`codegraph status .` — 「Index is up to date」). 공유 체크아웃의 `.codegraph/` 는 쓰지 않음.
- 기준 commit: `064f59fd9`(worktree `~/tossos-work/a196-impl`, HEAD = base, 작업 트리의 Go 변경 0).
- 원문: `codegraph-callers.txt`(12 심볼, `codegraph callers -l 50 <심볼> -p .`).
- 주의: callers 는 **이름** 단위다 — `binding`(11건 중 10건이 riskbucket·journal·strategyflow 의 다른 `binding`)·`prepare`(8건 전부 다른 `prepare`)·
  `supervisedProofs`(soak 시험 헬퍼 동명 함수) 가 섞인다. 아래 production 호출자는 HEAD grep 으로 좁힌 값.

## 질의와 결과 (production 호출자 · 자리 수)

| 질의 | CodeGraph 결과 | HEAD 로 좁힌 production 호출 자리 | 쓰임 |
|---|---|---|---|
| `judgeAttestation` | 4(시험 3 + `prepare`) | `prepare` call@283 — **1자리** | D1·D3 편집 지점 |
| `binding` | 11(이름 섞임) | `judgeAttestation` call@91 — **1자리**(+ 시험 `TestA135SupervisedProofAtIsTheBindingProof`) | D1 의 `ProofAt` |
| `stageKey` | 1(`prepare`) | `prepare` call@294 — **1자리** | D4·D5 |
| `attestationWatch.prepare` | **0 — 못 찾음**(동명 8건만) | `checkOnce` `w.prepare()`@232 — **1자리**(grep `w\.prepare()`) | D4·D5 |
| `attestationWatch.eventFor` | 1(`prepare`) | `prepare` call@298 — **1자리** | D5·D7 |
| `attestationWatch.checkOnce` | 10(시험 9 + `Run`) | `Run` call@218 — **1자리** | 편집 대상 아님(「보냄」 기록 L241·L257) |
| `daysLeftCeil` | 1(`judgeAttestation`) | call@112 — **1자리** | D3 조건부 |
| `newAttestationWatch` / `AttestationWatch` | 2 / — | `Context.AttestationWatch` → `cmd/tossctl/engine.go:749` 보조 실행자 배선 1 | 무편집 |
| `soak.BuildAttestation` | 15(시험 14 + `runSoakAttest`) | `runSoakAttest`(cmd/tossctl/soak.go:525) — **1자리** | D1·D2 전제 |
| `acceptSupervised` | 1(`BuildAttestation`) | call@242 — **1자리**, 인자 `c.Validity`(같은 함수 L222 에서 `withDefaults` 뒤 값) | D1·D2 |
| `supervisedProofs`(cmd/tossctl) | 10(동명 soak 시험 헬퍼 섞임) | `runSoakAttest` :520 — **1자리**, 인자 `criteria.Validity` | D1 이중 필터 |
| `verifylive.SucceededEndpoints` | 12(시험 다수) | `supervisedProofs` call@438 — **1자리**, `maxAge = validity` | D1 이중 필터 |

## 영향 시험 (`codegraph affected`)

돌리지 않았다 — 편집 표면이 `internal/app/engine/attestation_watch.go` 한 파일의 함수 4개와 같은 패키지 시험 파일이고, 호출 사슬이 전부 그 파일 안
(`Run → checkOnce → prepare → judgeAttestation/stageKey/eventFor`)에서 닫혀 위 callers 로 열거된다. 하네스 주석대로 기본 판별식은 Go 시험을 못 찾으므로
2단계에서 필요하면 `codegraph affected internal/app/engine/attestation_watch.go --filter '*_test.go'` 로 잰다.

선택 시험(2단계, 전체 시험 아님): `./internal/app/engine -run 'A196|A135|Attestation'`, D2 전제 시험이 놓일 `./internal/soak -run '<그 시험>'`.
