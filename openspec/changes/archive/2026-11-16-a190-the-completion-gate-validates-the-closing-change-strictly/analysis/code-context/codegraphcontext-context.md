# CodeGraphContext context

- not-applicable: 바뀌는 표면이 셸 스크립트 1개(`tools/gate.sh`)·새 Python 스크립트 1개·시험·문서뿐이고,
  CodeGraph 와 같은 이유로 셸·Makefile 호출 관계는 이 도구의 그래프 밖이다. 문맥 확장은 `git grep` 전수(HEAD)로 대신했고
  그 결과는 `codegraph-baseline.md` 「한계와 대체 증거」 표에 있다. `codegraphcontext update` 는 돌리지 않았다.
