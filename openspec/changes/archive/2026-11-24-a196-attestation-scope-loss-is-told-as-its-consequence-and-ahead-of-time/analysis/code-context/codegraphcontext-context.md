# CodeGraphContext · GBrain (a196, 1단계)

- **CodeGraphContext: not-applicable(돌리지 않음).** 사유: 편집 표면이 한 파일(`attestation_watch.go`)의 함수 4개 + 같은 패키지 시험 함수 2개이고, 호출 사슬은
  CodeGraph 1.6.0 callers 와 AST 22개(`../cited-ast/`)로 직접 열거했다. 보조 문맥 도구는 hard evidence 를 바꾸지 못한다(WORKFLOW 권위 경계).
  「후보 누락」 위험은 `evidence-reconciliation.md` 의 grep 교차 확인으로 대신했다 — **그 채널(CodeGraphContext)은 안 봤음.**
- **GBrain: not-applicable(돌리지 않음).** 식별자(`judgeAttestation`·`acceptSupervised`·`SucceededEndpoints`·`EventCredentialExpiring`)를 정확히 알았으므로 grep 이
  맞는 도구다(CLAUDE.md GBrain 안내). **그 채널은 안 봤음.**
