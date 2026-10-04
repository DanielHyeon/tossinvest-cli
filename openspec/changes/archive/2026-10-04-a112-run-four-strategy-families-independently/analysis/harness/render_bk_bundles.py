#!/usr/bin/env python3
"""a112 breakout 덮개 B2 — NewFXSeal 수리의 편집 뒤 번들(render_5222_bundles.main 재사용).

편집 전 번들: `analysis/measurements/lot-bk/pre-edit/`(HEAD 538439ed). RED: `lot-bk/red-b2.log`. 변이: `lot-bk/mutation-b2.tsv`.
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-bk/mutation-b2.tsv`"
T = "`a112_fx_seal_digest_test.go`"
FUNCS = [
    {"bundle": "internal-breakoutlane--newfxseal", "file": "internal/breakoutlane/types.go", "func": "NewFXSeal", "title": "NewFXSeal",
     "revision": "편집 전 2 분기 → 3: 역방향 갈래(B1) 안에 호출자 digest 검증(B2 — 주어진 모양 그대로)을 정규화 **앞**에 더했다. 편집 전에는 정규화 뒤 digest 를 "
                 "덮어쓰고 B3 에서 자기와 비교해 그 방향의 호출자 digest 가 검증되지 않았다(공허 검사).",
     "scen": {
         "B1": ("역방향(instrument→account) · 통화 다름 → 검증 후 정규화", f"{T} `TestAnInverseFXSealVerifiesTheCallersDigest` · `TestGstackRepairQuoteAndFXExactBoundaries`(inverse scale 6)",
                "no — 갈래 불변(몸통이 바뀜)", "yes"),
         "B2": ("**(새)** 주어진 모양의 digest 불일치 → 거절(digest 없음 · 다른 봉인 · 정규형 digest · 봉인 뒤 비율/창 변조)",
                f"{T} `TestAnInverseFXSealVerifiesTheCallersDigest` · `TestAdversarialInvalidUTF8ConfigAndFXDirectionScale`(반전)",
                f"yes — 편집 전 다섯 위조 전부 수락(`red-b2.log`), 변이 B2-1(덮어쓰기 복원) · B2-2(정규형 대조) CAUGHT({LEDGER})", "yes"),
         "B3": ("정규형 검증(방향 · 통화 · 비율 · 스케일 · 창 · 재봉인 digest)", f"{T}(정규화 결과 재봉인 · fxValid 재검증) · 기존 FX 거절 시험들",
                f"no — 갈래 불변, 변이 B2-3(재봉인 생략) CAUGHT", "yes"),
     },
     "invariants": ["역방향 봉인은 주어진 모양 그대로의 digest 가 맞아야 정규화된다 — 그 뒤 정규형으로 다시 봉인한다(fxValid 재검증과 맞물림).",
                    "생산 호출자(패키지 밖) 0 — breakout 미배선이라 생산 효과 0(grep: `FXSealInput{` · `breakoutlane.NewFXSeal` 패키지 밖 0)."],
     "state": "상태 변경 없음 — 값 검증 · 정규화.",
     "safety": ["High-risk(FX · 사이징) — 보수 방향: 이전에 수락되던 입력(digest 없음 · 틀림 · 봉인 뒤 변조된 역방향 봉인)을 거절한다. 수락 집합은 「주어진 모양으로 바르게 봉인된 역방향」 만큼 줄었다."]},
]

def _ev_bundle() -> dict:
    """반전한 EV:314 시험의 경량 번들 — 분기는 이 시험 자신의 판정 갈래라 AST id 를 그대로 쓴다."""
    file, func = "internal/breakoutlane/evaluator_test.go", "TestAdversarialInvalidUTF8ConfigAndFXDirectionScale"
    ast, _ = base.ast_for(file, func)
    ids = [b["id"] for b in ast.get("branches") or []]
    return {"bundle": "internal-breakoutlane--testadversarialinvalidutf8configandfxdirectionscale", "file": file, "func": func, "title": func + " (시험)",
            "revision": "역방향 FX 봉인 단언을 반전 — 앞 판은 digest 없는 역방향 봉인의 성공을 박아 결함(호출자 digest 미검증)을 핀하고 있었다. 이제 digest 없이는 거절, "
                        "주어진 모양으로 봉인하면 정규화 통과(Manager 판정 B2).",
            "scen": {i: ("이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화)", "이 시험 자신", "no — 시험 코드",
                         "yes — 반전 전 판은 수리 뒤 실패(`red-b2.log` 의 반대 방향)") for i in ids},
            "invariants": ["역방향 봉인은 주어진 모양 그대로 봉인돼야 정규화된다."], "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]}


if __name__ == "__main__":
    base.main(FUNCS + [_ev_bundle()], tag="a112 B2", pre="lot-bk", ledger=LEDGER)
