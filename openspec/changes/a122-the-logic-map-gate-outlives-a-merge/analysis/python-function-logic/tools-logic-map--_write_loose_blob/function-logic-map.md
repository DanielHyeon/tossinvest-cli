# Function Logic Map: `_write_loose_blob` (Python, a122 task 7.5.25 — 새 함수)

`ast.after-7.5.25.json` — 분기 0. blob 을 임시 객체 저장소에 loose 객체(`"blob <크기>\0" + 바이트` 를 zlib 로)로 쓴다.
`git hash-object -w` 대신 쓰는 까닭: git 이 파일을 **다시** 읽지 않아야 판정한 바이트가 git 에 건넨 바이트이고,
편집 파일 수만큼 프로세스를 띄우지 않는다(스냅숏의 프로세스는 편집 수와 무관하게 셋이다: `rev-parse` · `ls-files` ·
`update-index`). 형식이 틀리면 git 이 객체를 못 읽어 diff 가 실패한다 — 변이 `AH18` 이 그것을 잰다.
