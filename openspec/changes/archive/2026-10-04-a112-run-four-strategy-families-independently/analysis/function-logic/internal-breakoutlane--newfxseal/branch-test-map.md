# Branch Test Map: `NewFXSeal`

- Source SHA-256: `138a1b686e25353d1f3f6b4f191b626355c457f3e0f3771a004c4afa6f13fe82`; AST branch locations are authoritative.
- Revision: **modified (a112 B2, 2026-10-01).** 편집 전 2 분기 → 3: 역방향 갈래(B1) 안에 호출자 digest 검증(B2 — 주어진 모양 그대로)을 정규화 **앞**에 더했다. 편집 전에는 정규화 뒤 digest 를 덮어쓰고 B3 에서 자기와 비교해 그 방향의 호출자 digest 가 검증되지 않았다(공허 검사).
- 편집 전 번들: `analysis/measurements/lot-bk/pre-edit/internal-breakoutlane--newfxseal/`. 변이 원장 `analysis/measurements/lot-bk/mutation-b2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 155:2 — 역방향(instrument→account) · 통화 다름 → 검증 후 정규화 | `a112_fx_seal_digest_test.go` `TestAnInverseFXSealVerifiesTheCallersDigest` · `TestGstackRepairQuoteAndFXExactBoundaries`(inverse scale 6) | no — 갈래 불변(몸통이 바뀜) | yes |
| B2 | if at 158:3 — **(새)** 주어진 모양의 digest 불일치 → 거절(digest 없음 · 다른 봉인 · 정규형 digest · 봉인 뒤 비율/창 변조) | `a112_fx_seal_digest_test.go` `TestAnInverseFXSealVerifiesTheCallersDigest` · `TestAdversarialInvalidUTF8ConfigAndFXDirectionScale`(반전) | yes — 편집 전 다섯 위조 전부 수락(`red-b2.log`), 변이 B2-1(덮어쓰기 복원) · B2-2(정규형 대조) CAUGHT(`analysis/measurements/lot-bk/mutation-b2.tsv`) | yes |
| B3 | if at 165:2 — 정규형 검증(방향 · 통화 · 비율 · 스케일 · 창 · 재봉인 digest) | `a112_fx_seal_digest_test.go`(정규화 결과 재봉인 · fxValid 재검증) · 기존 FX 거절 시험들 | no — 갈래 불변, 변이 B2-3(재봉인 생략) CAUGHT | yes |
