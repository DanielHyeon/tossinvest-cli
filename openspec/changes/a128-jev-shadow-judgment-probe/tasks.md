# a128 Tasks

## 0. 사람 관문

- [ ] 0.1 TypeSafe API 키 조달 — 사람이 발급해 `TYPESAFE_API_KEY` env 로 공급
      (저장소·기억·로그에 키 저장 금지). 키 전까지 1.x 구현은 mock 으로 진행 가능.
- [ ] 0.2 외부 API 호출량 기본값 승인 — 세션당 샘플링 간격 10분 × 상위 10종목 × 판단 2
      (세션당 ≤ 약 780 판단 호출 × 4세션, 비용은 키 발급 시 요금표로 확인해 여기 기록).
- [ ] 0.3 (선택) WTS 뉴스 소스용 `tossctl auth extend` — 사람 실행. 안 하면 뉴스 결손
      모드로 측정(원장에 명시됨).

## 1. 구현 (Opus 팀메이트 — 새 파일만, 생산 무접촉)

- [x] 1.1 TypeSafe API 계약 동결 — docs.typesafe.ai/api.md·primitives/noul.md 를 읽고
      요청/응답 계약을 원문 인용으로 `analysis/typesafe-api-contract.md` 에 기록.
      추측 필드 금지(계약은 영수증에서).
      — 2026-10-07 읽음(7 페이지). p = `answers.<id>.noul`. 문서 미기재 7항은 계약 문서 6 절.
- [x] 1.2 `tools/a128-jev-shadow-probe/` — judgment client(stdlib HTTP, env 키,
      타임아웃·재시도), state 조립기(공개 시장 데이터만 — 계좌 필드 차단),
      수집 루프(rankings→stocks→prices/orderbook, rate limit 백오프, 세션 경계 준수),
      원장 writer(JSONL append, 행 2종), 라벨러(h분 뒤 /prices 재조회),
      보고서 생성기(p 구간표·구간당 n<30 보류 규칙).
      — 설계와 다른 점: (a) 원장 행이 **3종**(judgment·label·**gap**) — 랭킹/시세 읽기 포기·판단
      API 포기처럼 판단 행이 생기지 못한 결손은 실을 행이 없어 침묵 생략이 되므로(spec 「결손은 행에
      명시」). (b) 뉴스는 이 로트에서 미배선 — state `news: "absent"` + `missing_inputs` 와 행
      `collection_gaps` 에 사유 명시(사람 관문 0.3 전). (c) official 은 복제가 아니라
      `internal/official` 을 import(L1c 선례) — 토큰 캐시·401 adoption 을 공유해야 엔진과 토큰
      전쟁(a082)이 안 남. 대신 브로커 표면은 읽기 GET 5개짜리 `marketSource` 인터페이스로만
      다루고 주문·계좌 이름은 AST 시험이 금지. (d) J1·J2 를 한 요청에 묶음(noul.md 권고) → 판단
      **요청** 수는 0.2 추정의 절반. (e) 판단 시각 + h 가 장 마감을 넘으면 더 판단하지 않음(라벨이
      장 밖 가격을 재지 않게) — KR 기본값에서 세션당 34 사이클. 모델 기본값 `jev-1.13.0` 버전 핀.
- [x] 1.3 시험 — mock TypeSafe 서버·mock 시세 픽스처로: 계좌 필드 차단, 뉴스 결손 명시,
      원장 행 스키마, 라벨 결손 사유, 보고서 보류 규칙, 키 부재 시 시작 거부.
      `go vet`·`make lint` 통과. FLM: not-applicable(새 파일만, High-risk 무접촉 —
      비례 원칙).
      — 영수증(2026-10-07, go1.26.5, 실 API 호출 0 — httptest 만):
      `go vet ./tools/a128-jev-shadow-probe/...` rc=0 · `go vet -tags tossos_testseams …` rc=0 ·
      `gofmt -l tools/a128-jev-shadow-probe/` 빈 출력 · `rtk proxy go test -count=1 ./tools/a128-jev-shadow-probe/...`
      rc=0 (`-v`: 최상위 29 PASS + 하위 7 PASS, FAIL 0) · 같은 명령 `-race` rc=0.
      `make lint` 전체는 돌리지 않음 — 그 gofmt 범위는 `./cmd ./internal ./tools/logic-map` 라 이 디렉터리를
      안 보고, 나머지는 `go vet ./...` 이므로 위 두 vet + gofmt 로 이 패키지 몫을 대신함.
      변이 점검(제자리 sed → 복원 sha256 대조 OK, 무변이 대조군 GREEN): 바이트 가드 삭제 2(judge·storeState)·
      기한 초과 검사 삭제·보류 기준 29·4xx 재시도·뉴스 결손 삭제·키 검사 우회·horizon 창 무시·모든 오류 재시도·
      라벨 불변식 삭제·O_TRUNC·J2 적중 반전 = 12/12 CAUGHT. (모든 오류 재시도는 처음 SURVIVED →
      503 비재시도 단언 추가 후 CAUGHT; 바이트 가드는 타입이 이미 막아 행동 시험으로 안 잡혀 AST 고정으로 못 박음.)
- [x] 1.4 팀메이트 커밋(푸시 금지) — 허용 경로: `tools/a128-jev-shadow-probe/`,
      `openspec/changes/a128-jev-shadow-judgment-probe/`. 그 밖 0.
      — 451ddba4, 14파일 전수 허용 내, Opus 트레일러 확인(Manager 배터리 2026-10-07).

## 2. 리뷰·착지

- [x] 2.1 gstack 리뷰(경량 — 비례 원칙) + Manager 검증(경로 전수·트레일러·시험 재실행).
      — 2026-10-07 Manager /review: 구현 6파일 정독, critical 0 · informational 2
      (① input_tokens 요청 단위 의미 — 주석 명기로 수리 d80ae85a ② 원장 꼬리 torn line 시
      readLedger 가 전체 거부 = fail-closed, 수리 불요·운영 시 꼬리 절단으로 복구).
      Review Army·적대 다중 리뷰: not-applicable(비례 원칙 — High-risk 무접촉·새 파일만).
      Manager 재실행 영수증: go vet rc=0 · go test rc=0 · gofmt 빈 출력.
- [x] 2.2 Manager 푸시 — 451ddba4 + d80ae85a (e67c514a..d80ae85a).

## 3. 측정 (2일 — 사용자 지시 2026-10-07)

- [ ] 3.1 1일차: KR 세션 + US 세션 각 1회 프로브 실행(에이전트 가능 — read-only).
      실행 로그·원장 보존.
- [ ] 3.2 2일차: 동일. 라벨러 완주 확인(마지막 판단 + h분까지).
- [ ] 3.3 원장·state 파일 커밋(계좌 데이터 0 확인 후).

## 4. 판정·종결

- [ ] 4.1 calibration 보고서 생성 — J1 예측력 유무(평평하면 J1 폐기 기록),
      J2 p≥0.9 구간의 적중률·표본 수, 2일 한계 명기.
- [ ] 4.2 사람 결정 기록 — 90% 임계 채택/기각/보류, 후속 change(수동 보조·엔진 veto
      gate) 착수 여부. 후속 change 범위 초안을 review.md 에 남긴다.
- [ ] 4.3 gate → archive(Opus) → Story 경로 갱신.
