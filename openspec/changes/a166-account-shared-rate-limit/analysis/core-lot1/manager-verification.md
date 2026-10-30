# a166 첫 내부 모듈 Manager 판정 — 2026-10-31

**부분 병합 승인.** 검증 대상 `cbf53d45294082f9149ef3663f44829574aab7d0`의 사용되지 않는 `internal/accountbudget` 모듈에 한정한다. 기존 official/engine/CLI 배선과 운영값은 이 로트에서 바뀌지 않았다. 전체 a166 완료·archive·운영 활성화 승인이 아니다.

Manager는 `abaa22aaa`의 파일 검증·상태 갱신·큐·소유권·전송 코드를 손으로 읽고, `dc149bc51` 및 `cbf53d452`의 후속 차이와 원문 영수증을 대조했다. AST가 모든 분기를 증명했다는 판정은 아니다. 개발 중 발견한 drain 전 전송 여부 snapshot, full dead queue 등록 실패, cleanup 오류 폐기, proxy 선택 및 transport Clone 부작용은 제품 시험·변이와 수정으로 다뤘다. 원본 CodeGraph 미초기화와 미검증 환경은 구현자 기록대로 남긴다.

## 확인한 증거

- 최종 core 독립 Terra: 일반 9.638초, race 12.894초 PASS. 원문 `/tmp/a166-core-final-independent-normal.log`, `/tmp/a166-core-final-independent-race.log`를 직접 읽었다. 독립 리뷰는 같은 SHA의 delta에서 P0/P1을 발견하지 못했다.
- 작성자 최종 core: `template-final-go125-normal.log` 10.032초, `template-final-go125-race.log` 12.660초, `template-final-go126-normal.log` 10.045초, `template-final-go126-race.log` 12.707초 PASS 원문을 읽었다. Go 1.25 결과는 작성자 실행이며 최종 SHA의 별도 독립 Go 1.25 재실행으로 표시하지 않는다.
- 공식 클라이언트 후속 로트의 실제 a112 clone/외부 transport alias 시험이 같은 core SHA에서 race 1.528초 PASS. `/tmp/a166-lot2a-corefix-race.txt`를 직접 읽었다. 이 영수증은 후속 로트 작업 중인 시험이며 공식 클라이언트 전체의 고정 SHA 완료 증거는 아니다.
- 핵심 변이 5건 및 Clone 복원·임의 TLSNextProto 허용 변이의 기대 RED는 `review.md`, `template-review.md`와 해당 원문에 있다. gstack checklist 수동 범위 검토를 수행했으며 전체 자동 교차 파이프라인 완료로 주장하지 않는다.

## 후속 조건

실제 모든 참여자의 동일 namespace/inode/instance, 신뢰되는 clock.System 계열의 생산 배선, official 권한 봉인과 현재 주문 attempt 판정, replay/probe, 모든 429 기록·폴링 hint, 운영 환경 검증은 뒤 로트다. 서로 다른 root의 복사본이나 임의 fake clock을 넣고 계좌 전체 상한을 보장한다고 주장하지 않는다. tmpfs 재부팅 영속성·실제 전원 장애·다른 OS 실행·전체 저장소 시험/전체 gate·실계좌는 미검증이다. a087 nil과 a169 측정 보류를 유지한다.

Terra는 이 승인과 독립 영수증을 저장하고 tasks 1.1~1.4 및 PM 진행표를 실제 범위대로 갱신한 뒤 non-force merge/push를 진행한다. 모든 a166 작업을 완료로 체크하거나 archive하지 않는다. 후속 코드가 바뀌면 해당 영향 범위를 다시 검증한다.
