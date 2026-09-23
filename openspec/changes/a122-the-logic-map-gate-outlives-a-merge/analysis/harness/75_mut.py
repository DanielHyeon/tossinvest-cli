#!/usr/bin/env python3
"""task 7.5 변이 — 새 배치 fetch 의 갈래마다 **실제로** 빨개지는 시험이 있는가.

사본 대상 + **무변이 대조군**이 먼저다 ([[mutation-revert-needs-the-right-baseline]] ·
[[mutation-must-reach-the-thing-under-test]]). 대조군이 초록이 아니면 멈춘다.
"""
import os
import re
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
# 사본은 **프로세스별**이다 (task 7.5.2.3): 두 판이 한 사본을 쓰면 한쪽의 변이가 다른 쪽의 기준이 되고,
# 그러면 CAUGHT/SURVIVED 가 뒤섞인다 — 2026-09-20 에 배경 판과 전경 창이 실제로 그렇게 됐다.
WORK = SP / f"75_mut_work.{os.getpid()}"
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
    "V1_rc_failure_is_absence_again": [
        ('        said = _first_line(process.stderr, "")',
         '        return found\n        said = _first_line(process.stderr, "")')],
    "V2_rc_message_drops_what_git_said": [
        ('            f"(rc {process.returncode}" + (f": {said}" if said else "") + ")"',
         '            f"(rc {process.returncode})"')],
    "V3_any_missing_line_counts_as_absent": [(
        '        if header == request + b" missing" or _SUBMODULE_HEADER.fullmatch(header):',
        '        if header.endswith(b" missing") or _SUBMODULE_HEADER.fullmatch(header):')],
    # 앵커에 앞 줄을 붙여 자리를 **특정**한다 — 7.5.2.4 가 diff 파서에 같은 철자의 줄을 하나 더
    # 만들었고, 하네스는 그때 `count != 1` 로 멈췄다(조용히 아무 자리나 고르지 않는다).
    "V4_unknown_header_is_absence": [(
        '        match = _OBJECT_HEADER.fullmatch(header)\n        if match is None:\n',
        '        match = _OBJECT_HEADER.fullmatch(header)\n        if match is None:\n            continue\n')],
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
    "V9_fingerprint_not_rechecked": [
        ('            _raise_if_inputs_moved(root, inputs)\n            return candidate, ""',
         '            return candidate, ""')],
    "V11_resolve_measures_twice": [
        ('    computed, why = compute_landing(root, base, inputs)',
         '    computed, why = compute_landing(root, base, _measure_landing_inputs(root, head, evidence))')],
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
    "W1_judged_list_read_separately": [
        ('    bundles = _select_pinning(root, evidence)\n    floor, why = _walk_floor(root, bundles, head)',
         '    bundles = _select_pinning(root, _read_evidence(evidence.directory))\n    floor, why = _walk_floor(root, bundles, head)')],
    "W5_no_recheck_before_a_declared_refusal": [
        ('        _raise_if_inputs_moved(root, inputs)\n        # **복구 경로를 말한다**',
         '        # **복구 경로를 말한다**')],
    "W8_target_reads_the_disk_again": [
        ('            target, root, index, require_calls, landing, prefetched, held=evidence.held,',
         '            target, root, index, require_calls, landing, prefetched,\n            held=_read_evidence(evidence.directory).held,')],
    "W9_holding_reads_the_disk_again": [
        ('        judged = held.get(ast_path)             # 판정과 **같은 읽기** — 심링크면 따라간 바이트다',
         '        judged = ast_path.read_bytes() if ast_path.is_file() else None')],
    "W10_call_table_reads_the_disk": [
        ('        (text, _parsed(evidence.held.get(target / "ast.json"))) for target, text in bundle_texts.items()',
         '        (text, _parsed((target / "ast.json").read_bytes() if (target / "ast.json").is_file() else None))\n        for target, text in bundle_texts.items()')],
    "W11_bundle_text_keeps_bytes_only_while_listed": [
        ('    if ast_raw is not None:\n        paths["ast.json"] = target / "ast.json"',
         '    if ast_raw is not None and (target / "ast.json").exists():\n        paths["ast.json"] = target / "ast.json"')],
    "W12_prefetch_selection_not_contained": [
        ('        except ValueError:\n            sources = []',
         '        except ZeroDivisionError:\n            sources = []')],
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
    "W17_object_shape_not_checked": [
        ('    if any(value.get(key) and not isinstance(value[key], dict) for key in _AST_OBJECTS) \\\n',
         '    if False and any(value.get(key) and not isinstance(value[key], dict) for key in _AST_OBJECTS) \\\n')],
    "W18_unmeasured_repairs_read_as_none": [(
        '    if repairs is None:\n        raise RuntimeError(',
        '    if repairs is None:\n        return []\n        raise RuntimeError(')],
    "W19_repairs_measured_as_empty": [
        ('    repairs = None if why else _self_repair_commits(root, evidence.directory, head)',
         '    repairs = [] if why else _self_repair_commits(root, evidence.directory, head)')],
    "W20_stale_context_half_kept": [
        ('    facts.clear()',
         '    facts.pop("head", None)')],
    "W21_base_shape_read_per_bundle": [(
        '        if at_base[source] is not None and hashlib.sha256(at_base[source]).hexdigest() == digest\n',
        '        if _committed_bytes(root, base, source) is not None\n'
        '        and hashlib.sha256(_committed_bytes(root, base, source)).hexdigest() == digest\n')],
    "W22_decode_without_newline_translation": [(
        '    return io.TextIOWrapper(io.BytesIO(raw), encoding="utf-8").read()',
        '    return raw.decode("utf-8")')],
    "W23_unreadable_evidence_is_absent": [
        ('            if held[path] is None:\n                errors.append(f"{target.name}: {name} could not be read")',
         '            if held[path] is None:\n                errors.append(f"{target.name}: missing {name}")')],
    # --- 7.5.2.1: 명령 하나는 역사 하나와 증거 읽기 하나 위에서 판정한다 ---
    "Y1_record_read_at_symbolic_head": [
        ('    raw = _committed_bytes(root, head, relative)',
         '    raw = _committed_bytes(root, "HEAD", relative)')],
    "Y2_ancestry_at_symbolic_head": [
        ('    if not _is_ancestor(root, candidate, head):',
         '    if not _is_ancestor(root, candidate, "HEAD"):')],
    "Y3_floor_at_symbolic_head": [
        ('"--format=%H",\n         head, "--", *paths],',
         '"--format=%H",\n         "HEAD", "--", *paths],')],
    "Y4_repair_signal_at_symbolic_head": [
        ('"--no-merges", "--format=%H", head, "--", *paths],',
         '"--no-merges", "--format=%H", "HEAD", "--", *paths],')],
    "Y5_walk_at_symbolic_head": [
        ('f"{start}..{inputs.head}"',
         'f"{start}..HEAD"')],
    "Y6_repairs_after_at_symbolic_head": [
        ('["git", "rev-list", f"{candidate}..{head}"]',
         '["git", "rev-list", f"{candidate}..HEAD"]')],
    "Y7_cleanliness_at_symbolic_head": [
        ('["git", "diff", "--quiet", head]',
         '["git", "diff", "--quiet", "HEAD"]')],
    "Y8_window_count_at_symbolic_head": [
        ('f"{base}..{head}"',
         'f"{base}..HEAD"')],
    "Y9_ancestry_fault_is_a_no": [
        ('    if process.returncode in (0, 1):\n        return process.returncode == 0',
         '    if True:\n        return process.returncode == 0')],
    "Y10_floor_fault_is_empty": [
        ('    if process.returncode:\n        raise RuntimeError(\n            "cannot find the commit that put',
         '    if process.returncode:\n        return ""\n        raise RuntimeError(\n            "cannot find the commit that put')],
    "Y11_walk_fault_is_a_reason": [
        ('    if process.returncode:\n        raise RuntimeError(\n            f"cannot walk the history after {start[:12]}: "',
         '    if process.returncode:\n        return "", "cannot walk"\n        raise RuntimeError(\n            f"cannot walk the history after {start[:12]}: "')],
    "Y12_cleanliness_fault_is_dirty": [
        ('    if dirty.returncode not in (0, 1):',
         '    if False:')],
    # Y13 · Y14 는 `_verdict` 에 **없는 이름** `analysis` 를 넣어 `NameError` 로 빨개졌다 (task 7.5.2.4
    # 재리뷰 시험품질). 함수가 갈릴 때 낡은 것이고, 그동안 이 둘은 "이른 반환이 디스크에 묻는다" ·
    # "targets 를 다시 나열한다" 를 **증명하지 않았다**. 같은 뜻을 오늘의 범위로 다시 쓴다 —
    # `evidence.directory` 가 바로 그 디렉터리다.
    "Y13_early_return_asks_the_disk": [
        ('    if not evidence.present:',
         '    if not evidence.directory.exists():')],
    "Y14_targets_listed_again": [
        ('    for target in evidence.targets:\n        target_errors, binding',
         '    for target in sorted(path for path in evidence.directory.iterdir() if path.is_dir()):\n        target_errors, binding')],
    "Y15_evidence_listed_by_glob": [
        ('    targets = tuple(analysis / name for name, is_dir in entries if is_dir)',
         '    targets = tuple(sorted(path.parent for path in analysis.glob("*/ast.json")))')],
    "Y16_absent_read_as_unreadable": [
        ('        except FileNotFoundError:\n            pass\n        except OSError:',
         '        except FileNotFoundError:\n            held[ast_path] = None\n        except OSError:')],
    "Y17_listing_failure_is_silent": [
        ('            named = exc.filename if isinstance(exc.filename, str) else ""',
         '            continue\n            named = exc.filename if isinstance(exc.filename, str) else ""')],
    "Y18_verdict_reads_the_evidence_again": [
        ('    facts["landing"] = landing\n    facts["required_count"] = len(required)',
         '    evidence = _read_evidence(analysis)\n    facts["landing"] = landing\n    facts["required_count"] = len(required)')],
    "Y19_record_command_reads_twice": [
        ('        landing, why = compute_landing(root, base, _measure_landing_inputs(root, head, evidence))',
         '        landing, why = compute_landing(root, base, _measure_landing_inputs(\n            root, head, _read_evidence(change_dir / "analysis" / "function-logic")))')],
    "Y20_recheck_skips_the_bytes": [
        ('    if now != inputs.evidence:\n        raise RuntimeError(INPUTS_MOVED)',
         '    if False:\n        raise RuntimeError(INPUTS_MOVED)')],
    "Y21_recheck_compares_only_the_bytes_held": [
        ('    if now != inputs.evidence:\n        raise RuntimeError(INPUTS_MOVED)',
         '    if now.held != inputs.evidence.held:\n        raise RuntimeError(INPUTS_MOVED)')],
    "Y22_recheck_skips_the_pins": [
        ('    if bundles != inputs.bundles:\n        raise RuntimeError(INPUTS_MOVED)',
         '    if False:\n        raise RuntimeError(INPUTS_MOVED)')],
    "Y23_selection_fault_is_not_a_move": [
        ('    except ValueError as exc:\n        raise RuntimeError(INPUTS_MOVED) from exc',
         '    except ZeroDivisionError as exc:\n        raise RuntimeError(INPUTS_MOVED) from exc')],
    "Y24_list_shape_not_checked": [
        ('            or any(value.get(key) and not isinstance(value[key], list) for key in _AST_LISTS):',
         '            or False:')],
    "Y25_parsed_value_skips_the_shape": [
        ('    return {} if raw is None else _parse_ast(raw)[0]',
         '    return {} if raw is None else json.loads(_decoded(raw))')],
    "Y26_non_dict_is_invalid_not_placeholder": [
        ('    if not isinstance(value, dict):\n        return {}, "placeholder"',
         '    if not isinstance(value, dict):\n        return {}, "invalid"')],
    "Y27_output_stays_strict": [
        ('        reconfigure(errors="backslashreplace")',
         '        reconfigure(errors="strict")')],
    "Y28_resolve_raises_on_a_loop": [
        ('    resolved = Path(os.path.realpath(path))',
         '    resolved = path.resolve()')],
    # --- 7.5.2.2: 판정은 내놓는 순간에도 거기 있는 것의 판정이다 · 번들 파일은 정규 파일만 ---
    "Z1_non_regular_files_are_read": [
        ('        if not stat.S_ISREG(mode):\n            raise NotRegularFile(errno.EINVAL, NOT_REGULAR, str(path))',
         '        if False:\n            raise NotRegularFile(errno.EINVAL, NOT_REGULAR, str(path))')],
    "Z2_unreadable_bundle_file_is_skipped": [
        ('            except FileNotFoundError:\n                continue                        # 끊긴 링크',
         '            except OSError:\n                continue                        # 끊긴 링크')],
    "Z3_evidence_read_bypasses_the_type_check": [
        ('            held[ast_path] = _read_regular(ast_path)',
         '            held[ast_path] = ast_path.read_bytes()')],
    "Z4_prose_read_bypasses_the_type_check": [
        ('                raw = _read_regular(path)',
         '                raw = path.read_bytes()')],
    "Z5_no_recheck_at_the_exit": [
        ('        moved = _judged_state_moved(root, str(facts["head"]), book)\n        return [moved] if moved else verdict',
         '        return verdict')],
    "Z6_exit_recheck_ignores_head": [
        ('    if now != head:\n        return JUDGED_STATE_MOVED',
         '    if False:\n        return JUDGED_STATE_MOVED')],
    "Z7_exit_recheck_ignores_evidence": [
        ('    changed = _reads_moved(root, book)',
         '    changed = ""')],
    "Z8_record_writes_without_recheck": [
        ('            moved = _recording_moved(change, change_dir, root, head, book) if landing else ""',
         '            moved = ""')],
    "Z9_window_line_without_head": [
        ('function(s) — judged at HEAD {head[:12]}"',
         'function(s)"')],
    "Z10_context_not_cleared": [
        ('    facts.clear()',
         '    pass')],
    "Z11_recursion_escapes_the_parse": [
        ('    except (ValueError, RecursionError):',
         '    except ValueError:')],
    "Z12_empty_computation_says_nothing": [
        ('        named = computed[:12] if computed else f"none — {why}"',
         '        named = computed[:12]')],
    "Z13_file_failure_named_as_the_directory": [
        ('            where = Path(named).name if named else target.name',
         '            where = target.name')],
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
    "T12_walk_passes_no_bundles": [
        ('        refusal, names = _landing_refusal(root, base, candidate, inputs)',
         '        refusal, names = _landing_refusal(root, base, candidate, inputs._replace(bundles=[]))')],
    "T14_unheld_merged_guard_flipped": [(
        '        if committed is None and before:', '        if committed is None or before:')],
    "T15_anchor_is_not_the_root": [
        ('    anchor = Path(os.path.realpath(root))',
         '    anchor = Path(os.path.realpath(root / "openspec"))')],
    "T16_committed_bytes_is_a_second_spelling": [(
        '    return _committed_many(root, ref, [relative])[relative]',
        '    process = subprocess.run(\n'
        '        ["git", "show", f"{ref}:{relative}"],\n'
        '        cwd=root, capture_output=True, timeout=30, check=False,\n'
        '    )\n'
        '    return None if process.returncode else process.stdout')],
    # --- 7.5.2.2 BTM 을 쓰다가 찾은 빈 칸 둘 — 갈래는 있는데 그것을 지우는 변이가 없었다 (구간 `89:` 로 따로 돈다) ---
    # 7.5.2.3: `raw is None` 갈래가 없어졌다(정규 파일이 아니면 이름 댄 예외). 같은 구멍 — 못 푸는
    # 바이트를 조용히 넘기기 — 를 되살리는 변이로 다시 겨눈다.
    "Z14_bundle_text_skips_what_it_cannot_decode": [
        ('            raise NotUtf8Text(errno.EILSEQ, NOT_UTF8, str(paths[name])) from exc',
         '            continue')],
    "Z19_the_descriptor_is_left_open": [
        ('    finally:\n        os.close(descriptor)',
         '    finally:\n        pass')],
    "Z16_a_folder_is_skipped_like_a_fifo": [
        ('        if stat.S_ISDIR(mode):\n            raise IsADirectoryError',
         '        if False:\n            raise IsADirectoryError')],
    "Z17_the_name_is_used_without_a_type_check": [
        ('            named = exc.filename if isinstance(exc.filename, str) else ""',
         '            named = exc.filename')],
    "Z18_the_folder_regression_returns": [
        ('    descriptor = os.open(path, os.O_RDONLY | os.O_NONBLOCK)\n'
         '    try:\n'
         '        mode = os.fstat(descriptor).st_mode\n'
         '        if stat.S_ISDIR(mode):\n'
         '            raise IsADirectoryError(errno.EISDIR, os.strerror(errno.EISDIR), str(path))',
         '    descriptor = os.open(path, os.O_RDONLY | os.O_NONBLOCK)\n'
         '    try:\n'
         '        handle_first = open(descriptor, "rb", closefd=False)\n'
         '        mode = os.fstat(descriptor).st_mode\n'
         '        if stat.S_ISDIR(mode):\n'
         '            raise IsADirectoryError(errno.EISDIR, os.strerror(errno.EISDIR))')],
    "Z15_unreadable_prose_is_called_missing": [
        ('            except OSError:\n                errors.append(f"{target.name}: {name} could not be read")',
         '            except OSError:\n                errors.append(f"{target.name}: missing {name}")')],
    # --- 7.5.2.3: 재확인의 입력 집합은 판정의 입력 집합이다 (원장 · 깔때기 · 쓰기 직전 거절) ---
    "AA1_ledger_forgets_failed_reads": [
        ('    _remember("file", str(path), outcome)',
         '    _remember("file", str(path), outcome) if not isinstance(value, OSError) else None')],
    "AA2_ledger_forgets_listings": [
        ('    _remember("dir", str(path), outcome)', '    pass')],
    "AA3_ledger_forgets_tree_walks": [
        ('    _remember("glob", f"{root}\\n{pattern}", outcome)', '    pass')],
    "AA4_ledger_forgets_what_it_chose": [
        ('    outcome, kind = _kind_outcome(path)\n    _remember("kind", str(path), outcome)\n    return kind',
         '    outcome, kind = _kind_outcome(path)\n    return kind')],
    "AA5_recheck_does_not_ask_head_after": [
        ('    return _head_moved(root, head)\n\n\ndef _raise_if_inputs_moved',
         '    return ""\n\n\ndef _raise_if_inputs_moved')],
    "AA6_recheck_does_not_ask_head_first": [
        ('    moved = _head_moved(root, head)\n    if moved:\n        return moved\n    changed = _reads_moved(root, book)',
         '    changed = _reads_moved(root, book)')],
    "AA7_record_does_not_reask_the_refusals": [
        ('    refusal, _ = _recording_refusal(\n'
         '        change, change_dir, root, head, _read_evidence(change_dir / "analysis" / "function-logic"))\n'
         '    return refusal or _judged_state_moved(root, head, book)',
         '    return _judged_state_moved(root, head, book)')],
    "AA8_record_asks_the_refusals_before_history": [
        ('    moved = _head_moved(root, head)\n    if moved:\n        return moved\n'
         '    # 거절은 **쓰는 순간의** 디스크에 대한 질문이므로 증거를 다시 읽어서 묻는다.',
         '    # 거절은 **쓰는 순간의** 디스크에 대한 질문이므로 증거를 다시 읽어서 묻는다.')],
    "AA9_no_size_cap": [
        ('    if len(raw) > READ_CAP:', '    if False:')],
    "AA10_reads_past_the_cap": [
        ('            raw = handle.read(READ_CAP + 1)', '            raw = handle.read()')],
    "AA11_undecodable_prose_is_not_named": [
        ('            try:\n                texts[name] = _decoded(raw)\n            except UnicodeDecodeError:',
         '            texts[name] = _decoded(raw)\n            try:\n                pass\n            except UnicodeDecodeError:')],
    "AA12_unlistable_bundle_is_silent": [
        ('        except OSError as exc:\n            unlistable[target] = _why(exc)',
         '        except OSError as exc:\n            pass')],
    "AA13_unreadable_evidence_dir_is_an_exemption": [
        ('    except (FileNotFoundError, NotADirectoryError):\n        return Evidence(analysis, False, (), {}, {}, {})',
         '    except OSError:\n        return Evidence(analysis, False, (), {}, {}, {})')],
    "AA14_unreadable_review_is_an_empty_marker": [
        ('    except (OSError, UnicodeDecodeError) as exc:\n'
         '        return [UNREADABLE.format(what="review.md", why=_why(exc))], False',
         '    except (OSError, UnicodeDecodeError):\n        review_text = ""')],
    "AA15_unreadable_base_is_called_missing": [
        ('    except FileNotFoundError as exc:\n        raise ValueError(\n            "missing base-commit.txt; run "',
         '    except OSError as exc:\n        raise ValueError(\n            "missing base-commit.txt; run "')],
    "AA16_borrowed_change_ignores_local_bundles": [
        ('        if any(_listed(bundle) for bundle in local):', '        if False:')],
    "AA17_ledger_ignores_a_split_read": [
        ('    if previous is not None and previous != outcome and not book.diverged:',
         '    if False:')],
    "AA19_listing_fingerprint_drops_the_kind": [
        ('    joined = "\\n".join(f"{name}\\t{\'d\' if is_dir else \'f\'}" for name, is_dir in entries)',
         '    joined = "\\n".join(name for name, is_dir in entries)')],
    # --- 통합 diff 문법의 상태 (7.5.2.4) ---
    # 본문 줄이 파일 이름을 정하면 그 파일의 요구가 사라지거나 편집 전 리비전으로 내려앉는다.
    "AB1_body_lines_name_the_file_again": [
        ('        elif in_body:\n            # 본문이다. 여기서 `--- `·`+++ ` 는 소스 줄이지 파일 이름이 아니다.\n            hunk(line)',
         '        elif False:\n            # 본문이다. 여기서 `--- `·`+++ ` 는 소스 줄이지 파일 이름이 아니다.\n            hunk(line)')],
    "AB2_the_body_never_opens": [
        ('        elif hunk(line):\n            # 파일의 **첫** 훅이다. 본문을 냈다는 표시는 여기 한 번이면 된다 — 뒤의 훅은 같은\n            # 구역에 속하므로 더 말할 것이 없다 (task 7.5.23).\n            in_body = True', '        elif hunk(line):\n            # 파일의 **첫** 훅이다. 본문을 냈다는 표시는 여기 한 번이면 된다 — 뒤의 훅은 같은\n            # 구역에 속하므로 더 말할 것이 없다 (task 7.5.23).\n            in_body = False')],
    "AB3_a_new_file_does_not_close_the_body": [
        ('            new_source = ""\n            in_body = False', '            new_source = ""')],
    "AB4_only_the_first_hunk_of_a_file_counts": [
        ('            # 본문이다. 여기서 `--- `·`+++ ` 는 소스 줄이지 파일 이름이 아니다.\n            hunk(line)',
         '            # 본문이다. 여기서 `--- `·`+++ ` 는 소스 줄이지 파일 이름이 아니다.\n            pass')],
    "AB5_dev_null_is_an_ordinary_name": [
        ('            old_source = "" if value == "/dev/null" else value.removeprefix("a/")',
         '            old_source = value.removeprefix("a/")')],
    # --- task 7.5.22: 본문 없는 `*.go` 는 조용히 비지 않는다 ---
    "AC1_a_suppressed_body_is_ordinary": [
        ('        if added == b"-" and deleted == b"-":',
         '        if False:')],
    "AC2_mode_only_is_refused_too": [
        ('        if added == b"-" and deleted == b"-":',
         '        if added in (b"-", b"0") and deleted in (b"-", b"0"):')],
    "AC3_a_rename_keeps_only_the_new_name": [
        ('        pair = [item for item in chunks[index + 1:index + 3] if item]',
         '        pair = [item for item in chunks[index + 2:index + 3] if item] * 2')],
    "AC4_the_new_guard_stands_first": [
        ('''    records = _numstat_records(raw_output)
    for _, _, paths in records:''',
         '''    records = _numstat_records(raw_output)
    for added, deleted, paths in records:
        if added == b"-" and deleted == b"-":
            raise RuntimeError(
                "modified Go file has no textual diff (binary or -diff attribute): "
                + paths[-1].decode("utf-8", "replace")
            )
    for _, _, paths in records:''')],
    # --- task 7.5.23: 가드와 판정이 같은 diff 를 읽는다 ---
    "AD1_the_judged_diff_trusts_textconv": [
        ('            "--no-textconv",\n', '')],
    "AD2_the_judged_diff_trusts_an_external_diff": [
        ('            "--no-ext-diff",\n            "--no-textconv",\n',
         '            "--no-textconv",\n')],
    "AD3_a_vanished_body_is_no_change": [
        ('        if not had_body:', '        if False:')],
    "AD4_a_mode_only_file_must_have_a_body": [
        ('        if added in (b"-", b"0") and deleted in (b"-", b"0"):',
         '        if added == b"-" and deleted == b"-":')],
    # --- task 7.5.24: 짝은 이름이 아니라 순서로 ---
    "AE1_the_skip_takes_either_zero": [
        ('        if added in (b"-", b"0") and deleted in (b"-", b"0"):',
         '        if added in (b"-", b"0") or deleted in (b"-", b"0"):')],
    "AE2_the_two_views_need_not_agree_in_size": [
        ('    if len(bodied) != len(records):', '    if False:')],
    "AE3_a_section_is_marked_before_its_hunk": [
        ('            bodied.append(False)', '            bodied.append(True)')],
    "AE4_only_the_last_section_can_have_a_body": [
        ('            if bodied:\n                bodied[-1] = True',
         '            pass')],
    "AD6_the_guard_ignores_renames": [
        ('"--no-ext-diff", "--no-textconv", "--find-renames",',
         '"--no-ext-diff", "--no-textconv",')],
    "AC5_the_hunk_reads_the_base_side_twice": [
        ('''                int(match.group(3)),
                int(match.group(4) or 1),''',
         '''                int(match.group(1)),
                int(match.group(2) or 1),''')],
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


def ran_count(output: str) -> int:
    """`Ran N tests` 의 N — 판마다 같은 수를 돌았는지 보는 싼 환경 대조."""
    match = re.search(r"^Ran (\d+) tests", output, re.MULTILINE)
    return int(match.group(1)) if match else -1


def failing(output: str) -> list[str]:
    return sorted({line.split(" ")[1] for line in output.splitlines()
                   if line.startswith(("FAIL: ", "ERROR: "))})


def reached(target: Path, edits, pristine: str, control_ran: int) -> str:
    """변이가 **돌았는지**를 잰다. `"YES"` · `"no"` · `"?"`(못 쟀다) 셋 중 하나를 돌려준다.

    문자열이 바뀌었는지만 보는 하네스는 **눈먼 계측기와 진짜 음성을 같게 기록한다**
    (2026-09-18 독립 리뷰 P2; T6 이 정확히 그 결과였다 — `end < 0` 갈래는 182개 시험에서
    0회 도달인데 "동등 변이"로 적혔다, [[mutation-must-reach-the-thing-under-test]]).
    바꾼 줄마다 표식을 심고 스위트를 돌려 표식이 찍히는지 본다.

    **계측기가 거짓말하던 두 자리를 막는다 (task 7.5.22, 7.5.2.4 재리뷰).**

    1. 표식을 **문장이 아닌 자리** 앞에 심으면(여러 줄 호출의 중간 줄 · `elif`/`else`/`except` 머리)
       파이썬 문법이 깨진다. 그러면 인터프리터가 `SyntaxError` 트레이스백을 찍는데 **그 트레이스백이
       문제의 소스 줄을 그대로 인쇄하고**, 그 줄에 `REACHED` 가 들어 있다. 옛 판본은 그것을 읽고
       시험을 **0개** 돌린 판에 "도달함" 이라고 답했다(118 중 **30**). → 쓰기 전에 `compile()` 한다.
    2. 표식이 문법을 안 깨도 스위트가 **수집 단계에서** 죽으면 같은 일이 난다. → `Ran N tests` 가
       대조군과 같은 N 을 낼 때만 표식을 읽는다. 안 맞으면 "도달함" 이 아니라 `"?"` 다.

    3. **표식 head 가 파일에서 유일하지 않으면 `str.replace` 가 첫 자리에 심는다** (task 7.5.23,
       시험품질 재리뷰). 그러면 계측기는 **한 번도 안 본 줄**에 대해 자신 있게 `YES`/`no` 를 답한다.
       `compile()` 은 이것을 못 본다 — 엉뚱한 자리에 심어도 문법은 멀쩡하다. 실측: 표식 head 20개가
       유일하지 않고, 그중 **다섯**이 다른 문장에(넷은 **다른 함수**에) 앉았다. → 유일할 때만 심는다.

    4. **일부만 심겼으면 그렇게 말한다.** 변이가 줄 셋을 바꾸는데 하나만 심겼으면 `no` 는
       "셋 다 안 닿았다" 가 아니라 "하나가 안 닿았다" 이다. 그 판은 `"?"` 로 낸다.

    넷을 가르는 것이 요점이다 — **"안 닿음" 과 "못 쟀다" 는 다른 말이다.** 옛 판본은 둘 다 `False` 였다.
    `control_ran` 은 기본값이 없다 — 0 을 넘기면 위 2번이 조용히 꺼지므로 호출자가 반드시 준다.
    """
    text = pristine
    marked = text
    eligible = planted = 0
    for old_line, _ in edits if edits != "MOVE_FIRST" else []:
        head = old_line.splitlines()[0]
        if head.strip().startswith(("#", '"')) or not head.strip():
            continue
        eligible += 1
        if pristine.count(head) != 1:
            # 유일하지 않다. `str.replace(…, 1)` 은 **첫** 자리에 심으므로 엉뚱한 줄을 잴 수 있다.
            continue
        indent = head[: len(head) - len(head.lstrip())]
        candidate = marked.replace(
            head, f'{indent}import sys as _s; print("REACHED", file=_s.stderr)\n{head}', 1
        )
        if candidate == marked:
            continue
        try:
            compile(candidate, str(target), "exec")
        except SyntaxError:
            # 이 자리는 문장 앞이 아니다. 심지 **않는다** — 심으면 트레이스백이 표식을 인쇄해
            # 시험 0개 돈 판이 "도달함" 이 된다.
            continue
        marked = candidate
        planted += 1
    if not planted or planted != eligible:
        # 심을 자리가 없거나(주석·문자열만 바꾸는 변이) · 하나도 못 심었거나 · 일부만 심었다.
        # 셋 다 답이 아니다. `eligible == planted == 0` 을 따로 막지 않으면 표식 없는 판을
        # 돌려 놓고 "안 닿음" 이라고 답한다 — 옛 판본의 거짓말이 그 모양으로 되살아난다.
        return "?"
    target.write_text(marked, encoding="utf-8")
    try:
        _, output = run(SUITE)
    finally:
        # 스위트가 멎어도 표식 붙은 사본을 남기지 않는다 ([[mutation-revert-needs-the-right-baseline]]).
        target.write_text(pristine, encoding="utf-8")
    if ran_count(output) != control_ran:
        # 스위트가 대조군과 같은 수를 안 돌았다 — 표식을 읽을 자격이 없다.
        return "?"
    return "YES" if "REACHED" in output else "no"


if __name__ == "__main__":
    target = setup()
    pristine = target.read_text(encoding="utf-8")
    assert "_committed_many" in pristine, "대상이 사본에 없다"
    # 사본이 **원본과 같은지** 단언한다 (task 7.5.2.3). 창을 나눠 돌리다 한 판이 중간에 죽으면 사본에
    # 변이가 남고, 그 뒤의 모든 창이 **변이된 기준** 위에서 돈다 — 무변이 대조군이 빨개져야 알 수 있는데
    # 그 빨감의 이유를 찾는 데 한 시간이 든다. 계측기부터 못 박는다
    # ([[mutation-revert-needs-the-right-baseline]] · [[mutation-must-reach-the-thing-under-test]]).
    origin = (REPO / "tools" / "logic-map" / "check_analysis.py").read_text(encoding="utf-8")
    # 이 단언은 **구조상 참이다** — `setup()` 이 방금 무조건 rmtree+copytree 했다 (task 7.5.2.4 정정).
    # 두 판이 서로의 변이를 기준으로 삼는 것을 실제로 막는 것은 위의 **pid 별 `WORK` 이름** 하나다.
    # 싼 연기 감지기로 남겨 두되, 이것을 수리라고 적지 않는다.
    assert pristine == origin, "사본이 원본과 다르다 — 앞선 판이 변이를 남겼다"
    code, output = run(SUITE)
    if code != 0:
        print("STOP — 무변이 대조군이 빨갛다\n", output[-3000:])
        raise SystemExit(1)
    control_line = output.strip().splitlines()[-1]
    control_ran = ran_count(output)
    print(f"control GREEN {control_line} (Ran {control_ran})")
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
        target.write_text(pristine, encoding="utf-8")
        # 도달 계측은 **원복 뒤에** 부른다 — 표식은 원본 본문에 심는 것이지 변이된 본문이 아니다.
        # 계측기는 셋을 가른다 (task 7.5.22): 도달함 · 안 닿음 · **못 쟀다**. 옛 판본은 뒤의 둘을
        # 같은 `False` 로 뭉갰고, 표식이 문법을 깨면 트레이스백이 표식을 인쇄해 시험 0개 돈 판을
        # "도달함" 으로 만들었다.
        touched = "" if code else {
            "YES": "  · 도달함", "no": "  · **안 닿음**", "?": "  · **못 쟀다**",
        }[reached(target, edits, pristine, control_ran)]
        # 판이 대조군과 **다른 수의 시험**을 돌았으면 그 CAUGHT 는 변이의 증거가 아니라 환경의 증거다
        # (task 7.5.2.4, 재리뷰 시험품질: `/` 가 0 인 창에서 다섯 변이가 n=130·244·269 로 전부 CAUGHT).
        ran = ran_count(output)
        suspect = "" if ran == control_ran else f"  · **환경 의심** (Ran {ran} ≠ {control_ran})"
        print(f"{name:36s} {verdict:9s}{touched}{suspect} {len(names):2d} "
              f"{', '.join(n.split('.')[-1] for n in names[:3])}"
              + (f" 외 {len(names) - 3}" if len(names) > 3 else ""))
    # 대조군을 창 **끝에도** 돌린다 (task 7.5.2.4). 창 시작에만 돌리면, 도중에 환경이 무너진 판들이
    # 전부 CAUGHT 로 찍히고 하네스는 그것을 변이의 증거와 못 가른다 — 끝 대조군이 빨가면 창을 통째로 버린다.
    code, output = run(SUITE)
    if code != 0:
        free = shutil.disk_usage(WORK).free
        print(f"\nSTOP — 창 **끝** 무변이 대조군이 빨갛다 (남은 디스크 {free // (1 << 20)} MiB)."
              f" 이 창의 결과는 전부 버린다.\n", output[-3000:])
        raise SystemExit(1)
    print(f"창 끝 control GREEN {output.strip().splitlines()[-1]} (Ran {ran_count(output)})")
    print(f"\nSURVIVED {len(survived)}/{len(chosen)}" + (f": {survived}" if survived else ""))
    shutil.rmtree(WORK, ignore_errors=True)
