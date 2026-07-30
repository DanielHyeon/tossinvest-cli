# Review: console-adoption-controls

## requirement-change 리뷰 — 외부 종목 자동관리 메뉴 (2026-07-31)

- 보이스 구성: Manager + 적대적 Eng/UI 안전 리뷰.
- 사용자 의도: 수동 매수한 외부 종목을 기존 익절·보호선·손익 극대화 정책으로 자동
  관리하는 설정을 메뉴에서 찾고 제어할 수 있어야 한다.
- 현재 증거: `engine.adoption.enabled/include/exclude/default_stop_pct`,
  `judgeHoldings → adopt → adoptOne`, config 저장 seam과 `/settings` UI가 이미 이를
  수행한다. 새 엔진·정책 로직은 중복이며 허용하지 않는다.
- 결정: **APPROVE — UI 발견 가능성만 개정**. 기존 `console-adoption-controls` Story와
  change를 유지하고 navigation·anchor·제목·경계 설명만 수정한다.
- 안전 처분: 저장은 config만 변경하고 편입·주문을 즉시 실행하지 않는다. 실제 편입은
  다음 엔진 기동의 Verified reconcile 루프만 수행한다. LIVE 토글, Guardian, journal,
  주문 경로는 무변경이다.
- Function Logic Map: not-applicable — template 상수와 문자열 테스트만 바꾸며 기존 Go
  함수 내부 분기·side effect·fallback은 수정하지 않는다.

### Pre-Edit Gate

```text
- change/task: console-adoption-controls / 5.2-5.3
- target symbols: baseTemplates, settingsTemplates; rendered settings test only
- CodeGraph definition/impact: complete; template-local impact
- CodeGraphContext: advisory update timed out; no authority taken from it
- existing behavior: /settings already persists adoption controls; engine reconcile performs adoption
- Function Logic Map: not-applicable; no function body edit
- RED first: rendered menu/anchor/title/execution-boundary assertion
- config/DB/journal rollback: none; text/template-only revision
- safety: no LIVE toggle, engine start, policy calculation, journal mutation, or order path
- decision: edit allowed after RED evidence
```

### 구현·배포 리뷰

- 결과: **UI 개정 ACCEPT**. 기존 자동 편입 설정과 reconcile 경로를 새 기능처럼
  복제하지 않았고, 메뉴 직접 링크와 실행 경계 설명만 추가했다.
- 안전: 저장 POST·CSRF·검증·audit seam은 그대로다. 이번 개정은 설정 저장도 실행하지
  않았으며 주문·journal·엔진 기동·LIVE 권한을 추가하지 않았다.
- 회귀: 최종 console race, 전체 test/vet/validate, strict OpenSpec, PM 1:1,
  code-reviewer 정적 검사를 통과했다.
- 배포: healthy Compose image에서 무인증 `/settings` 렌더와 새 메뉴를 확인했고
  엔진 마커가 생성되지 않았음을 확인했다.
- 독립성: change 전체에는 2026-07-27 별도 컨텍스트 독립 검증 ACCEPT가 존재한다.
  이번 template-only 개정은 WORKFLOW의 UI 경량 리뷰 규칙에 따라 Manager 분리 리뷰
  패스, 자동 code-reviewer, 런타임 스모크로 재검토했다. 정식 SDD gate가 남아 있으면
  archive/Full SDD 완료로 승격하지 않는다.

## proposal-freeze 라운드 1 (2026-07-27)

- 보이스 구성: **적대적 Eng**(별도 컨텍스트 검증 에이전트 — 작성자 분리) + Manager.
  High-risk(reconciliation 경로 judgeHoldings·콘솔 불변식 개정) 가중 적용.
- 판정: **REVISE** — P1 3건·P2 7건·P3 4건. 전건 수용, 계약 문서에 반영(아래 처분).

### 처분

1. **[P1-수용]** 거부 블록의 Load 왕복이 사용자 목록을 유실 — seam Load를 merge·zeroing
   이전의 **파일 원문 블록 + 검증 사유 별도 반환**으로 규정(design D3, read-only 불변식
   delta, "거부된 블록의 목록 보존" 시나리오, task 1.2).
2. **[P1-수용]** include_symbols가 §0.5 audit 밖 — `recordGateSettings` 항목 추가
   (exit-policy delta SHALL, task 2.2, Impact에 interlock.go 명기).
3. **[P1-수용]** D1 방어선 서술 오류·blast radius 확장 — 방어 사슬을 실제대로 재서술
   (게이트 콘솔 편집 불가 + Verified 없으면 ReconcileDriver 생성 거부 + 2c 전 상수 차단),
   enabled false→true 저장에 **타이핑 확인 문구** 요구(시나리오+task 3.2), proposal에
   위협 모델 확장·수용 근거 명시.
4. **[P2-수용]** 동시 기록 lost-update — CreateTemp 유일 임시파일 + flock RMW + 동시 기록
   테스트, 파싱 불가 파일 저장 거부(골격은 ErrNotExist 한정) (task 1.2).
5. **[P2-수용]** include 라우트 dead-end — 선행조건 미충족 시 **무기록 거부 + 안내**
   시나리오 신설(zeroing될 블록 기록 금지 SHALL).
6. **[P2-수용]** alertUnmanaged 사유 행렬 — D2에 5행 행렬 명시 + exit-policy delta에
   "include 시도 실패는 실패를 말한다" SHALL NOT + 행렬 전수 테스트(task 2.1).
7. **[P2-수용]** actVerbs 어휘 확장(save·set·config·include·enable) — 허용 목록 밖 신규
   config-쓰기 라우트의 침묵 통과 방지(task 3.3).
8. **[P2-수용]** seam 좁힘 가드 — 메서드 2개 고정 AST 가드 + internal/console의
   config.Service 비명명 정적 가드 + 블록 밖 바이트 동일 테스트(delta SHALL, task 3.3·3.4).
9. **[P2-수용]** 저장 시점 audit — seam Save 성공 시 cmd/tossctl에서 audit append(이중
   기록 의도 명시 — 기동 diff의 flip 붕괴 한계 보완).
10. **[P2-수용]** 지정의 상시성 문구 — 화면·응답에 "재매수 재편입 포함 상시 규칙" 명시
    (delta·exit-policy SHALL).
11. **[P3-수용]** 목록은 시장 무관 심볼 단위 — exit-policy delta에 고정.
12. **[P3-수용]** "유일한 쓰기 표면" 과대 서술 — "config에 대한"으로 한정.
13. **[P3-수용]** 저장 응답에 현재 엔진 실행 여부 안내(마커 판독) — delta SHALL.
14. **[P3-수용]** Impact에 interlock.go 명기(2와 함께).

리뷰가 생존 확인한 것: 후보 술어의 게이트 순서 유지(include가 신선·Stabiliser·Verified
우회 불가), Verified의 구조적 게이트 ON 결합(ReconcileDriver 생성 거부), 빈 include의
§0.2 동일성, hot-reload 부재 사실, 자격증명 파일 분리(저장 경로에서 도달 불가), CSRF
자세(SameSite=Strict + 프로세스별 토큰 + 루프백).

## proposal-freeze 라운드 2 (2026-07-27)

- 라운드 1 반영본 재검토(동일 적대적 리뷰어 컨텍스트) — 14건 처분 전수 실반영 확인
  ("여러 건은 요구보다 엄격" 판정).
- 판정: **FREEZE-APPROVE** — task 1.1부터 구현 착수 승인. High-risk 의무(Pre-Edit +
  FLM: judgeHoldings·alertUnmanaged·recordGateSettings·mergeAdoption·validate, race
  테스트, 랜딩 시 적대적 코드 리뷰) 유지.
- P3 노트 2건 반영: ① exit-policy audit 문구를 이중 기록(저장 시점+기동 시점)으로 정정
  ② D1의 "검증 승인과 같은 패턴" 서술 정정(고정 문구 — 등가성 근거 명시, 서버측 상수
  비교·prefill 에코 금지).

## 구현 기록 (2026-07-27)

### Pre-Edit Gate (High-risk — reconciliation·audit)

- change/task: console-adoption-controls / 2.1·2.2
- 대상 심볼: engine.ReconcileDriver.judgeHoldings·alertUnmanaged(adoption.go),
  recordGateSettings(interlock.go)
- 기존 동작 근거: adoption.go 전문·게이트 순서 판독, reconcileloop_test 16종·
  adoption_audit_test 판독, recordGateSettings 항목표 판독 (HEAD)
- CodeGraphContext 보조: not-applicable(대상 파일 전문 직접 판독)
- upstream 상속 테스트 영향: no (엔진 신설 코드 — upstream 표면 아님)
- 실패 테스트 선행: yes — adoption_include_test.go 4건 RED 관측(adopted=0·exclude 사유
  미표기·'adoption is off' 오표기·audit 항목 부재)
- §0 위반 검토: 통과 — 보호 확대 방향만(include=편입 대상 확대), exclude 우선 유지, 게이트
  순서 불변(신선·Stabiliser·Verified 우회 불가), 편입 tx 무발의(adoptOne 무접촉), 손절
  즉시성 무접촉, zero-value 안전(빈 include=현행)

### base-commit 재고정 사유

base 최초 고정(fa19225) 후 병행 세션의 add-net-rr-measurement 구현 커밋(fc9cf51)이 같은
브랜치에 착지해 diff 범위가 오염되었다. 관행(병행 세션 게이트 간섭 — 재고정)에 따라 base를
fc9cf51(내 구현 직전 조상, 이 change의 파일 무접촉)로 재고정 — diff는 정확히 이 change의
수정만 담는다. 계약 문서는 fa19225 시점 freeze본에서 무변경.

### 검증 결과

- RED→GREEN: config 신규 12 테스트·engine 신규 5·console 신규 10(가드 RED: CSRF 목록
  불일치 실패 관측 후 개정) — `go test ./internal/config/ ./internal/app/engine/
  ./internal/console/ ./cmd/tossctl/ -race -count=1` **629 passed**
- FLM 15 번들(수정 기존 함수 10 + 신규 함수 3 + 개정 테스트 함수 2) —
  `check_analysis.py` PASS(격리 worktree)
- PM allowlist·fixture 등재, tracker --check PASS, strict validate PASS
- 콘솔 동작 테스트 일부(settings 화면 렌더 계열)는 구현과 동시 작성된 핀 테스트임을 명시
  — 세션+CSRF·확인 문구·무기록 거부 등 안전 단언은 seam 카운팅으로 실검증

## 독립 검증 (2026-07-27) — 라운드 1 REJECT → 보완

별도 컨텍스트 검증 에이전트 판정 REJECT(협소·보완 가능 — 동작·안전 불변식은 전부 clean,
계약 이행 증거 부족 5건). 보완 내역:

1. **[수정]** 사유 행렬 5행 전수 pin — Rejected/enabled-실패/기본 행 3건 테스트 추가
   (`TestARejectedBlockAlertNamesTheRefusal`·`TestAnEnabledFailedCycleSaysOnAndFailed`·
   `TestTheDefaultRowSaysOffAndUndesignated`).
2. **[수정]** cmd/tossctl 시임 실증 테스트 추가(`adoptionsettings_test.go`) — 저장 시점
   audit append 실증(SHALL pin), 실제 시임 경유 외과 기록 보존, typed-nil 가드.
3. **[수정]** 확인 문구 강화 — 거부된 현재 블록(=엔진상 OFF) 위로 켜는 저장도 문구 요구
   (`TestEnablingOverARefusedBlockStillNeedsThePhrase`).
4. **[기록-축소 명시]** actVerbs에 "set" 미포함 — contains 매칭이 GET `/settings` 자체를
   오검출하므로 의도적 제외(save·include·enable·config로 대체). 계약의 어휘 예시에서 좁힘.
5. **[기록-정정]** task 3.3의 `TestAnUnmanagedHoldingIsLabelledExactlyOnce` 개정 주장 —
   실제로는 무개정: 금지 문자열(`편입하기`·`action="/positions`)은 여전히 유효한 직접 수행
   가드이고 지정 폼은 `/settings/include`로 POST하므로 저촉되지 않는다. 주석 의도 서술만
   구계약 잔재로 남음(후속 소품 정리 대상).
6. **[노트 수용]** finding 7(중복 키 JSON에서 splice=first/Unmarshal=last 발산)·finding 8
   (선행 공백 파일의 오류 문구) — 병적 수기 편집 한정, fail 방향 무해. 후속 소품.

보완 후: console/engine/cmd 3패키지 -count=1 전부 ok.

## 독립 검증 라운드 2 (2026-07-27) — **ACCEPT**

blocker 2건 종결(사유 행렬 5/5 pin·cmd 시임 실증), P3 수정(거부 블록 위 켜기 문구) 보수
방향 확인, 기록 정정 확인. 격리 worktree @49c0b1c: 전체 go test 49패키지 0 FAIL, vet 무결,
4패키지 -race ok, check_analysis PASS, strict validate PASS. 잔여 P3/P4 노트(엔진 실행 안내
branch 무pin·목록 보존 시나리오 합성 pin·enabled 토글 audit 단언 보강·finding 7/8)는 후속
소품으로 이월.

## 사용자 결정 반영 (2026-07-27) — 확인 문구 제거

사용자 지시: 확인 문구(타이핑 승인)를 넣지 말 것 — 기존 지시의 재확인. 라운드 1 P1-3의
보완 통제(enabled 켜기 저장의 고정 문구)와 독립 검증 finding 3의 강화(거부 블록 위 켜기)를
**전부 제거**했다: settings.go 검사·템플릿 입력란·해당 테스트 2건 삭제, delta의 "확인 문구"
시나리오 삭제, proposal·design D1 재서술(Requirement 수준 수정 — 본 기록이 재리뷰 처분).
`TestEnablingIsAPlainSave`가 새 계약(세션+CSRF+클릭만)을 pin한다.

잔여 위험의 정직한 기록: 세션 토큰+CSRF 토큰을 모두 탈취한 공격자는 확인 문구 마찰 없이
config의 enabled를 뒤집을 수 있다. 방어는 루프백 전용·SameSite=Strict·프로세스별 토큰,
그리고 실효 사슬(엔진 기동+게이트 인터록 — Verified 아니면 편입 판정 자체 불가)과 저장·기동
이중 audit이다. §0.7은 유지된다(사람의 루프백 세션 클릭 = 직접 승인).

## 사용자 UX 결정 반영 (2026-07-27) — 입력 요구 제거

사용자 지시: 설정에서 자꾸 입력을 요구하지 말 것 — 손절폭은 StockOS처럼 마우스 슬라이더
(기본값 병기), 목록은 화면 기입이 아니라 포지션 버튼으로, 지정 버튼은 확인 메시지 1회로
즉시 승인. 반영(Requirement 수준 — 본 기록이 경량 재리뷰):

- 손절폭: range 슬라이더(2~20%, 기본값 5% 라벨·현재값 실시간 표시). 콘솔 무스크립트 관례를
  최소 inline 상호작용(슬라이더 표시·confirm 대화상자)까지로 완화 — 외부 자산·빌드 스텝
  금지는 유지.
- 지정 버튼: confirm 대화상자 1회 → 즉시 저장. 손절폭 미설정이면 **기본값 5%를 블록에
  명시 기록**하고 적용 사실을 안내(P2-5의 "무효 블록 기록 금지"는 유지 — 기록 블록은 항상
  유효). 종전 거부-후-입력-유도 동작 폐지.
- 지정 해제 버튼(confirm) 추가 — include 제거는 장래 편입에만 영향(A5 유지).
- 목록 텍스트 편집은 <details> 고급 접힘으로 이동(기능 보존·1차 경로 아님).
- pin: TestDesignationAppliesTheDefaultStopFraction·TestRemovingADesignationOnlyAffects
  TheFuture·TestTheStopFractionIsASlider. 콘솔 패키지 전체 green.

## 사용자 UX 결정 반영 (2026-07-27) — 라벨이 체크 상태를 반영

사용자 지시: ① 관리 열 헤더를 "관리 편입"으로, ② 미편입 행의 판정 라벨을 체크박스
상태에 연동 — 미체크 "관리 외(미편입)"·체크 "관리 편입". 반영(Requirement 수준 —
포지션 가시성 MODIFIED delta 추가, 본 기록이 경량 재리뷰):

- `positionRow.Label`에 B3 분기 추가(`!Managed && Designated` → "관리 편입") —
  Unknown(관리 여부 불명) 뒤에만 삽입해 원장 미판독 보수성 유지. FLM:
  `internal-console--positionrow.label`.
- 템플릿: 헤더 `<th>관리 편입</th>`, 지정 가능 행은 라벨을 체크박스의 라벨로 병합
  (한 행 한 문구 — `{{.Label}}`이 유일 정의, 고정 철자 이중화 제거). "편입 예약됨"
  병기는 유지 — 라벨은 예약의 표시이지 보호 성립이 아니다(각주 문단에 명시).
- 정직성 유지 근거: "관리 편입" 라벨만으로 보호가 걸렸다고 읽힐 위험은 "편입
  예약됨 — 엔진 가동 시 자동 편입(아직 손절·익절 미적용)" 병기와 각주("실제
  손절·익절은 엔진이 가동되어 대사 루프가 편입을 완료한 뒤부터")로 상쇄한다.
- pin: TestTheStatusColumnHeaderSaysAdoption·TestAnUnmanagedRowsLabelFollowsIts
  Checkbox(신규 파일 portfolio_label_test.go — RED 관측 후 GREEN). 콘솔 패키지
  128건 green. TestTheUnmanagedLabelIsSpelledOnce 불변(관리 외 철자 단일 정의 유지).
- P4(잔여): Label B5("관리 종료")·B7("엔진 관리(대기)")는 이번 변경 무접촉 기존
  분기로 직접 pin 테스트가 없다 — 다음 라운드 보강 후보.
