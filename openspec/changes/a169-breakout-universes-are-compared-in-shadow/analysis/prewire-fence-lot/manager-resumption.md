# 절차 예외와 구현 재개 판정

2026-10-31 Manager. 첫 소스 수정 전에 `9e6cd08b8`로 원본 AST와 요약만 고정됐고 완전한 분기별 FLM/BTM은 없었다. 소스/시험 초안 작성 뒤 이를 발견해 추가 편집·시험을 중지했다. 사전 준수로 소급하거나 이력을 다시 쓰지 않는다.

사후 보완한 canonical bundle은 원본 `b0f7f1d93` SHA와 AST41개 B1–B41, 흐름 G01–G17의 교차표, 오류/조기반환/잠금/소유권 및 새 F01–F07과 명명 시험을 구별한다. 독립 reviewer가 원본 SHA·모든41개 ID·시험 이름 존재 및 미검증 표시를 대조했다. 이는 실행 coverage나 시험 PASS가 아니다.

현재 지도와 예외 기록을 먼저 커밋한 뒤 구현을 계속하도록 승인한다. 최종 코드에는 current AST/분기/시험 지도를 별도로 갱신하고 원본 사전 AST를 보존한다. historical gap/ring·rollback·dirty·owner·Begin/Finish/cleanup·cached/adopt401·동시 waiter 변이/race 등 계획의 미완 검증은 그대로 남는다. 기존 함수별 미실행 경계도 완료 보고에서 숨기지 않는다.

시험 seam은 기존 Transport의 nil-default private hook 두 개만 승인했다. 초기 fenced sticky 검사 직후와 Acquire 직후이며 공개 setter·production assignment는 없다. build tag로 제외되는 필드라는 주장은 하지 않는다. 잠금 경합은 channel과 통제된 TryLock으로 증명하고 sleep을 안전 증거로 사용하지 않는다.

이 재개는 개발 작업 허가이며 병합·푸시·archive·실계좌·운영·전체 tests/gate 승인이 아니다. 최종 독립 검증과 Manager 판정을 다시 거친다.
