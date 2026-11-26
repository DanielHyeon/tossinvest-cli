# a203 tasks

> 제안 2026-11-26(worktree HEAD `2b86ea27d`, 브랜치 `docs/a203-proposal`). 사용자 승인 2026-11-26 「돌파 KR 과거 재생을 OpenSpec change 로
> 시작」. 브로커 조회(0.x · 3.x)는 단계마다 사람 승인을 다시 받음. 열린 질문은 proposal 「열린 질문」 H1~H5 · M1~M5.

## 0. 소급 깊이 실측 (사람 승인 H1 · H2 뒤)

- [x] 0.1 design §9 명령 원문 그대로, 허용 창(평일 20:00 KST ~ 미국 정규장 개장 전, 또는 주말 — 결정 M3; 「NXT 마감 20:00」은 실행 전 확인)에 G1~G7(필요하면 G8 · G9) 실행 — 공식 GET ≤ 9, 1초 간격,
      실패 · 429 에서 즉시 멈춤. 새 코드 0. — 2026-11-27 06:10:23~06:10:31 KST, GET 7 · 실패 0(G8 · G9 불필요), 근거 `analysis/depth/receipt.md`.
- [x] 0.2 `analysis/depth/receipt.md`: 요청 인자 · 응답 봉 timestamp · 개수 · 본문 sha256 · 실행 시각 · 실제 GET 수. 원문 JSON 은 저장소 밖.
- [x] 0.3 분기 기록: G3 성공 → 주 표본 100 세션 / 실패 → 깊이 안 세션 수(40 미만이면 여기서 멈추고 사람에게 보고). G7 결과로 세션 시각의
      출처(달력 · 일봉 + 기본 시각) 결정. a182 「730일」 · a128 「최소 9일」 과의 대조를 한 줄로 적음. — G3 성공(G1 2024-11-27 까지 응답) → 주 표본 100 세션 그대로, 세션 시각 출처는 달력(G7), 근거 `analysis/depth/receipt.md`.

## 1. 계약 확정 (수집 전)

- [ ] 1.1 사람 결정 H4 · H5 반영(Manager 결정 M1~M5 는 2026-11-26 반영됨 — proposal 「Manager 결정」). 세션 목록(대조 5 · 주 표본)을 확정해 `analysis/sample/sessions.txt` 와 그 지문을
      design §6 에 적고 **커밋** — 이 커밋이 사전등록 시점임. 이후 §4 · §5 · §7 숫자를 바꾸면 새 사전등록.
- [ ] 1.2 `base-commit.txt` 고정, CodeGraph 증거(`codegraph affected <files> --filter '*_test.go'`) · a166 census 대상 확인을
      `analysis/code-context/` 3 파일에 기록(단순 도구 change — 불일치 없으면 `not-applicable` 사유).
- [ ] 1.3 (가)안이면 편집 전 `dispatch` FLM/BTM 번들(`analysis/function-logic/…--dispatch/`, `scaffold_analysis.py` → `tools/logic-map`)을 만들고 커밋.

## 2. 구현 (RED → GREEN, 브로커 호출 0 — 가짜 소스만)

- [ ] 2.1 `replay` 하위 명령 골격: 시간대 거절(평일 08:00~20:00 KST · 미국 정규장, 결정 M3) · 1초 간격 · 첫 429 에서 전체 정지 · 실행당 GET 상한 ·
      캐시(저장소 밖, 재실행 GET 0). 시험: 거절 시각 경계, 429 뒤 추가 전송 0, 캐시 적중 시 전송 0.
- [ ] 2.2 캐시 리더 → `officialbars.PollClosedBars` → `SealBarSeries` 가 과거 PollAt 로 한 세션 390 봉을 빠짐없이 적재하는지 RED 부터
      (claims.md 미검증 항목). `NewEvidenceSnapshot` 이 「평가 시각 = 마지막 봉 + 2 분 이상」을 받는지 같은 RED 에서 확인.
- [ ] 2.3 재생 절차(design §3 의 3~7)와 대리 호가 P0 · S1 · S2(§4). 시험: 손으로 만든 세션 픽스처에서 ARMED 시도 i 의 ask = 봉 k+1+i 시가,
      라벨 시작 = 봉 k+2+i, 1 틱 스프레드가 1,000 원 경계에서 MaxSpreadPPM 안, 괴리 초과 → ENTRY_DRIFT_EXCEEDED → 다음 시도.
- [ ] 2.4 종목군 U-A(§5)와 판정(§7). 시험: 전일 일봉만 읽음(당일 봉을 바꿔도 종목군 불변), 1,000 원 · 명목 상한 경계, 동률 정렬,
      INSUFFICIENT 각 조건 · PASS 세 조건 경계, 부트스트랩이 a182 브랜치 `bootstrapLower` 와 같은 입력 · seed 에서 같은 값.
- [ ] 2.5 a140 `static_test.go` 세 시험이 새 파일을 포함해 통과(가안). a166 census 시험 데이터 갱신(M2 범위 — `shapes.json` `main.go||dispatch` 1건 · `sites.json` 의 a203 새 파일 자리만)과
      `go test ./tools/a140-breakout-signal-shadow/ ./internal/accountbudgetruntime/` · `go vet` · `make lint` 원문을 `analysis/verify/` 에.

## 3. 대조 검증 (사람 승인 H3 의 첫 부분 — 대조 5 세션만 수집)

- [ ] 3.1 대조 1(a140 종목군 그대로): 종목-세션 150 개의 ARMED 전 종결 단계가 a140 원장과 **전부** 일치하는지. 하나라도 다르면 봉 차이를 설명하기
      전에는 4 로 가지 않음.
- [ ] 3.2 대조 1 의 PROPOSED 비교: a140 22 건 대비 일치 수, 대리 ask − a140 ask(틱) 분포, 건당 순 수익률 차이 — P0 가정의 실측 오차로 보고.
- [ ] 3.3 대조 2(U-A 종목군): 세션별 a140 종목군과의 겹침 · 제안 수 · 건당 순 수익률을 나란히 보고.

## 4. 주 표본 재생 (사람 승인 H3)

- [ ] 4.1 주 표본 수집(허용 창, 기본 권고 주말 · 첫 429 정지 · 이어 받기). 실행마다 GET 수 · 소요 시간 · 정지 사유를 `analysis/replay/runs.md` 에.
- [ ] 4.2 재생 · 라벨 · 보고서(`analysis/replay/report.txt`): 첫 줄 호가 한계, 둘째 줄 종목군 한계, 판정(P0), 민감도 S1 · S2(판정 아님), 퍼널 ·
      거절 분포 · 결손 · 캐시 지문 · 호출 수. 원장은 `~/.local/state/tossos-a203/` 에 두고 지문(sha256 · 줄 수 · 바이트)만 커밋(결정 M4).
- [ ] 4.3 결과를 사람에게 보고. PASS 면 「실거래 전환 Story 제안 근거」로만 적고, 엔진 배선 · 서명 · 토글은 별도 Story.

## 5. 종결

- [ ] 5.1 `review.md` — 경량 리뷰 1회, 변이 원장 · 다중 적대 리뷰 `not-applicable` 사유(비례 원칙: 오프라인 측정 도구, High-risk 경로 · 게이트 판정 변경 0).
- [ ] 5.2 Manager 독립 검증 → (가)안이면 `--record-landing` → 완료 게이트(사용자 명시 요청 시) → archive → Story 경로 · PM sync.
