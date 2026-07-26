# TossOS SDD memory

파일 기반 memory가 정본이고 GBrain은 보조 검색 계층이다.

```bash
scripts/memory-recall.sh "키워드"
scripts/memory-retain.sh docs/memory/episodes/EP-TOS-....md
python3 scripts/memory_index.py promote EP-TOS-... --evidence "검증 근거"
```

`retain`은 항상 `episodic`으로 기록한다. 테스트·실측·리뷰 근거가 생긴 뒤에만
명시적으로 `canonical` 승격한다. 시크릿, 개인정보, 검증되지 않은 수익성 결론은 저장하지 않는다.
