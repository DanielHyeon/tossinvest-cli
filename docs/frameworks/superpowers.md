# Superpowers TDD discipline

모든 동작 변경은 다음 네 단계의 증거를 남긴다.

1. RED: Requirement/branch와 연결된 테스트가 기대 이유로 실패한다.
2. GREEN: 최소 구현으로 그 테스트가 통과한다.
3. REFACTOR: 동작·안전 불변식을 보존하며 중복과 복잡도를 줄인다.
4. VERIFY: 대상 패키지, 전체 테스트, vet, 필요 시 race/crash/recovery를 실행한다.

기존 함수 내부 편집은 RED 전에 Function Logic Map과 Branch Test Map이 있어야 한다.
문서·설정 작업은 정적 검사 또는 smoke를 RED로 사용한다.

## Completion evidence

```text
Requirement/branch:
RED command + observed failure:
GREEN command + result:
Refactor:
Final verification:
```
