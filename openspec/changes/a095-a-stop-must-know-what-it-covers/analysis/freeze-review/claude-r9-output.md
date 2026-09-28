# a095 9판 — 좁은 재검증 Claude 독립 보이스 결과 (2026-09-29)

프롬프트 `claude-r9-prompt.md` (sha256 `50374920…6bd5`, 실행 사본 일치). 트리 HEAD `5d2f1b6e` export + a095 오버레이(diff 0), 끝까지 존재.
21 도구 호출. `analysis/freeze-review/` 미열람. **VERDICT: APPROVE.**

| # | 결과 | 이유 | 근거 |
|---|---|---|---|
| 1 | **RESOLVED** | 「선후 관계」가 a092 독립(범위)을 유지하면서 착수 순서는 이 절이 정하지 않는다고 적고 0.7(Manager 스케줄링)을 가리킨다. 순서를 스펙 전제로 적은 문장이 proposal · design · tasks · issues · 두 델타 어디에도 없다(전제로/깨지면/순서 조건/a092 뒤·이후/7.0 grep — 인용된 2판 이력 tasks:189와 design의 8판 철회 주석 :211-212만) | tasks.md:182, 189-199; design.md:211-212, 338; proposal.md:208 |
| 2 | **RESOLVED** | 0.7이 Manager 인용문과 글자 그대로 같다(굵게 표기 · 띄어쓰기 한 곳 차이). 깨질 때의 행동 「그 기록 없이 착수 금지」가 명시됐고 「0. 게이트 선행」에서 2.0 Pre-Edit 앞에 있다. 7.3이 승인 기록 인용을 요구한다. proposal 이관 행과 design 「남는 경합」이 같은 문장이다. §3.23의 WORKFLOW 「역할 분리」 인용이 정확하다 | tasks.md:26-27, 170-174; proposal.md:208; design.md:206-212; review.md:930-940; docs/WORKFLOW.md:74-76 |
| 3 | **RESOLVED** | N3 문장이 코드에서 참이다 — `validate`는 `!Enabled && len(Include)==0 && DefaultStopPct==0`일 때만 건너뛰고(engine.go:161) 나머지는 `ValidateStopPct`(허용 [0.02, 1.0))로 간다(engine.go:164, exitpolicy/adoption.go:49-50, 69). `mergeAdoption` B3(engine.go:279)이 `Adoption{Rejected}`로 만든다(:280). 두 ast.json 좌표가 맵과 일치. 2.6a는 사실 칸 · key만 단언하고 등급은 단언하지 않으며 두 거부 모양 픽스처를 둔다 | engine.go:157-168, 268-284; 두 ast.json; tasks.md:83-86 |

| id | 등급 | 내용 | 근거 | 제안 |
|---|---|---|---|---|
| R9-1 | P2 | 0.7은 a092 착지 전 창에서 exit goroutine의 `n.mu` 대기 증가를 **Manager 승인만으로** 수용하게 한다. 그 대기 호출은 손절 루프 안의 `o.alert`(관측 두절 · 판정 거부 · 제안 거부 · 청산 지연)이며 안전 불변식 4(손절 즉시성 — 승인으로 면제할 수 없음)에 닿는다. Q7은 원래 사용자행이었고 §3.23은 사용자 확인을 받지 않았다고 적는다. 빈도 증가는 실재하나 새 차단 자리는 없다 | exitloop.go:831, 1633, 1657, 1687; design.md:206-209; tasks.md:21, 210; review.md:935-936 | 0.7 승인 기록이 추가 대기의 상한 또는 불변식 4를 약화하지 않는 이유를 적게 하거나, 수용을 Q7 후속으로 사용자에게 보내기 |
| R9-2 | P3 | 0.7의 확인 기준이 느슨하다 — 무엇이 a092 「모든 보유자」 착지인지(그 문구는 a092 tasks.md에 없음), 승인 기록을 어디에 두는지 이름 대지 않음 | tasks.md:26, 173; a092 tasks.md `모든 보유자` 0건 | a092의 구체 산출물(task id · archive · commit)에 묶고 기록 위치(review.md §) 명시 |
| R9-3 | P3 | §3.23 인용은 Manager에게 「분할」 · 검토를 주지 「스케줄링」을 이름 대지 않음 — 스케줄링 권한은 「팀 운영 방식상」과 Manager 처분에 기댐 | review.md:933-936; WORKFLOW.md:76 | 「분할」에서 추론했다고 적거나 명시 규칙 인용 |
| R9-4 | P3 | mergeadoption 맵 :40이 0 블록을 「Enabled 거짓 · include 없음」이라 적지만 exclude도 사라진다 — 거부된 엔진의 exclude 종목은 「설정 거부」(adoption.go:405가 exclude 경우보다 앞)로 간다 | engine.go:280; adoption.go:405 | 맵에 「exclude 없음」 추가, 2.6a에 exclude 목록 가진 거부 블록 픽스처(선택) |
| R9-5 | P3 | proposal 이관 행의 중첩 굵게 표기가 렌더링을 깬다 | proposal.md:208 | 안쪽 굵게 제거 |

> **VERDICT: APPROVE** — r8 N1 · N3가 해소됐다. 남은 것(불변식 4 수용 권한 · 0.7 확인 기준의 정밀도)은 P2/P3이다.
