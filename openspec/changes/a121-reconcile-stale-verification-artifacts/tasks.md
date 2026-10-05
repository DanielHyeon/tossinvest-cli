## 1. Contract and hard evidence

- [x] 1.1 Reserve `STORY-TOS-a121`, capture the implementation base, and validate this change strictly.
      — 2026-10-05: STORY 기존재(모순 없음, acceptance 2 는 제약으로 판독 — design 「로트 1 처분」),
      base 재고정 9408fc95 → de147cc2 단독 커밋 b81380dd(영수증은 커밋 메시지 — 자기 Go 커밋 0,
      대상 소스 0파일, required 242→0), strict validate 통과. Manager 승인(WORKFLOW 규칙 2) 이 줄이 기록.
- [x] 1.0 Obtain answers to design Revision 1 Q1–Q5 and write them into design/spec, replacing every
      `[비움 — Qn]`. Until then 1.3 cannot run (Q1 unanswered means the command can only refuse).
      — 2026-10-05 Manager: 전 결정 반영(Q1=(a)+측정 전 잠정 거절·Q2=(a) 기존 키 재사용·Q3=구현 로트
      보수 상수 이월·Q4/Q5/Q6=거절). design 3자리·spec 3자리 치환, "Q1 is unanswered" 시나리오를
      "retention measurement is absent" 로 재서술. Q1 측정(CLOSED 발동 잔존·보존 기간, 조회 전용)은
      사람 실측 큐 — M-A 세션이 만드는 발동이 표본이 될 수 있어 동승 후보. **단(freeze P1-5):
      Q1 표본 발동은 a063 artifact 의 심볼로 만들지 않는다** — 같은 심볼의 발동 이력이 CLOSED 에
      남으면 G1-2 가 그 artifact 를 영구 거절한다.
- [ ] 1.2 Produce CodeGraph and Go AST Function Logic/Branch Test Maps for the existing functions that will
      be edited. By Revision 1 these are at least `verifylive.Artifact.terminal` (third ending), `newVerifyCmd`
      (registering the command), `verifylive.BuildReport` and `verifylive.BuildProgress` (label reconciled
      absent), and — only if Q2 picks the tag form — `verifylive.outstandingLines`. `PendingCleanup` and
      `M0Unsettled` are called, not edited; record their caller evidence.
      The branch evidence Revision 1 already cites is in `analysis/ast-evidence/` (HEAD `cb378a63`);
      regenerate it at the implementation base.
      — 2026-10-05 로트 1 착지(29a0d745): FLM 번들 4(terminal 0분기·newVerifyCmd 0·BuildReport 10·
      BuildProgress 7, BTM 은 base 커버리지 실측), PendingCleanup/M0Unsettled 호출자 증거
      (`analysis/code-context/`), ast-evidence 16/16 바이트 동일 + heldAfter·LastEntry 추가,
      outstandingLines 비편집 확인(Q2=(a)). check_analysis rc=0. **추가 편집 대상(S1 처분):**
      `Report.WriteText`·`Progress.WriteText` — FLM 은 구현 로트 시작 시 생성. freeze P2-11: proposal 이
      주장한 `runCleanup` 분기(AST 미열거)도 같은 시점에 번들로 만든다. [x] 는 두 유보 포함.
- [ ] 1.3 Complete proposal-freeze adversarial and gstack reviews; record the accepted no-live-mutation boundary.

## 2. RED

- [ ] 2.1 Add failing tests that a DELETE 404 or a generic operator observation cannot reconcile an artifact.
- [ ] 2.2 Add failing tests for complete fresh OPEN+CLOSED pagination, exact opaque identity, profile/account
      binding, stale/partial/ambiguous reads, and idempotent append-only reconciliation.
- [ ] 2.2.1 Add failing tests for each Revision 1 G1 refusal: a live successor under a new identifier in the
      symbol's OPEN group; the artifact's identifier in CLOSED (`COMPLETED` with `triggeredOrderId`, and
      `EXPIRED`); **a different identifier in CLOSED with a non-empty `triggeredOrderId` or `COMPLETED`**; any
      OPEN plain order on the symbol; a row whose symbol or market differs; all other conditions true while the
      Q1 retention measurement is absent; an artifact older than an injected retention bound (and a zero
      `CreatedAt` on the outstanding line); the Q3 constant absent; a status transition between the two reads
      (compared as multisets of (group, id, status, triggeredOrderId)); a duplicate (group, id) within one read;
      the fixed read order (conditional OPEN → plain OPEN → CLOSED) pinned structurally; the pair exceeding the
      freshness bound; repeated cursor, empty cursor with `hasNext`, page cap, and a read error in either read;
      the Q6 refusal message content ("만료된 artifact 는 이 경로로 영구히 대사되지 않는다"); the positive
      controls (symbol bytes taken from the artifact's record line, instrument GET succeeding and echoing the
      symbol) and their refusals when the control fails.
- [ ] 2.2.2 Add failing tests for each Revision 1 G2/G3 refusal: zero or several candidates from
      `PendingCleanup`; a conditional whose hold is not released; an artifact named by `M0Unsettled`; mixed
      `AccountRef`s, or `attest.Mask(current reference)` differing from them; account sequence 0; missing
      `--config-dir`; credentials from `TOSSCTL_OPENAPI_KEY`/`SECRET`; `--record` override; missing `--market`;
      the record changing between selection and append; a second request after reconciliation appends nothing;
      the basis digest carries its version and domain tag and is independent of read order. Also a positive test
      that the a063-shaped artifact (hold released by a failed `conditional-cancel`) is the one candidate.
      로트 1 추가(2026-10-05): `attest.Mask` 경계 — 빈 참조("(none)")·4자 이하(전부 `*`, 길이 비교로
      퇴화) 자격 참조의 거절 시험; 대사 줄 `StepID` 가 `Steps()` 카탈로그 ∪ {cleanup, abort} 밖이고
      `conditional-cancel` 이 아님을 고정하는 구조 시험(R1 — StepID 만 비교하는 소비자 넷:
      `LastEntry`·`heldAfter`·`m0ManualReconcileIDs`·`baselineSellable`).
      freeze 추가(2026-10-05): 계좌 후보 복수(DisplayName 비공란 2+·끝 4자리 충돌) 거절; 환경 변수
      자격 **하나만** 설정된 상태의 거절; flock·rate-budget lease 획득 실패의 거절; 기록 파일이
      `\n` 으로 끝나지 않을 때(찢긴 꼬리) 추가 전 거절; 중복/모호 신원 거절을 S7 포괄이 아닌 자기
      문구로 단언. 재검 P2-d: 후보가 order-kind artifact 뿐이라 0 이 되는 객체 유형 사례의 거절 시험.
      codex 라운드 추가(2026-10-05, F1~F7): `{"result":null}`·`{"result":{}}`·컬렉션 키 결측 응답이
      빈 목록이 아니라 거절이 되는 시험(F2 — 두 번 같은 기형 응답 + 종목 조회 성공 조합 포함);
      기록의 깨진 마지막 줄(개행 포함)이 선택·추가 양쪽에서 거절되는 시험과 내부 줄 깨짐 시험(F3);
      CLOSED status 결측·미지 값·그룹 모순 값의 거절 + allowlist 가 골든/영수증 전사임을 고정하는
      시험(F4); `Second` 비공란 행(OPEN·CLOSED 각각)의 거절 시험(F6); 추가 직전 신선도·Q1 나이
      재검사의 시험 — 느린 요청/정지 시뮬레이션으로 Q3 창이 추가 승인까지 덮음을 단언(F7);
      끝 4자리 같은 다른 계좌로의 자격 교체 픽스처에서 사람 승인 출력이 두 마스크·계좌 수를
      표시하고 대사 줄에 마스크 외의 계좌 신원이 실리지 **않는** 시험(F1·R2-5 — digest 철회);
      Q1 측정 기록의 세 거절 사례 — 축출 모형 결측·행 수 기반 판정·판별 불능(F5·R2-4);
      null 값 필드(컬렉션 null·hasNext null·hasNext 와 커서 모순) 거절 — 조건주문·일반 주문
      페이지 양쪽(R2-1); 검증 seq 와 transport 캐시 seq 불일치 픽스처(첫 양수 seq 자동 캐시)에서
      명시 재결속이 없으면 실패하는 시험 + 기형 계좌 행(빈 번호·양수 seq) 거절(R2-2);
      개행 없는 완전 JSON 꼬리에의 추가 거절 — 엄격 해독과 개행 검사 병존 단언(R2-3).
- [ ] 2.3 Add failing tests that reconciliation removes only its exact outstanding artifact from resume cleanup
      planning while preserving failed cleanup evidence and every verification verdict. freeze 추가(P1-6):
      대사 전/후의 `RedoSet` 을 핀한다 — 대사 뒤 `subjectLost` 가 `conditional-register` 를 되살리는
      현 동작을 바꾸지 않고 기록으로 고정(설치 실행은 여전히 사람 일괄 승인 뒤).
- [ ] 2.4 Add failing tests that the reconciliation event cannot enter successful endpoints, soak attestation,
      or engine-interlock coverage, and that no live mutation is reachable.
- [ ] 2.4.1 Add a structural test that the reconciliation dependency exposes only official GET reads and one
      local append (no cancel/modify/create/place method reachable); a test that the reconcile line carries no
      `Calls` and changes no `SucceededEndpoints` result; a test that the command holds the journal execution
      flock and the rate-budget lease; and a test that an older-format reader still reports the reconciled
      artifact as outstanding and held (the safe direction of rollback), simulated with the pre-change
      `Artifact`/`outstandingLines`/`holdGate` pinned by AST or by building the implementation base.
      freeze 재검 추가(P2-b): 명시 census 내용 — 대사 파일(cmd·verifylive)에서 쓰기 7이름(PlaceOrder·
      CancelOrder·ModifyOrder·CreateConditionalOrder·ModifyConditionalOrder·ModifyConditionalOrderRef·
      CancelConditionalOrder) 호출 금지 + type assertion 금지 + `Broker` 를 반환하는 생성자 부재;
      보존 한도가 **생산 빌드에서 nil** 임을 고정하는 AST 핀 시험; 「one local append」의 예외는
      토큰 갱신 POST(auth 기반, 경계 밖) 하나뿐임을 경계 서술에 포함.

## 3. GREEN

- [ ] 3.1 Implement the minimal versioned append-only reconciliation event and safe record projection.
- [ ] 3.2 Implement the bounded official-read reconciliation command with redacted operator output and no
      mutation path.
- [ ] 3.3 Preserve backward-compatible record handling or fail closed before write; document the rollback rule.

## 4. VERIFY and handoff

- [ ] 4.1 Run focused tests, `make test`, `make vet`, `make validate`, `make sdd-sync`, and `make sdd-check`.
- [ ] 4.2 Complete independent adversarial diff/test review followed by gstack review and Manager verification.
- [ ] 4.3 Record a read-only, redacted reconciliation observation only with explicit human approval for the
      selected profile; do not perform a live order mutation.
- [ ] 4.4 Synchronize PM, run `make gate CHANGE=a121-reconcile-stale-verification-artifacts`, and archive only
      after it succeeds. a063 remains independently unarchived until its own tasks complete.
