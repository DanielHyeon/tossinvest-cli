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

## proposal-freeze 적대 리뷰 (2026-10-13, 분리된 컨텍스트 — code-reviewer 서브에이전트, 저장소 무편집)

판정: **APPROVE-WITH-FIXES** — P1 4 · P2 9 · P3 7. 「봉인 방식·등록 입구·20 자리 FieldAccount」는 성립으로 확인됨.
절차 지적: 리뷰 시점에 새 leaf 파일(`internal/redact/` 등)과 두 `record.go` 의 구조체 필드가 이미 작업 트리에 있었음 —
기존 함수 편집은 0 이었고, RED-1(`analysis/red/red-1.txt`)은 base + 시험 파일만(redact 비시험 파일 0)으로 잼. RED-2 는 새 leaf 뒤 · 배선 전.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| R-1 | P1 | cobra 가 루트의 `ErrOrStderr()`(원시 os.Stderr)에 `Error: <err>` 를 한 번 더 씀 — R3(main 줄)이 못 덮음 | **수용** — `main` 이 루트에 `SetErr(redact.Writer(os.Stderr))`. 모든 명령의 stderr 경로가 규칙을 지남(stdout 은 대화형 출력이라 감싸지 않음 — `account list` 가 계좌를 보여야 함). 구조 핀 + spec 시나리오 |
| R-2 | P1 | 열 수 없는 봉인 → 빈 값은 섞인 기록에서 fail-closed 가 아님(빈 계좌 주기도 연속일수·endpoint 증거에 셈) · 옛 읽기는 평문 + 봉인 섞인 기록으로 **발급**함 · 키 분실 뒤 새 키 생성 | **수용** — (a) 열 수 없는 봉인 줄은 **해독 오류**(마지막 줄 찢김 관용에도 안 걸림) · (b) 기록에 봉인 줄이 있는데 키가 없으면 쓰는 쪽이 키를 만들지 않고 거절 · (c) `RecordFormatVersion` 2(soak · verifylive) — 옛 읽기는 「newer format」으로 거절 · (d) spec 시나리오 |
| R-3 | P1 | 키 실패가 실주문 **뒤** Append 에서 날 수 있음(verify runner `:538-545`) | **수용** — 키 준비(읽기/생성 + 시험 봉인)는 `OpenRecorder` 에서, 브로커 호출 전. `Append` 는 메모리 키만 씀 |
| R-4 | P1 | 정리 명령의 soak 감지가 컨테이너 경로에 안 맞음 · soak 의 열린 fd 는 rename 뒤 사라진 inode 에 씀 · audit 는 교체 창에 줄을 잃음 | **수용(a·c), (b) 대체** — (a) soak · verify `Append` 가 쓰기 전 경로와 fd 의 inode 를 대조해 다르면 다시 엶 · (c) 정리 명령은 원본 fd 를 rename 뒤에도 쥐고 그 사이 늘어난 줄을 같은 규칙으로 새 파일에 옮김(크기가 두 번 연속 같을 때까지). (b) 엔진 잠금 요구 대신 이 둘로 rename 이 그 자체로 안전 — 단, 옛 바이너리(재열기 없음)가 기록을 쥔 동안엔 안전하지 않으므로 절차는 「교체 뒤」(design 배포 3) |
| R-5 | P2 | 기존 마스커 `obs.MaskAccount`(4 호출) 미기재 | 수용(기재) — 넘겨받은 계좌 하나만 `[account]` 로 바꾸는 선행 부분 규칙이고 그 출력도 핸들러를 지남. 편집하지 않음(약화 없음) — D1 표에 행 추가 |
| R-6 | P2 | census 는 이름 기반(역할 기반 주장 과장) · `DisplayName` 누락 · 클로저 건너뜀 | 수용 — census 에 `DisplayName` 추가·클로저 포함으로 재실행, D1 문구를 「이름 정규식 + 싱크 인자 위치」로 정정(완전성은 census 가 아니라 싱크의 규칙이 보장) |
| R-7 | P2 | `adj-`·`corr-` 유도 id 누락 | 수용 — 규칙에 `adj-`(24 hex) · `corr-`(32 hex) 추가. `fillObservationID`(접두 없는 64 hex)는 로그 0 · 입력에 관측 시각 포함 — 잔재로 명명 |
| R-8 | P2 | 핸들러가 Any 값 마샬·Stringer 호출 중 패닉 시 안전 루프를 죽임 | 수용 — 속성 하나의 치환을 recover 로 감싸 실패하면 `[unloggable]` 문자열로 대체 |
| R-9 | P2 | 기록 전체 문자열 치환이 식별자를 바꿀 수 있음 | 수용 — 키가 `id`·`_id`·`ids`·`digest` 로 끝나거나 `digest` 를 포함하면 그 값은 치환하지 않음 + 식별자 불변 시험 |
| R-10 | P2 | 정리 명령을 root 로 돌리면 root 소유 키 → 컨테이너가 못 읽음 | 수용 — `--apply` 는 실행 uid 가 대상 디렉터리 소유자와 다르면 거절. 갱신 상태의 별도 사유 코드는 범위 밖(오류 문구가 키 파일 경로를 말함) |
| R-11 | P2 | audit 치환 범위가 spec 의 「보존」과 모순 | 수용 — `RedactEntry` 는 `Detail` 만 치환(실측 계좌 위치), spec 문구 일치 |
| R-12 | P2 | attest-autorun.log 생산자·정리 주체 없음 | 수용(명명) — 저장소 밖 호스트 리다이렉트(서비스 단위는 journald). 과거 줄 정리는 engine.log 와 같이 Manager · 사람(범위 밖), 새 줄은 출력 줄 삭제 + stderr 감쌈으로 0 |
| R-13 | P2 | spec 시나리오 부족 · 기존 시험 수정 목록 · 전역 등록 | 수용 — 시나리오 추가, 수정한 기존 시험은 receipts 에 목록, 전역 등록은 시험 사이 센티널만 더함(제거 불필요 — 문서화) |
| R-14 | P3 | HEAD 의 게이트 수락 detail 은 마스크(`interlock.go:420`), 원문은 옛 빌드 | 수용 — proposal 표에 주석 |
| R-15 | P3 | `accountSeq` 판정 불일치 | 수용 — 값은 치환 안 함, 이름이 `account_seq` 인 필드는 버림(D0 문구) |
| R-16 | P3 | 마스크 끝 4자리에 구분자가 끼면 규칙이 꼬리를 남김 | 수용 — `\*{3,}[0-9](?:-?[0-9])*` |
| R-17 | P3 | Register 반복 비용 | 확인 — 이미 멱등(집합 검사, 정규식 재컴파일 없음) |
| R-18 | P3 | Writer 반환 길이 | 확인 — `len(p)` 반환(시험 고정) |
| R-19 | P3 | docker json-file 로그 · 런타임 크래시 출력 잔재 | 수용(잔재 목록) |
| R-20 | P3 | 키 형식 모호 | 수용 — 32 무작위 바이트의 hex 64자, 다른 길이 거절(시험) |
