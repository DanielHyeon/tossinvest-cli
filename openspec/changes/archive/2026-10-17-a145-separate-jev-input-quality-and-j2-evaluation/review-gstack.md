# a145 gstack 설계 리뷰

판정: **설계 범위 PASS**. 구현, 저장된 r2 응답의 존재, 실증 비교, 운영 적합성의 PASS가 아니다. 호환 가능한 실제 저장 r2 응답이 없으면 해당 실증 비교는 **BLOCKED**이며 합성 fixture는 비교 결과를 대신하지 않는다.

- **구조:** 불변 스냅샷과 별도 품질 sidecar, 독립 r2 lot을 사용하며 r1과 a128 원본을 보존한다. a128의 활성 capability와 정식 사양 동기화 순서는 구현·archive 전 확인이 필요하다. 범용 프레임워크나 운영 경로 추가는 제안하지 않는다.
- **비교 정확성:** 동일 state·관측·모델·기타 입력과 기존 J1 점수를 고정한다. 이 결과는 공동 r2 응답의 성능이 아닌 고정 J1 규칙 민감도다. 부분 응답 또는 provenance 불일치 시 사이클 전체를 제외한다. K=0 및 r2 생존 수가 원래 K 미만인 경우 비교를 보류하고 별도 집계한다. 후보와 사이클의 가중치도 명시되어 있다.
- **누출 및 표본:** 과거 입력의 실제 존재와 당시 가용성을 구별하고, 미래 정보가 섞인 state는 제외한다. 라벨 결손 사이클, 반복·중첩 관측, 시장·세션 및 품질 계층을 구분하며 불충분한 표본은 결론 유보한다. 가격변화는 비용 전 지표이며 인과적·미래 성과를 입증하지 않는다.
- **테스트 계획:** 결손·빈 값·미지원, 무결성, 시점, 고정 J1, 짝 결손, K 경계, 동률과 네트워크 미사용이 계획돼 있다. 이는 계획이며 테스트 실행이나 구현 검증이 아니다.
- **성능 및 단순성:** 오프라인 분석으로 한정된 설계에서 별도 성능상 차단점은 확인하지 못했다. 실제 데이터 크기와 구현 성능은 검증하지 않았다.
- **독립 리뷰:** `outputs/a145-adversarial-design-review-20261017.md`의 부분 응답과 시점 근거 지적을 수정본에서 확인했다. 새 차단 지적 없음.

이번 리뷰는 문서만 읽었으며 테스트·모델 호출·API·운영 변경을 수행하지 않았다. 요청된 범위에 따라 gstack의 bootstrap, 업데이트, 텔레메트리, 외부 전송, 설정·메모리 기록 및 별도 QA 산출물 생성은 생략했다. 저장된 실제 r2 응답의 존재 여부는 이번 리뷰에서 독립 확인하지 않았다.

## 구현 리뷰 — 2026-10-17

**범위 PASS; 전체 gate·실증 r2·운영 PASS는 아님.** 검토 대상은 `tools/a145-jev-offline-analysis/analysis.py` SHA-256 `0e433cd95a58fbb285f6ae4c21c3b446314a2f1237ec690f8fc28946572f8bae`, `test_analysis.py` SHA-256 `50eb29bad292cdbe567195ffd327b854d5386866bbfdda4533160e060a91d744`, README SHA-256 `6976821e6bf6aa1f86dba5ff1e7a01808ea12864257450aac39785fb97ce3a63`다. 변경 시 재검토가 필요하다.

- 60분 외 horizon이 60분으로 표시될 수 있는 문제와 J1/J2 모델 버전 불일치·혼합 문제를 발견해 작성자가 수정했다. 음성 fixture가 포함됐다.
- 경로 검증, 입력 무결성·label join, cycle별 선택 및 결손 제외, r2 전 후보 짝 검증과 provenance, source write 및 network 경계를 검토했다. 남은 명확한 차단 지적은 없다. 제공된 응답 기록의 정합성은 실제 모델 응답의 진정성 증명이 아니다.
- 실제 CLI 출력 `/tmp/a145-final-cli.stdout`은 `r2 status: BLOCKED`, exit는 0이다. 최종 결과 manifest는 state 4,710개, judgment·label 각 9,420개다. `/tmp/a145-focused.log`의 오프라인 테스트 13건 및 exit 0을 확인했다. `/tmp/a145-source-before.txt`와 `/tmp/a145-source-finalcheck.txt`는 동일했다. 기존 분석 자료 재계산 결과이며 r2 실증 결과가 아니다.
- 독립 리뷰 `outputs/a145-adversarial-implementation-review-20261017.md`는 검토 범위 PASS, 세 세션의 네 지표를 기존 분석과 대조해 차이 0으로 보고했다. 같은 리뷰는 최종 gate 및 운영 판단을 범위 밖으로 명시한다.
- gstack checklist 적용. 요청된 범위에 따라 자동 수정, 업데이트, telemetry, 외부 전송 및 별도 설정·메모리 기록을 생략했다. 테스트와 CLI는 작성자의 실제 로그를 확인했으며 이 리뷰에서 재실행하지 않았다. 전체 repository gate는 별도 담당자 검증 대상이다.

- 추가 fixture에서 K<10 동률, 중복·고아 레코드, socket 차단 CLI 및 recorded 응답 경계를 확인했다. 당시 `sdd-check`는 CodeGraph hard index stale/status 미확인으로 exit 2, `make gate`는 2/11에서 미완료 task 7건으로 exit 2다. 전체 검증 PASS가 아니다.

## SDD 동기화 timeout 수정 리뷰

**수정 범위 PASS; 전체 gate는 별도 확인 필요.** `tools/sdd/sdd_sync.py` SHA-256 `7e015db1872b33e3dc6c7e867d600a83c36d933b9052b4b769a9fdc0533dc60a`, `tools/sdd/test_sdd_sync.py` SHA-256 `15df5c5dd17865c5c8f95923dc2f98576048d4a7cdf198fae2982a3a8173658d` 기준이다. Advisory stats timeout만 포착해 실패 목록에 남긴 뒤 hard 성공 기록에 도달한다. Hard 실패는 성공 집합에 추가되지 않으며 빈 집합은 기록하지 않는다. 정상 stats 분기는 유지된다. Hard 성공·실패 각각을 포함한 테스트의 RED exit 1과 GREEN exit 0을 기존 로그에서 확인했다. 수정은 격리 작업트리 기준이며 향후 변경 시 재검토가 필요하다.
