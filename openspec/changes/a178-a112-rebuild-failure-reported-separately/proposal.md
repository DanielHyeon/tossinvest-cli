# a178 — a112 M-B1 수집기: prescribed rebuild 실패 원인을 구분해 보고 (정의)

## Why
2026-10-27 a175 완료 gate 가 1차 ⑦ `make test`, 2차 ⑧ `make test-seams` 에서 `tools/a112-mb-us-source` 의 e2e 시험으로 실패했다(`TestTrimpathBuiltCollectorCompletesOfflineIdentityPreflightBeforeReader`, `TestTrimpathBuiltCollectorProcessHoldsBeforeExternalReadOnMissingCache`). 메시지는 둘 다 `prescribed build identity: running executable does not match the prescribed rebuild` 였다. 같은 시험을 단독으로 돌리면 PASS(90s)이고 gate 안에서는 200~250s 가 걸렸다.

근본 원인 (코드 원문, HEAD `526650e50`):
- `run` 은 수집 전체에 실제 시각 120초 시한 `collectorCtx` 를 건다(`main.go:19` `overallBudget`, `:82`). 그 안에서 `verifyPrescribedBinary(collectorCtx, …)` 가 빈 GOCACHE 로 prescribed rebuild 를 돈다(`main.go:116`, `identity.go:487`).
- `identity.go:488-489` 는 세 원인을 한 문구로 합친다: ① `builder.rebuild` 오류(시한 초과 `context deadline exceeded` 포함), ② 빌드 명령 불일치, ③ digest 불일치. 그래서 **부하 때문에 시한을 넘긴 것과 실제 신원 불일치를 구분할 수 없다**. 진단이 거짓 방향을 가리킨다.
- 120초 시한 자체는 spec `breakout-retest-strategy-lane` 「M-B0/M-B1 measurement는 production evidence authority와 분리된다」(101행)의 SHALL 이다: Go 배포 스냅숏과 prescribed identity/rebuild 는 같은 120초 시한 안에서 완료되어야 한다. 사람 결정(2026-10-27, B-1): **spec 과 시한은 바꾸지 않는다.**

## What Changes
1. `verifyPrescribedBinary` 의 실패를 셋으로 나눈다. 셋 다 지금처럼 오류(→ 수집기 HOLD)이고 판정은 바뀌지 않는다.
   - rebuild 오류: `prescribed rebuild failed: %w` — 원래 오류를 감싼다. `errors.Is(err, context.DeadlineExceeded)` 가 참이 된다.
   - 명령 불일치: `prescribed rebuild used an unexpected command`
   - digest 불일치: 기존 문구 `running executable does not match the prescribed rebuild` 유지
2. 시험: 세 갈래 각각의 단위 시험(가짜 builder). 시한 초과는 `context.DeadlineExceeded` 를 돌려주는 가짜 builder 로 재현하고 `errors.Is` 와 문구를 단언한다.
3. gate 관례 기록: ⑦⑧ 이 a112 e2e 시험으로 실패하고 메시지가 `prescribed rebuild failed: … deadline exceeded` 이면, 원인이 부하이므로 조용할 때 1회 재시도한다. 메시지가 digest 불일치면 재시도하지 않고 조사한다(`docs/WORKFLOW.md` 함정 절 — 문서 1줄, 이 change 커밋에 포함).

## Scope / Non-goals
- 범위: `tools/a112-mb-us-source/identity.go:verifyPrescribedBinary`, 같은 패키지 시험, `docs/WORKFLOW.md` 한 줄.
- 비목표: 120초 시한·`overallBudget`·spec 101행 변경(사람 결정 B-1), e2e 시험을 gate 의 별도 단계로 옮기기(B-2 기각), 시험 전용 시한 seam(시험이 prescribed 빌드 명령으로 만든 바이너리의 digest 를 rebuild 결과와 비교하므로 태그·환경변수 seam 은 그 비교를 깨거나 운영 바이너리에 SHALL 을 끄는 스위치를 만든다), 수집기 동작·HOLD 조건 변경, gate ⑥ CodeGraph 15초 probe(별개 — 워밍업 관례로 처리).

## Risk
Low. 측정 도구(운영 caller 0, spec 「production evidence authority와 분리」). 바뀌는 것은 오류 문구와 감싸기뿐이고 성공·실패 판정은 같다. 시한을 넘기면 지금처럼 HOLD 다.

## 완료 기준
Story acceptance(`STORY-TOS-a178`).

## not-applicable / 적용
- FLM/BTM: **적용**. 기존 함수 `verifyPrescribedBinary` 본문을 고친다 → gate ⑤ 번들 1(편집 전 작성). 기존 시험 함수를 고치면 그것도 번들 대상. 근거 AST: `analysis/evidence/verifyPrescribedBinary.ast.json`(source sha `de8e6176…`, :470-495, 분기 B1~B7·return 7 — B6(:488) 이 세 원인을 OR 로 합치는 자리).
- 변이 원장·다중 적대 리뷰: 비례 원칙(도구 진단 문구, 게이트 판정·High-risk 경로 아님). 변이 1회: B6 분리 뒤 rebuild 오류 갈래의 `%w` 를 `%v` 로 바꾸면 `errors.Is` 시험이 FAIL 하는지 확인.
