# 보이스 ② — 수용 집합 동등성 (a112 8.5)

공통 브리프: 같은 디렉터리의 `brief.md` 를 **먼저 전부** 읽고 그 안전 규칙을 지켜라. 네 배역은 아래 하나다.

**배역: 편집 전과 뒤가 같은 입력에 같은 판정을 내는지 실행으로 대조한다.** 반드시 **커밋된 전/후의 격리 사본**에서 — 작업 트리 금지:
`0b441270^` vs `0b441270`(B1 · B2), `65444341^` vs `65444341`(A). 각 사본을 `git -C /mnt/D/Axipient/workspace/TossOS archive <commit> | tar -x -C <사본>` 로 만들고,
두 사본에 **같은 시험 드라이버**(예: `/tmp` 의 `*_test.go` 를 각 사본 패키지에 복사해 `go test -run`)를 넣어 표를 만든다.
1. B1 — `LoadProductionFamilyActivation` 의 수락/거절과 **sentinel 종류**(Undeclared / Unavailable / Revoked / Expired / ctx)가 같은가. 입력 공간: 설정 결속 열 필드 · 몸통 결속 열하나 ·
   수명 셋(해석 실패 · 순서 · 상한) · 서술자(모르는 레인 · 표 드리프트 · desired/effective · 중복 · 개수) · 파일(없음 · 핀 불일치 · 비정규 · 뒤 데이터 · 모르는 필드) · 폐기 —
   단일 변형과 **조합**(최소 두 필드 동시 · 무작위 수백)을 두 사본에서 돌려 판정 쌍을 비교하라. 메시지 문구 차이는 의도(필드명)이니 판정 · sentinel 만 비교.
2. B2 — `collectMarket` 의 닫힘 사유(Reason) · Ready · 항목 수 · RefusedCount 가 같은가(실은 활성화만 의도적으로 다르다). 앞 판 시험 fixture 를 두 사본에 같이 넣어 대조.
3. A — `Evaluate` 의 결정(phase · refusal · transitions · candidate · final · proposalID · seal · snapshotDigest)이 같은가 — RVOL 1.2~1.5 봉 · 입장 봉 · 첫 touch 를 섞은 입력 수백.
차이가 나오면 그 입력과 두 출력을 그대로 적어라. 출력 형식은 브리프 「출력」.
