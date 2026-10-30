# 원문 log 무손실 보존

e1df1f1c9의 원문 바이트를 .raw.gz에 보관한다. 표시용 .log는 줄끝 공백만 정리했다. gzip 해제 후 SHA-256 검증.

- a166-egress-local-proof.log: aad150070970149a0402212f25fb8ea73d90210317ba327059ee129807094af1
- a166-reconcile-bypass-mutation.log: c07680fe50609fdeadf6c6fac2cb5c20a5ec5ea4d842dadaf125f23c44424fe1
- a166-reconcile-local-green.log: 5c85c647875c9a37acba0e4e50b6ffb6529092217b5525f1b33244ad83703eb2
- a166-reconcile-required-green-v2.log: 4ce5f5d55445afb6224527804a9474256133fdc9be8ae6c556eec5a68eda7d0d
- a166-reconcile-required-green.log: 15cef71d00a3226ffe9c968ade59fc2ed729064f97889200fc28a0f7c53fd3b8
- a166-reconcile-required-tag-race.log: 09e7bca5cf5c8226d35a8d1854008c64516df01657666cb773aab8fbeb5c4473

검증 흐름 결함: e1df1f1c9 생성 명령은 `git diff --cached --check` 실패에도 다음 commit을 실행했다. 검사 실패를 통과로 보고하지 않았고 이 후속 정리에서 `&&` 성공 조건으로 연결했다. 앞으로 검증 실패 시 후속 mutation을 실행하지 않는다.
