#!/usr/bin/env python3
"""task 7.5 변이 — 새 배치 fetch 의 갈래마다 **실제로** 빨개지는 시험이 있는가.

사본 대상 + **무변이 대조군**이 먼저다 ([[mutation-revert-needs-the-right-baseline]] ·
[[mutation-must-reach-the-thing-under-test]]). 대조군이 초록이 아니면 멈춘다.
"""
import shutil
import subprocess
import sys
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
WORK = SP / "75_mut_work"
# 스위트 **전체**를 돈다 (task 7.5.2) — 226개가 50초 안팎이라 고른 부분집합의 이득이 없고, 고르면
# 새 시험 클래스를 목록에 안 넣는 것만으로 변이가 "살아남는다".
SUITE = ["test_check_analysis"]

MUTATIONS = {
    # --- 배치 프레이밍 (7.5 · 1.4 · 7.5.1 — 파서를 7.5.1 에서 엄격하게 다시 썼다) ---
    "T1_output_not_nul_framed": [(
        '["git", "cat-file", "--batch", "-Z"],', '["git", "cat-file", "--batch", "-z"],')],
    "T2_line_framed_both_ways": [
        ('["git", "cat-file", "--batch", "-Z"],', '["git", "cat-file", "--batch"],'),
        ('cwd=root, input=b"".join(spec + b"\\0" for spec in asked),',
         'cwd=root, input=b"".join(spec + b"\\n" for spec in asked),'),
        ('        end = data.find(b"\\0", position)', '        end = data.find(b"\\n", position)'),
    ],
    "T3_any_type_counts_as_content": [(
        '        if match.group("type") == b"blob":', '        if True:')],
    "T4_non_blob_content_not_skipped": [(
        '            answered[relative] = data[position:position + size]\n'
        '        position += size + 1',
        '            answered[relative] = data[position:position + size]\n'
        '            position += size + 1')],
    "T6_truncation_falls_back_to_break": [(
        '        if end < 0:\n', '        if end < 0:\n            break\n')],
    "T7_size_from_content_scan": [(
        '        size = int(match.group("size"))',
        '        size = data.find(b"\\0", position) - position')],
    "T18_no_prefill_of_unanswered_paths": [(
        '    found: dict[str, bytes | None] = {relative: None for relative in wanted}',
        '    found: dict[str, bytes | None] = {}')],
    "U1_leftover_bytes_ignored": [(
        '    if position != len(data):', '    if False:')],
    "U3_strict_utf8_encoding": [(
        '    asked = [spec.encode("utf-8", "surrogateescape") for spec in specs]',
        '    asked = [spec.encode("utf-8") for spec in specs]')],
    # --- 7.5.1: "못 물었다" 는 `None` 이 아니다 ---
    "V1_rc_failure_is_absence_again": [(
        '        said = process.stderr.decode("utf-8", "replace").strip().splitlines()\n',
        '        return found\n'
        '        said = process.stderr.decode("utf-8", "replace").strip().splitlines()\n')],
    "V2_rc_message_drops_what_git_said": [(
        '            f"(rc {process.returncode}" + (f": {said[0][:160]}" if said else "") + ")"',
        '            f"(rc {process.returncode})"')],
    "V3_any_missing_line_counts_as_absent": [(
        '        if header == request + b" missing" or _SUBMODULE_HEADER.fullmatch(header):',
        '        if header.endswith(b" missing") or _SUBMODULE_HEADER.fullmatch(header):')],
    "V4_unknown_header_is_absence": [(
        '        if match is None:\n', '        if match is None:\n            continue\n')],
    "V5_terminator_not_checked": [(
        '        if data[position + size] != 0:', '        if False:')],
    "V6_overshoot_not_checked": [(
        '        if position + size >= len(data):', '        if False:')],
    "V7_nul_checked_in_path_only": [(
        '    for spec in specs:\n        if "\\0" in spec:',
        '    for spec in wanted:\n        if "\\0" in spec:')],
    "V8_truncation_counts_blobs": [(
        '            raise RuntimeError(_TRUNCATED.format(count=index, total=len(wanted), ref=ref[:12]))\n'
        '        header',
        '            raise RuntimeError(_TRUNCATED.format(count=len(answered), total=len(wanted), ref=ref[:12]))\n'
        '        header')],
    "V9_fingerprint_not_rechecked": [(
        '            _raise_if_inputs_moved(root, analysis, inputs)\n            return candidate, ""',
        '            return candidate, ""')],
    "V11_resolve_measures_twice": [(
        '    computed, why = compute_landing(root, base, analysis, inputs)',
        '    computed, why = compute_landing(root, base, analysis)')],
    "V12_verdict_ignores_the_prefetch": [(
        '            blob = (prefetched[relative] if prefetched is not None and relative in prefetched\n'
        '                    else _committed_bytes(root, revision_ref, relative))',
        '            blob = _committed_bytes(root, revision_ref, relative)')],
    "V13_guard_seven_reads_per_source": [(
        '    at_base = _committed_many(root, base, sources)\n'
        '    at_candidate = _committed_many(root, candidate, sources)\n'
        '    if all(at_base[source] == at_candidate[source] for source in sources):',
        '    if all(_committed_bytes(root, base, source) == _committed_bytes(root, candidate, source)\n'
        '           for source in sources):')],
    "V14_a_handler_copies_its_own_list": [(
        '        required = changed_existing_functions(root, base, landing)\n'
        '    except GATE_FAULTS as exc:',
        '        required = changed_existing_functions(root, base, landing)\n'
        '    except (OSError, RuntimeError, ValueError) as exc:')],
    # --- 7.5.2: 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰) ---
    "W1_judged_list_read_separately": [(
        '    fingerprint, bundles, held = _evidence_fingerprint(root, analysis)\n',
        '    fingerprint, _, held = _evidence_fingerprint(root, analysis)\n'
        '    bundles = _pinning_bundles(root, analysis)\n')],
    "W2_fingerprint_drops_bytes": [(
        '    return Fingerprint(_head_commit(root), _digests(reads), pins), bundles, dict(reads)',
        '    return Fingerprint(_head_commit(root), (), pins), bundles, dict(reads)')],
    "W3_fingerprint_drops_head": [(
        '    return Fingerprint(_head_commit(root), _digests(reads), pins), bundles, dict(reads)',
        '    return Fingerprint("", _digests(reads), pins), bundles, dict(reads)')],
    "W4_fingerprint_drops_pins": [(
        '    return Fingerprint(_head_commit(root), _digests(reads), pins), bundles, dict(reads)',
        '    return Fingerprint(_head_commit(root), _digests(reads), ()), bundles, dict(reads)')],
    "W5_no_recheck_before_a_declared_refusal": [(
        '        _raise_if_inputs_moved(root, analysis, inputs)\n        # **복구 경로를 말한다**',
        '        # **복구 경로를 말한다**')],
    "W7_verdict_not_bound_to_the_landing": [(
        '    if judged is not None:\n        before, now', '    if False:\n        before, now')],
    "W8_target_reads_the_disk_again": [(
        '            target, root, index, require_calls, landing, prefetched, evidence,',
        '            target, root, index, require_calls, landing, prefetched,')],
    "W9_holding_reads_the_disk_again": [(
        '        if held is not None:\n            judged = held.get(ast_path)',
        '        if False:\n            judged = held.get(ast_path)')],
    "W10_call_table_reads_the_disk": [(
        '        (_bundle_text(path, evidence), _parsed(evidence.get(path.parent / "ast.json")))',
        '        (_bundle_text(path, evidence), _parsed((path.parent / "ast.json").read_bytes()\n'
        '         if (path.parent / "ast.json").exists() else None))')],
    "W11_bundle_text_reads_the_disk": [(
        '            if known is not None and path.name == "ast.json":',
        '            if False:')],
    "W12_prefetch_selection_not_contained": [(
        '            except ValueError:\n                sources = []',
        '            except ZeroDivisionError:\n                sources = []')],
    "W13_advice_fault_replaces_the_verdict": [(
        '        except GATE_FAULTS as exc:\n            facts["base_shaped_fault"] = str(exc)',
        '        except ZeroDivisionError as exc:\n            facts["base_shaped_fault"] = str(exc)')],
    "W14_submodule_is_a_fault": [(
        '        if header == request + b" missing" or _SUBMODULE_HEADER.fullmatch(header):',
        '        if header == request + b" missing":')],
    "W15_version_hint_on_every_failure": [(
        'if process.returncode == 129 else ""', 'if True else ""')],
    "W16_parser_fault_does_not_name_the_path": [(
        '                f"{index + 1} of {len(wanted)} ({relative}) at {ref[:12]}: {header[:80]!r}"',
        '                f"{index + 1} of {len(wanted)} at {ref[:12]}: {header[:80]!r}"')],
    "W17_shape_not_checked": [(
        '    if not isinstance(value["start"], dict) or not isinstance(value["end"], dict) \\\n',
        '    if False and not isinstance(value["start"], dict) or False and not isinstance(value["end"], dict) \\\n')],
    "W18_unmeasured_repairs_read_as_none": [(
        '    if repairs is None:\n        raise RuntimeError(',
        '    if repairs is None:\n        return []\n        raise RuntimeError(')],
    "W19_repairs_measured_as_empty": [(
        '    repairs = None if why else _self_repair_commits(root, analysis)',
        '    repairs = [] if why else _self_repair_commits(root, analysis)')],
    "W20_stale_context_kept": [(
        '    for stale in ("landing_evidence", "base_shaped_fault"):',
        '    for stale in ():')],
    "W21_base_shape_read_per_bundle": [(
        '        if at_base[source] is not None and hashlib.sha256(at_base[source]).hexdigest() == digest\n',
        '        if _committed_bytes(root, base, source) is not None\n'
        '        and hashlib.sha256(_committed_bytes(root, base, source)).hexdigest() == digest\n')],
    "W22_decode_without_newline_translation": [(
        '    return io.TextIOWrapper(io.BytesIO(raw), encoding="utf-8").read()',
        '    return raw.decode("utf-8")')],
    "W23_unreadable_evidence_is_absent": [(
        '            if raw is None:\n                errors.append(f"{target.name}: {name} could not be read")\n'
        '                continue',
        '            if raw is None:\n                errors.append(f"{target.name}: missing {name}")\n'
        '                continue')],
    # --- 배치가 실제로 배치인가 (7.5) ---
    "T8_pinning_at_fetches_one_by_one": [(
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles])\n'
        '    for _, source, digest in bundles:\n'
        '        blob = blobs[source]',
        '    for _, source, digest in bundles:\n'
        '        blob = _committed_bytes(root, candidate, source)')],
    "T9_unheld_fetches_one_by_one": [(
        '        committed = blobs[relative]\n'
        '        if committed is None and before:\n'
        '            committed = blobs[before]',
        '        committed = _committed_bytes(root, candidate, relative)\n'
        '        if committed is None and before:\n'
        '            committed = _committed_bytes(root, candidate, before)')],
    "T10_only_the_first_source_is_asked": [(
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles])',
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles][:1])')],
    "T11_pre_archive_path_not_batched": [(
        'root, candidate, [path for _, relative, before in watched\n'
        '                          for path in ((relative, before) if before else (relative,))],',
        'root, candidate, [relative for _, relative, _ in watched],')],
    "T12_walk_passes_no_bundles": [(
        '            root, base, candidate, bundles, floor, repairs, inputs.held,',
        '            root, base, candidate, [], floor, repairs, inputs.held,')],
    "T14_unheld_merged_guard_flipped": [(
        '        if committed is None and before:', '        if committed is None or before:')],
    "T15_anchor_is_not_the_root": [(
        '    anchor = root.resolve()', '    anchor = (root / "openspec").resolve()')],
    "T16_committed_bytes_is_a_second_spelling": [(
        '    return _committed_many(root, ref, [relative])[relative]',
        '    process = subprocess.run(\n'
        '        ["git", "show", f"{ref}:{relative}"],\n'
        '        cwd=root, capture_output=True, timeout=30, check=False,\n'
        '    )\n'
        '    return None if process.returncode else process.stdout')],
}


def setup() -> Path:
    if WORK.exists():
        shutil.rmtree(WORK)
    shutil.copytree(REPO / "tools" / "logic-map", WORK / "logic-map",
                    ignore=shutil.ignore_patterns("__pycache__"))
    (WORK / "sdd").mkdir(parents=True)
    shutil.copy(REPO / "tools" / "sdd" / "sdd_doctor.py", WORK / "sdd")
    return WORK / "logic-map" / "check_analysis.py"


def run(names: list[str]) -> tuple[int, str]:
    process = subprocess.run([sys.executable, "-m", "unittest", *names],
                             cwd=WORK / "logic-map", capture_output=True, text=True, timeout=5400)
    return process.returncode, process.stderr


def failing(output: str) -> list[str]:
    return sorted({line.split(" ")[1] for line in output.splitlines()
                   if line.startswith(("FAIL: ", "ERROR: "))})


def reached(target: Path, edits) -> bool:
    """변이가 **돌았는지**를 잰다 — 바뀐 줄에 도달하지 못하면 SURVIVED 는 음성이 아니라 침묵이다.

    문자열이 바뀌었는지만 보는 하네스는 **눈먼 계측기와 진짜 음성을 같게 기록한다**
    (2026-09-18 독립 리뷰 P2; T6 이 정확히 그 결과였다 — `end < 0` 갈래는 182개 시험에서
    0회 도달인데 "동등 변이"로 적혔다, [[mutation-must-reach-the-thing-under-test]]).
    바꾼 줄마다 표식을 심고 스위트를 돌려 표식이 찍히는지 본다.
    """
    text = target.read_text(encoding="utf-8")
    marked = text
    for _, new_line in edits if edits != "MOVE_FIRST" else []:
        head = new_line.splitlines()[0]
        if head.strip().startswith(("#", '"')) or not head.strip():
            continue
        indent = head[: len(head) - len(head.lstrip())]
        marked = marked.replace(
            head, f'{indent}import sys as _s; print("REACHED", file=_s.stderr)\n{head}', 1
        )
    if marked == text:
        return False
    target.write_text(marked, encoding="utf-8")
    _, output = run(SUITE)
    target.write_text(text, encoding="utf-8")
    return "REACHED" in output


if __name__ == "__main__":
    target = setup()
    pristine = target.read_text(encoding="utf-8")
    assert "_committed_many" in pristine, "대상이 사본에 없다"
    code, output = run(SUITE)
    if code != 0:
        print("STOP — 무변이 대조군이 빨갛다\n", output[-3000:])
        raise SystemExit(1)
    print("control GREEN", output.strip().splitlines()[-1])
    survived = []
    # 구간 실행: `75_mut.py 0:8` — 이 환경에서는 백그라운드로 넘어간 프로세스가 살아남지 못해서
    # 한 번에 다 못 돈다. 구간마다 무변이 대조군은 **그대로 먼저** 돈다(위).
    window = next((a for a in sys.argv[1:] if ":" in a), ":")
    lo, _, hi = window.partition(":")
    chosen = list(MUTATIONS.items())[int(lo) if lo else None:int(hi) if hi else None]
    print(f"구간 {window} — 변이 {len(chosen)}/{len(MUTATIONS)}")
    for name, edits in chosen:
        text = pristine
        for old, new in edits:
            assert text.count(old) == 1, f"{name}: 앵커 {old[:60]!r} 가 {text.count(old)}회"
            text = text.replace(old, new)
        target.write_text(text, encoding="utf-8")
        code, output = run(SUITE)
        names = failing(output)
        verdict = "CAUGHT" if code else "SURVIVED"
        if not code:
            survived.append(name)
        touched = "" if code else ("  · 도달함" if reached(target, edits) else "  · **안 닿음**")
        print(f"{name:36s} {verdict:9s}{touched} {len(names):2d} "
              f"{', '.join(n.split('.')[-1] for n in names[:3])}"
              + (f" 외 {len(names) - 3}" if len(names) > 3 else ""))
        target.write_text(pristine, encoding="utf-8")
    print(f"\nSURVIVED {len(survived)}/{len(chosen)}" + (f": {survived}" if survived else ""))
