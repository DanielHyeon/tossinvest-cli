# 사전 분기 시험 지도

- 기존 OFF retry/auth/freshness test golden; required nil/unknown/ref·seq drift·foreign/uncomparable guard fn0, client 결속 제거 변이.
- key 길이 구분, 재시작·서로 다른 factory 동일 key, 다른 account/kind/instance 분리; 정기·보호·adoption·재편입 공유, 기존 재편입 권한 보존.
- shared hint valid ≤30 / invalid / absent / >30 / HTTP date, token429·503, 429+body error 및 nil 결과·local fault에서 fn 재시도0/freshness0, 만료 후 복원. AuthFatal+collector 조건에서도 latch·durable escalation 보존.
- 실제 official TLS token200 Finish fault→GET0, GET401 Finish fault→token/reGET0 및 token429→GET0; transport prewire latch 제거 변이 RED, pretransport body Close exactlyonce 및 original/ErrNotSent 보존. 독립 collector 격리와 exitSideRetrier 복사.
- fake 자격·임시 path, local TLS exact DialContext, lo-only/no-route namespace. Go1.25 선택 normal/tagged race, scoped gstack·독립 리뷰. 운영/브로커·전체 tests/gate 금지.
