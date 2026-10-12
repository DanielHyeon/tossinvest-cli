# a131 review

## Pre-Edit Gate (2026-10-13, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a131-no-account-in-any-log / 1.2
- 대상 심볼(패키지.함수): obs.NewLogger · official.(*Client).accountsLocked · main.newEngineRunCmd · main.main ·
  main.newRootCmd · main.runSoakAttest · audit.(*Log).Record · soak.(*Recorder).Append · soak.LoadCycles ·
  verifylive.(*Recorder).Append · verifylive.(*lockedRecord).appendLine · verifylive.LoadEntries ·
  verifylive.readRecordStrictNoTail (+ 타입 soak.Cycle · verifylive.Entry 필드 하나씩, 새 leaf: internal/redact/ ·
  internal/audit/redact.go · internal/soak/account_seal.go · internal/verifylive/account_seal.go · cmd/tossctl/scrub_account_records.go)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/)
  — NewLogger 생산 2 · accountsLocked 2(Accounts · ensureAccountSeq) · runEngineRun 1 · LoadCycles 생산 2 · LoadEntries 생산 7 ·
  appendLine 1(Reconcile) · readRecordStrictNoTail 1 · runSoakAttest 1 · affected --filter 1119 줄
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
  (R4: CGC 색인 0 — census·HEAD grep 으로 대신, R5: engine.log 는 자식 프로세스의 stdout/stderr)
- 기존 동작 파악 근거: base 233d5d9c HEAD 직접 읽기 · AST 번들 13(analysis/function-logic/) · census(analysis/census/census-base.tsv) ·
  운영 파일 개수 실측(값 미열람) · 기존 시험 TestSoakAttestWritesAVerifiableAttestation(soak 기록 → attestation 왕복, CLI) ·
  TestBuildAttestationPassesTheEngineInterlock · internal/verifylive reconcile_* 시험군
- Function Logic Map / Branch Test Map: analysis/function-logic/ 13 번들(편집 전, 작성기 analysis/harness/flm_write.py)
- upstream 상속 테스트 영향: no — 편집 패키지 중 upstream 은 cmd/tossctl(main · root)뿐, 동작 불변(오류 줄의 계좌만 치환).
  관련 패키지 전체 go test 로 회귀 확인
- 실패 테스트 선행 작성: yes — analysis/red/red-1.txt(패키지 없음) · red-2.txt(새 leaf 만 있고 배선 전: 행동 실패 9 + 구조 실패 2 +
  audit · cmd 빌드 실패). 호환 시험 둘(TestA131SealedSoakRecordStillAttestsTheSameAccount · TestA131LegacyAndSealedLinesReadTheSame)은
  base 에서도 통과 — 회귀 핀이며 RED 대상 아님
- 설정·DB·journal 변경과 rollback: 원장·설정·attestation 형식 변경 없음. 기록 파일 형식에 `account_ref_sealed` 추가(옛 읽기는 계좌 없음으로
  fail-closed — design 「배포」의 함께 교체 · 롤백 절)
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문·청산·대사·체결 판정 무편집, 로그·기록 바이트만. 손절 경로의 로그 쓰기에 순수 치환 추가
  (줄당 정규식 몇 번). 등록은 메모리 쓰기 하나. 봉인 실패는 측정 도구 기록 쓰기의 오류(엔진 경로 아님)
```
