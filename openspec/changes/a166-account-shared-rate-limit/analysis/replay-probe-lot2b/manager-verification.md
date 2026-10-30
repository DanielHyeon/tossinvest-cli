# 재전송·단발 측정 전송 Manager 판정

2026-10-31, 후보 `6099998ed54af9467800a4dab901264d92c14f4d`와 후속 `7c0aa64e0f16940d490137573888a1fb86d5f8fa`를 검토했다. 명시 transport가 없는 required replay를 헤더 준비 전에 거절하도록 수정한 최종 후보의 부분 병합을 승인한다. OFF 경로의 nil/default 동작은 유지한다. 이 승인은 a166 전체 완료·운영 활성화·archive 승인이 아니다.

독립 Terra의 선택 replay normal/race 5.443/29.990초, probe normal/race 2.177/3.201초와 최종 nil delta normal/race 0.018/1.030초 PASS 원문을 직접 확인했다. 작성자의 nil fallback 복원 변이는 새 시험에서 기대한 RED였다. 구현 원문과 delta는 손으로 읽었으며 AST/분기 증거 범위는 작성자의 pre-edit·post 기록에 따른다. 새 P0/P1 발견 없음.

Probe 정적 검사는 명시된 주문·계좌 메서드 금지와 `official.*` selector 허용 목록을 보존한다. 기존에도 client 인스턴스의 모든 메서드를 제한한 것은 아니므로 시세 메서드의 잠재적 허용을 이번 회귀로 분류하지 않는다. 현재 새 인스턴스 호출에서 별도 endpoint 호출은 발견하지 못했다. AuthHeaders 호출 방식에 대한 낡은 주석은 실제 concrete client 사용에 맞춰 정정한다.

Terra가 독립 검증 원문을 보존하고 위 주석을 정정한 뒤, 현재 main을 통합하고 영향 범위 strict·PM·diff 검사 후 non-force commit/merge/push한다. 기본 실행 경로는 여전히 OFF이며 runtime factory와 모든 호출자 연결은 후속 작업이다. 전체 시험·전체 gate·실계좌·브로커·배포·활성화는 미실행이다. gstack은 범위 지정 수동 체크리스트 검토이며 전체 자동 파이프라인을 실행했다는 뜻이 아니다. impeccable: N/A — 전송 및 검증 문서 변경으로 UI가 없다.
