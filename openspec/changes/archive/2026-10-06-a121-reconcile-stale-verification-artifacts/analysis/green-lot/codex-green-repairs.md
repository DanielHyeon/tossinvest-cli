# a121 4.2 외부 모델(codex, clean) GREEN 패스 — 발견 6건 수리 기록

- 작성: 2026-10-05, Terra(GREEN 수리), worktree `/tmp/a121-green` = `a25e1c3b` + 수리
- 원문: `/tmp/claude-1000/a121-codex-green-out.txt`(정적 리뷰, 생산은 Q1 nil 이라 지금은 추가 불가 — 전부 미래 활성 경로)
- RED 선행: 아래 새 시험을 `a25e1c3b` 코드 사본에서 돌려 실패를 실측했다(새 심볼을 쓰는 CG-2·CG-3·CG-5 시험은 그
  사본에서 컴파일 단계에서 실패 — 가드 부재). 영수증은 본문 각 항.

## CG-1 — OCO·신원이 둘째 읽기에서 탈출 (FIXABLE → 수리)

- 결함: multiset 비교 tuple 은 (그룹, id, status, triggeredOrderId) 라 `HasSecond`·`Symbol`·`Market`·필수 필드가 둘째
  읽기에서만 어긋나도 같다고 보고, 행 검사는 첫 스냅숏에만 돌았다.
- 수리: 행 검사(OCO → 필드 결측 → 심볼 → 시장)를 **두 스냅숏 각각**에 돌린다(reconcile.go 10단계). multiset 비교는 유지.
- RED: `TestReconcileValidatesTheSecondSnapshotToo`(second-leg · other-symbol · other-market · missing-field). 수리 전 사본:
  네 사례 모두 `got <nil>`(추가됨) — 실측.

## CG-2 — 기록 배제가 쓰기 직렬화가 아님 (FIXABLE → 수리, 잔여 1 기록)

- 결함: 배제는 configDir 의 flock·lease 뿐이라 같은 기록 파일(하드링크 별칭·`--record` override)의 다른 작성자를 못
  막고, 지문 검사 뒤 경로로 append fd 를 다시 열었다.
- 수리: record.go 에 `lockRecordForAppend`(기록 파일 자체 `O_RDWR|O_APPEND` + `flock LOCK_EX|LOCK_NB`, 별칭은 같은
  inode 라 같은 잠금)와 `lockedRecord.contents`(잠근 fd 로 처음부터 읽음)를 두고, 대사는 **잠근 뒤 그 fd 로 읽은 바이트**로
  엄격 해독·지문·개행을 판정하고 같은 fd 로 쓴다(판정한 바이트 = 쓸 파일). 쓰기 직전에 fd 크기가 판정 길이와 다르면
  쓰지 않는다(`RefuseRecordChanged`).
- 새 거절 코드: `RefuseRecordLocked`.
- RED: `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile`(같은 경로·하드링크 별칭),
  `TestReconcileDetectsAWriterInterleavingAtTheAppend`(쓰기 순간 끼어든 잠그지 않는 작성자 → read-back 실패),
  `TestReconcileFinalGateSitsAtTheWriteBoundary`(잠금 뒤 재읽기는 잠근 fd 의 contents — 경로 재읽기 금지, AST).
- **잔여(결함 기록 — 후속 change 후보):** `verify run`·`abort`·콘솔의 `Recorder`(`OpenRecorder`)는 기록 파일을 flock
  하지 않는다. 그래서 대사끼리는 직렬화되지만, 잠그지 않는 Recorder 작성자가 대사의 쓰기 순간에 끼어드는 것은 **막지
  못하고 잡기만 한다**(쓰기 직전 크기 검사 + read-back 대조 → 침묵 없는 실패). 완전한 직렬화는 `OpenRecorder` 가 같은
  잠금을 쥐는 편집(기존 고위험 함수 — 동시에 같은 파일을 여는 시험 25개 파일 영향 조사 필요)이라 이 수리 범위 밖이다.

## CG-3 — 자기 부분 쓰기 (FIXABLE → 수리 범위 내, 상속 결함 기록)

- 수리: 대사 줄을 개행까지 **미리 직렬화**해 **한 번의 Write** 로 쓰고 fsync 한 뒤 **다시 읽어 대조**한다
  (`lockedRecord.appendLine`). 부분 쓰기·동기화 실패는 `errRecordAppendIncomplete`("inspect its tail before any
  `tossctl verify run --resume`"), 대조 불일치는 `errRecordAppendUnverified` — 복구는 못 해도 침묵하지 않는다.
- RED: `TestReconcileAppendsInOneWriteEndingWithANewline` · `TestReconcileReportsAPartialAppendLoudly` ·
  `TestReconcileReportsAFailedSyncLoudly` · `TestReconcileVerifiesTheAppendByReadingItBack`(쓰기·동기화 자리
  `recordWrite`·`recordSync` 주입).
- **상속 결함(고치지 않음 — 후속 change 후보):** `Recorder.Append`(record.go:335)는 부분 쓰기 뒤 찢긴 접두사를 남기고,
  `LoadEntries`(record.go:461)는 해독 불능 마지막 줄을 조용히 버리므로, 이후 `verify run` 이 그 찢긴 꼬리에 이어 붙여
  꼬리를 내부 손상으로 만든다(codex 3 의 연쇄). 모든 작성자가 추가 전 불완전 꼬리를 거절(또는 명시 복구)해야 한다 —
  runner 쪽 편집이라 이 change 범위 밖.

## CG-4 — 승인 읽기 (FIXABLE → 수리)

- 수리(cmd `verifyReconcileApprove`): 줄을 **끝까지**(개행) 읽어야 하고 EOF·읽기 오류는 거절, ctx 취소를 select 로 존중
  (대기 중 취소 = 거절, 잠금을 오래 쥐지 않음 — 읽기 고루틴은 단말 입력이 올 때까지 남는다: CLI 수명 안의 한계),
  질문 표시 실패는 거절. 승인 내용(두 마스크·계좌 수)은 `ReconcileParams.Out` = **단말 채널**(질문과 같은 stderr)로
  보내고, 표시 실패는 `RefuseApprovalDisplay`(`ReconcileApproval.WriteText` 가 오류를 돌려줌).
- RED: `TestReconcileApprovalNeedsACompleteLine`(y+EOF · y+읽기 오류) · `TestReconcileApprovalHonoursCancellation` ·
  `TestReconcileApprovalRefusesWhenTheQuestionCannotBeShown` · `TestReconcileRefusesWhenTheApprovalCannotBeShown`(verifylive) ·
  `TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion`(tagged). 수리 전 사본: 다섯 모두 FAIL — 실측(취소 시험은
  3초 동안 막힘).

## CG-5 — 최종 신선도 게이트 위치 (FIXABLE → 수리, 측정 한계 기록)

- 수리: 잠금·fd 준비와 줄 직렬화를 마지막 신선도·Q1 재검 **앞**으로 옮기고, 재검 뒤에는 한 번의 쓰기만 둔다. 줄의
  `ReconciledAt` 과 재검이 같은 시각 `at` 을 쓴다.
- 핀: `TestReconcileFinalGateSitsAtTheWriteBoundary`(AST — 잠금 → contents → 직렬화 → 게이트 → 쓰기, 게이트와 쓰기 사이
  다른 호출 0) · 기존 `TestReconcileRechecksTheWindowAtBothContractPoints`(표지를 decodeRecordStrict·appendLine 로 갱신).
- **측정 한계:** 경과는 `time.Now` 의 `Sub` 로 잰다. Go 의 단조 시계는 플랫폼에 따라 시스템 일시 정지(suspend) 동안
  멈출 수 있어, 정지가 게이트와 쓰기 사이(또는 두 읽기 사이)에 끼면 벽시계로는 한도를 넘었어도 통과할 수 있다.
  창은 쓰기 한 번으로 좁혔지만 0 은 아니다.

## CG-6 — 토큰 캐시 신원 교체 (INVESTIGATE → 완화 수리, 결속 결함 기록)

- 수리(완화): 좁은 생성자가 돌려주는 읽기 래퍼가 **첫 계좌 범위 읽기 직전**(사람 승인 뒤) **읽기용 클라이언트로**
  `Accounts()` 를 다시 불러, 계좌가 정확히 하나이고 계좌번호·seq 가 생성 때 검증한 값과 같은지 확인한다. 다르면
  `errReconcileIdentityChanged` 로 읽지 않는다(대사는 read-error 로 거절). 승인 대기 중의 캐시 교체를 덮는다.
- RED: `TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead`(다른 계좌·같은 seq / 같은 계좌·다른 seq / 계좌 둘).
  수리 전 사본: FAIL(`got <nil>`, 목록 읽기 진행) — 실측.
- RED 시험 개정 1건: `TestReconcileNarrowReaderBindsTheValidatedSequence` 의 `/accounts` 호출 수 기대값 1 → 2(판정 +
  재확인). 원래 픽스처(둘째 호출에 다른 계좌 seq 8)는 이제 재확인이 거절하므로 둘째 응답을 같은 계좌로 바꿨다(지연
  재해석이면 셋 이상이 되어 여전히 잡힘).
- **결함 기록(후속 change — 기록 작성기 F1 writer-side 와 묶음):** 토큰 캐시(`token.go`)는 토큰·만료만 담고 자격에
  결속되지 않는다. 재확인 뒤에도 토큰 갱신이 다른 자격의 캐시 토큰을 채택할 창이 남는다(읽기 도중의 교체). 캐시-자격
  결속은 이 change 범위 밖이다.

## RED 시험 수정(이번 수리가 요구한 것)

1. `reconcile_seal_test.go` 의 `reconcileAllowedOutside`: 추가 통로를 Recorder 3이름 → `lockRecordForAppend`·
   `lockedRecord.contents/appendLine/release` 넷으로 **대체**(CG-2·CG-3). 읽기 통로 `readRecordRaw` 는 유지.
2. `verify_reconcile_test.go` 의 `TestReconcileNarrowReaderBindsTheValidatedSequence`(위 CG-6).
3. 내 GREEN 시험 `TestReconcileRechecksTheWindowAtBothContractPoints` 의 표지 갱신(RED 아님).

## 새 편집 대상

없음 — 기존 함수 내부 편집 0. record.go 에는 새 함수·타입(`lockRecordForAppend`·`lockedRecord`·`recordWrite`·
`recordSync`)만 더했고, flock 은 새 파일 `record_lock_unix.go`/`record_lock_other.go`(비 unix 는 fail-closed).
