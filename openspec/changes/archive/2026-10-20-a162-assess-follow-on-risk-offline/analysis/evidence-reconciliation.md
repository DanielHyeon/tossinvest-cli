# a162 설계 근거 대조

- 캡처 HEAD: `3415cf89f713ccdf61604963f29da9e0d5116294`.
- AST의 source SHA-256과 당시 source SHA-256은 일치한다.
  - continuation `evaluateStage`: `7f309e45426d1cb740f2a76925d210ae74c88e2df4e05ce99fe9de5ad52d8b9b`
  - reversal `admitExistingRisk`: `525db732e564b987d6cf8d1d5ab29a2f54a18e8ed63d0600abf76c9dac692e58`
  - weekly `admitExistingRisk`: `678e87a72f784bd0e924b1bc2a256ea6c31a3c7ae857d9d2a93b154a5ece17ed`
- CodeGraph index는 당시 최신이었다. exact qualified symbol query 세 건은 결과가 없었다. 이는 API 부재 증명이 아니라 CodeGraph symbol-resolution 한계다.
- AST와 현재 source가 같은지 구현 직전에 다시 대조해야 한다. 이 자료는 설계 근거이며 qFinal, reservation, lease, journal, dispatch 또는 broker authority를 제공하지 않는다.
- full graph sync와 제품 시험은 실행하지 않았다.

## CodeGraph query 형식 보정

qualified package 이름을 붙인 세 query는 결과가 없었다. 같은 최신 index에서 unqualified `ProposeKR` function query는 세 lane 정의를 반환했고, 각 file-disambiguated node는 proposal API의 호출 trail을 반환했다. 원문은 `code-context/codegraph-query-ProposeKR-unqualified.json` 및 `code-context/codegraph-node-*-ProposeKR.txt`에 보존했다. 이 보정은 source API와 test/strategyflow adapter caller를 찾는 데만 썼으며, runtime authority 또는 external input getter의 존재를 증명하지 않는다.

US도 같은 방식으로 unqualified `ProposeUS` function query와 lane별 file-disambiguated node trail을 수집했다. 세 lane의 US proposal API 정의와 test/strategyflow adapter caller가 확인되며, 이 근거 역시 external input getter나 runtime authority를 뜻하지 않는다.
