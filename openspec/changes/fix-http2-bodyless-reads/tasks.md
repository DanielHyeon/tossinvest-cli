## 1. RED

- [x] 1.1 실제 TLS HTTP/2 bodyless GET/HEAD/stream 회귀 테스트를 추가한다.
- [x] 1.2 known-length와 unknown-length body가 400으로 거부되는 RED를 추가한다.

## 2. GREEN

- [x] 2.1 server request body 판정을 protocol-independent `ContentLength` 계약으로 수정한다.
- [x] 2.2 focused/full/race/vet/SDD 검사를 통과한다.

## 3. 배포 검증

- [ ] 3.1 main push CI를 통과한다.
- [ ] 3.2 Compose 재배포 후 TLS HTTP/2 REST/SSE 카나리를 통과한다.
