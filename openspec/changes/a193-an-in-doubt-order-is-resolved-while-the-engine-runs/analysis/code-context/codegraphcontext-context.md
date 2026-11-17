# CodeGraphContext · GBrain (a193, 1단계)

- **CodeGraphContext: not-applicable(이번에 돌리지 않음).** 사유: 영향 표면이 새 파일 하나 + 기존 함수 1(배선 슬라이스 원소) + 시험 함수 1이고, 호출 사슬은
  CodeGraph 1.6.0 hard evidence 와 AST 40개로 직접 열거했다. 보조 문맥 도구가 hard evidence 를 바꾸지 못하므로(WORKFLOW 권위 경계) 시간 대비 이득이 없다고 판단.
  「후보 누락」 위험은 아래 evidence-reconciliation 의 grep 교차 확인으로 대신했다 — **그 채널(CodeGraphContext)은 안 봤음.**
- **GBrain: not-applicable(돌리지 않음).** 의미 검색 대신 정확한 식별자(`Resolve`·`PendingAttempts`·`ReasonUnresolvedInDoubt`·`EventOrderInDoubt`)를 알았으므로
  `rg` 가 맞는 도구다(CLAUDE.md GBrain 안내). **그 채널은 안 봤음.**
