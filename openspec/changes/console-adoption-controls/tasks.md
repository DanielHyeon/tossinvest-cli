# Tasks: console-adoption-controls

> 선행: proposal-freeze **적대적** 리뷰(라운드 1 REVISE 반영 완료 — review.md) 후 착수.
> 파일 표면: `internal/config`·`internal/app/engine`(adoption.go·interlock.go)·
> `internal/console`·`cmd/tossctl`·PM registry — 활성 change add-net-rr-measurement
> (risk/journal/execgw/obs/costs/measure)·verify-execution-capability와 겹치지 않는다.
> 실효: 저장은 즉시, 엔진 행동은 다음 엔진 기동 + automation gate ON(§0.7) 이후
> (Verified 아니면 ReconcileDriver 생성 거부 — 구조적 차단).

## 1. config — include_symbols와 저장 seam [T]

- [x] 1.1 RED→GREEN: `Adoption.IncludeSymbols`(+`Included()` 판정 — 시장 무관 심볼 단위,
      normaliseSymbols 공유), validate 확장(`enabled ∨ include≠∅`면 stop pct 범위 요구),
      zero-value 안전(빈 include = 현행), 거부 시 블록 전면 zeroing + Rejected 유지,
      include∧exclude 동시 등재 저장 허용(판정은 엔진 쪽 exclude 승)
- [x] 1.2 RED→GREEN: raw 블록 Load + 외과적 Save — config.json의 `engine.adoption` 키만
      교체(미지 키 보존 + **블록 밖 바이트 동일** 테스트), `os.CreateTemp` 유일 임시파일
      + flock 하의 read-modify-rename(동시 기록 테스트), 저장 전 validate 위반 거부,
      파싱 불가 파일 저장 거부(골격 생성은 os.ErrNotExist 한정), **Load는 merge·zeroing
      이전의 파일 원문 블록 + 검증 사유 별도 반환**(거부 블록 목록 유실 방지 테스트)

## 2. engine — 종목별 편입 판정과 audit (High-risk: reconciliation) [T]

- [x] 2.1 Pre-Edit 선언 + FLM(judgeHoldings·alertUnmanaged) 선행. RED→GREEN: 후보 판정
      `(enabled ∨ Included(sym)) ∧ ¬Excludes(sym)` — enabled=false·include 심볼만 편입
      후보, exclude 우선(동시 등재는 무관리 알림), 빈 include는 현행과 동일 경로, include가
      신선·Stabiliser·RECONCILE·Verified 게이트를 우회하지 않음(게이트 순서 불변).
      **알림 사유 행렬 전수 테스트**(design D2 — include 시도 실패는 실패를 말하고 "꺼져
      있다" 금지). race 테스트 포함, 기존 reconcileloop/adoption 테스트 전수 green
- [x] 2.2 RED→GREEN: `recordGateSettings`에 `engine.adoption.include_symbols` 항목 추가
      (interlock.go — Pre-Edit 선언 + FLM), 기동 audit 단언 테스트

## 3. console — 설정 화면과 종목 지정 [T]

- [x] 3.1 RED→GREEN: `AdoptionSettings` seam(Load 원문+사유/Save 두 메서드) + `/settings`
      (GET) — 원문 블록 표시(거부 사유 병기), enabled·stop pct·exclude·include 폼, 반영
      시점·상시 지정·편입 해제 부재·automation gate 편집 불가 문구, seam 미배선 안내
- [x] 3.2 RED→GREEN: `/settings/save`·`/settings/include` POST — `session0(mutating(...))`,
      CSRF 거부 시 config 무변경(카운팅 seam 단언), 서버측 재검증 거부, **enabled
      false→true 전환은 고정 확인 문구 일치 요구**(불일치 거부), include 라우트는 선행조건
      미충족(무효 stop pct) 시 무기록 거부 + 안내, include 추가 멱등, 저장 응답에 엔진
      실행 여부(마커) 안내
- [x] 3.3 RED→GREEN: positions 행별 [관리 편입 지정] — 관리 외(미편입) 보유 행에만, include
      기지정 심볼은 "편입 지정됨" 표시, 원장 부재 공지 문구 "지정만 한다"로 갱신. 가드
      개정: `consoleStateChanging`+CSRF 목록에 두 라우트, **actVerbs에 config-쓰기 어휘
      확장**(save·set·config·include·enable — 허용 목록 밖 침묵 통과 방지), 전사 문장
      갱신, 기존 직접 수행 금지 가드는 무개정 유지(지정 폼은 /settings/include로 비저촉 — review 기록). 신설 가드: AdoptionSettings 메서드 2개 고정(AST), internal/console의
      config.Service/NewService 비명명. 계좌 동사·journal 쓰기·브로커 인터페이스·
      AdoptPosition 금지 가드 무개정
- [x] 3.4 cmd/tossctl 배선: raw-JSON 외과 편집 + flock + audit append 구현 주입, nav 설정
      링크, Save 후 `engine.adoption` 밖 바이트 동일 테스트

## 4. 완료 게이트 [M]

- [x] 4.1 FLM: 수정 기존 함수 전수(judgeHoldings·alertUnmanaged·mergeAdoption·validate·
      recordGateSettings 등 diff 교차분) 산출 + `check_analysis.py` 통과, §0 검토 기록
      (보호 확대 방향만·편입 tx 무발의 불변·게이트 편집 불가·zero-value 안전·blast
      radius 수용 근거)
- [x] 4.2 PM registry allowlist 등재 + tracker fixture + `--check` 통과
- [x] 4.3 `go test ./internal/config/ ./internal/app/engine/ ./internal/console/ -race` +
      `make test`(회귀 0) + `make vet` + `openspec validate console-adoption-controls
      --strict --no-interactive` + 격리 worktree `make gate CHANGE=console-adoption-controls`
      + 독립 검증

## 5. 요구사항 개정 — 외부 종목 자동관리 메뉴 [M/T]

- [x] 5.1 기존 Story↔change 1:1을 유지하고 proposal/design/operator-console delta를
      메뉴 발견 가능성 요구로 개정한다. 새 편입·정책·주문 로직은 범위 밖으로 고정하고
      requirement-change 리뷰를 기록한다.
- [x] 5.2 RED: navigation에 `외부 종목 자동관리`와 `/settings#adoption`이 없거나,
      첫 섹션에 anchor·수동 매수·기존 공통 정책·엔진 대사 실행 경계가 없으면 실패하는
      template 회귀 테스트를 추가한다.
- [x] 5.3 GREEN: 기존 `/settings` handler/seam을 유지한 채 navigation label/href,
      adoption section anchor, 제목과 설명만 최소 변경한다.
- [x] 5.4 VERIFY: focused/race/full/vet/strict/PM 검사를 실행하고 독립 UI·안전 리뷰,
      Function Logic Map not-applicable 근거, 배포 후 무인증 브라우저 스모크를 기록한다.
      구현·배포 검증은 `verification.md`에 기록했다. change 전체의 기존 별도 컨텍스트
      ACCEPT와 이번 UI 경량 Manager 분리 리뷰 패스를 함께 적용했다.
