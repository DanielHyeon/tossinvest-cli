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
# 작업 자리는 기본 `_work/`(무시되는 디렉터리)다. `A122_HARNESS_WORK` 로 다른 파일시스템을 줄 수 있다 (task 6.4 보수) —
# 이 저장소는 /mnt/D(ntfs-3g) 위라 사본과 전용 GOCACHE 가 거기 있으면 스위트 한 판이 8~10 분, ext4 스크래치면 직접 실행과
# 비슷하다. 산출물의 뜻은 같다 — 사본 · 캐시 자리만 바뀐다.
SP = Path(os.environ.get("A122_HARNESS_WORK") or Path(__file__).resolve().parent / "_work")
SP.mkdir(parents=True, exist_ok=True)
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
        '        if end < 0:\n            # **부분 답을', '        if end < 0:\n            break\n            # **부분 답을')],
    "T7_size_from_content_scan": [(
        '        size = int(match.group("size"))',
        '        size = data.find(b"\\0", position) - position')],
    "T18_no_prefill_of_unanswered_paths": [(
        '    found: dict[str, bytes | None] = {relative: None for relative in wanted}',
        '    found: dict[str, bytes | None] = {}')],
    "U1_leftover_bytes_ignored": [(
        '    if position != len(data):\n        # 정상 응답은', '    if False:\n        # 정상 응답은')],
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
        '        if position + size >= len(data):\n            # 내용과', '        if False:\n            # 내용과')],
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
    # AA3 · AA7 · AA19 는 task 7.5.9 · 7.5.10 · 7.5.16 이 겨누던 줄을 바꿔 다시 걸었다 — 뜻은 같다(원장이 트리 목록을
    # 잊는다 · 기록 직전 거절을 다시 안 묻는다 · 목록 지문에서 종류가 빠진다). 트리 순회(`rglob`)는 추적 목록으로 바뀌었다.
    "AA3_ledger_forgets_tree_walks": [
        ('    _remember("tracked", str(root), outcome)', '    pass')],
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
         '    return _judged_state_moved(root, head, book) or refusal',
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
        ('        len(raw).to_bytes(8, "big") + raw + (b"d" if is_dir else b"f")',
         '        len(raw).to_bytes(8, "big") + raw')],
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
    # --- task 7.5.27 의 AF1~AF11 · 7.5.28 의 AG2~AG5 · AG7~AG9 는 뺐다 (task 7.5.25) — 겨누던 코드
    # (`_git_view_pins` · `_hidden_by_index_flags` · `_ident_go_paths`)가 스냅숏으로 바뀌어 **없다**.
    # 앵커가 없는 변이를 남기면 하네스가 단언에서 멈춘다. 그 변이들의 결과는 7.5.27 · 7.5.28 의 VERIFY 에 있다.
    # --- task 7.5.28: 줄 경계 · 머리 줄의 탭 ---
    "AG1_the_diff_splits_on_every_line_break": [
        ('    diff_lines = process.stdout.decode("utf-8", "strict").split("\\n")',
         '    diff_lines = process.stdout.decode("utf-8", "strict").splitlines()')],
    "AG6_the_header_keeps_its_tab": [
        ('    return value[:-1] if value.endswith("\\t") else value', '    return value')],
    # --- task 7.5.25 · 7.5.31 에서 남은 것 — 겨누는 줄이 7.5.34 뒤에도 같은 뜻으로 있다 ---
    # 나머지 AH · AI 는 7.5.34 가 그 코드를 지워(스냅숏 인덱스 · 대체 저장소 · git 밖 대조) 앵커가 없다. 그 결과는
    # review 7.5.25 · 7.5.31 에 있고, 같은 뜻의 자리는 아래 AJ 가 새 코드에서 다시 겨눈다.
    "AH10_the_fsmonitor_runs": [
        ('SNAPSHOT_PINS = ("-c", "core.fsmonitor=false")', 'SNAPSHOT_PINS = ()')],
    "AH14_an_executable_is_refused": [
        ('GO_FILE_MODES = (b"100644", b"100755")', 'GO_FILE_MODES = (b"100644",)')],
    "AH18_the_loose_header_is_wrong": [
        ('        handle.write(zlib.compress(b"blob %d\\0" % len(data) + data))',
         '        handle.write(zlib.compress(b"blob %d \\0" % len(data) + data))')],
    "AH21_a_failed_rev_parse_is_read": [
        ('    if described.returncode:', '    if False:')],
    "AI1_replace_refs_are_followed": [
        ('os.environ["GIT_NO_REPLACE_OBJECTS"] = "1"\n', '')],
    # --- task 7.5.34: 판정 blob 을 실제 저장소에서 oid 로 찾지 않는다 — 검증한 객체와 격리한 저장소 ---
    # `_verified_objects`
    "AJ1_nothing_asked_still_runs_git": [
        ('    if not wanted:\n        return {}\n    process = subprocess.run(\n        ["git", "cat-file", "--batch"],',
         '    process = subprocess.run(\n        ["git", "cat-file", "--batch"],')],
    "AJ2_a_failed_cat_file_is_read": [
        ('        raise RuntimeError(_first_line(process.stderr, "git cat-file --batch failed"))',
         '        pass')],
    "AJ3_a_missing_header_is_read": [
        ('        if end < 0:\n            raise RuntimeError(f"`git cat-file --batch` stopped before',
         '        if False:\n            raise RuntimeError(f"`git cat-file --batch` stopped before')],
    "AJ4_the_answer_may_name_another_oid": [
        ('header[0] != oid.encode("ascii") or ', '')],
    "AJ5_the_answer_may_be_another_kind": [
        ('header[1] != kind or ', '')],
    "AJ6_the_size_is_not_read_as_a_number": [
        (' or not header[2].isdigit():', ':')],
    "AJ7_a_short_body_is_read": [
        ('        if position + size >= len(data):\n            raise RuntimeError(f"`git cat-file --batch` stopped before',
         '        if False:\n            raise RuntimeError(f"`git cat-file --batch` stopped before')],
    "AJ8_the_terminator_is_not_checked": [
        ('        if data[position + size] != 0x0A:', '        if False:')],
    "AJ9_the_hash_is_not_checked": [
        ('        if hashlib.new(algorithm, kind + b" %d\\0" % size + body).hexdigest() != oid:', '        if False:')],
    "AJ10_the_hash_omits_the_header": [
        ('        if hashlib.new(algorithm, kind + b" %d\\0" % size + body).hexdigest() != oid:',
         '        if hashlib.new(algorithm, body).hexdigest() != oid:')],
    "AJ11_leftover_bytes_are_ignored": [
        ('    if position != len(data):\n        raise RuntimeError(f"`git cat-file --batch` left',
         '    if False:\n        raise RuntimeError(f"`git cat-file --batch` left')],
    # `_verified_go_entries`
    "AJ12_a_failed_ls_tree_is_read": [
        ('            raise RuntimeError(_first_line(listed.stderr, "git ls-tree failed"))', '            pass')],
    "AJ13_a_short_ls_tree_record_is_read": [
        ('            if len(fields) != 3 or not raw_path or not FULL_OID.fullmatch(fields[2]):',
         '            if False:')],
    "AJ14_a_gitlink_is_asked_as_a_tree": [
        ('            if fields[1] == b"tree":', '            if True:')],
    "AJ15_the_commit_line_is_not_read": [
        ('        if not first.startswith(b"tree ") or not FULL_OID.fullmatch(first[5:]):', '        if False:')],
    "AJ16_the_walk_trusts_the_listing": [
        ('            if wanted.get(tree, (b"",))[0] != b"tree":', '            if False:')],
    "AJ17_a_short_tree_entry_is_read": [
        ('                if space < 0 or nul < 0 or nul + 1 + width > len(data):', '                if False:')],
    "AJ18_subtrees_are_not_walked": [
        ('                if mode == TREE_MODE:', '                if False:')],
    "AJ19_the_base_keeps_every_file": [
        ('                elif name.endswith(b".go"):', '                else:')],
    # `_worktree_entries`
    "AJ20_a_failed_ls_files_is_read": [
        ('        raise RuntimeError(_first_line(listed.stderr, "git ls-files failed"))', '        pass')],
    "AJ21_a_short_ls_files_record_is_read": [
        ('        if len(fields) != 4 or not raw_path:', '        if False:')],
    "AJ22_the_worktree_keeps_every_file": [
        ('        if not raw_path.endswith(b".go"):\n            continue', '        if False:\n            continue')],
    "AJ23_an_odd_name_is_decoded_strictly": [
        ('        path = os.fsdecode(raw_path)', '        path = raw_path.decode("utf-8", "strict")')],
    "AJ24_an_unmerged_file_is_taken": [
        ('        if stage != b"0":', '        if False:')],
    "AJ25_a_symlink_is_taken": [
        ('        if mode not in GO_FILE_MODES:', '        if False:')],
    "AJ26_reads_skip_the_ledger": [
        ('            data = _read_regular(root / path)', '            data = _opened_bytes(root / path)')],
    "AJ27_a_sparse_file_is_a_deletion": [
        ('            if tag.upper() == b"S":', '            if False:')],
    "AJ28_assume_unchanged_keeps_the_index_blob": [
        ('            if tag.upper() == b"S":', '            if tag.upper() == b"S" or tag.islower():')],
    "AJ29_the_digest_is_not_a_blob_id": [
        ('        entries[raw_path] = (mode, hashlib.new(algorithm, b"blob %d\\0" % len(data) + data).hexdigest())',
         '        entries[raw_path] = (mode, hashlib.new(algorithm, data).hexdigest())')],
    "AJ30_the_ls_files_runs_the_fsmonitor": [
        ('        ["git", *SNAPSHOT_PINS, "ls-files", "-s", "-v", "-z"],', '        ["git", "ls-files", "-s", "-v", "-z"],')],
    # `_isolated_tree`
    "AJ31_the_split_index_leaks": [
        ('INDEX_WRITE_PINS = (*SNAPSHOT_PINS, "-c", "core.splitIndex=false", "-c", "core.hooksPath=/dev/null")',
         'INDEX_WRITE_PINS = (*SNAPSHOT_PINS, "-c", "core.hooksPath=/dev/null")')],
    "AJ32_the_hooks_run": [
        ('INDEX_WRITE_PINS = (*SNAPSHOT_PINS, "-c", "core.splitIndex=false", "-c", "core.hooksPath=/dev/null")',
         'INDEX_WRITE_PINS = (*SNAPSHOT_PINS, "-c", "core.splitIndex=false")')],
    "AJ33_write_tree_takes_no_pins": [
        ('        ["git", *INDEX_WRITE_PINS, "write-tree", "--missing-ok"],',
         '        ["git", *SNAPSHOT_PINS, "write-tree", "--missing-ok"],')],
    "AJ34_a_failed_update_index_is_ignored": [
        ('        raise RuntimeError(_first_line(built.stderr, "git update-index failed"))', '        pass')],
    "AJ35_a_failed_write_tree_is_ignored": [
        ('        raise RuntimeError(_first_line(written.stderr, "git write-tree failed"))', '        pass')],
    "AJ36_the_write_tree_answer_is_not_read": [
        ('    if not FULL_OID.fullmatch(tree):', '    if False:')],
    "AJ37_absent_blobs_stop_the_tree": [
        ('"write-tree", "--missing-ok"]', '"write-tree"]')],
    # `_isolated_comparison`
    "AJ38_an_option_is_taken_as_a_revision": [
        ('        if revision.startswith("-"):', '        if False:')],
    "AJ39_a_short_rev_parse_is_read": [
        ('    if len(answer) != 1 + 2 * len(revisions) or not all(FULL_OID.fullmatch(item) for item in answer[1:]):',
         '    if False:')],
    "AJ40_unchanged_paths_are_loaded": [
        ('if before.get(path) != after.get(path))', 'if True)')],
    "AJ41_the_target_side_is_not_fetched": [
        ('        for side, label in ((before, "base"), (after, "target" if target else "the index")):',
         '        for side, label in ((before, "base"),):')],
    "AJ42_the_worktree_side_is_fetched_from_the_store": [
        (' and not (side is after and path in disk):', ':')],
    "AJ43_a_gitlink_is_fetched": [
        ('            if path in side and side[path][0] != GITLINK_MODE and not', '            if path in side and not')],
    # AJ44 첫 판(`fetched.get(oid) or disk[path]`)은 살아남았다 — 같은 oid 면 같은 바이트라 고를 갈래가 없었다. 한 표로
    # 바꾸고 그 표에 디스크 바이트를 싣는 줄을 겨눈다.
    "AJ44_the_disk_bytes_are_not_tabled": [
        ('    blobs.update((after[path][1], data) for path, data in disk.items())\n', '')],
    "AJ45_an_inherited_alternate_is_kept": [
        ('if key not in ("GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_DIFF_OPTS")}', 'if key not in ("GIT_DIFF_OPTS",)}')],
    "AJ46_the_store_is_the_real_one": [
        ('        environment["GIT_OBJECT_DIRECTORY"] = str(store)\n', '')],
    # AJ47(두 diff 의 `GIT_INDEX_FILE` 을 뺌)은 살아남았다 — 트리 둘을 견주는 diff 는 인덱스를 안 쓴다. **줄을 지웠다.**
    # `_safe_changed_go_paths` · `_changed_existing_functions`
    "AJ48_the_guard_runs_outside_the_comparison": [
        ('        "--numstat", "-z", *comparison.trees],\n        cwd=root,\n        env=comparison.environment,\n',
         '        "--numstat", "-z", *comparison.trees],\n        cwd=root,\n')],
    "AJ49_the_judgement_runs_outside_the_comparison": [
        ('            *comparison.trees,\n        ],\n        cwd=root,\n        env=comparison.environment,\n',
         '            *comparison.trees,\n        ],\n        cwd=root,\n')],
    "AJ50_the_old_side_is_not_checked": [
        ('        if old_source not in comparison.old:', '        if False:')],
    "AJ51_the_current_side_is_empty": [
        ('            current = _temporary_go(comparison.new[new_source]) if new_source in comparison.new else None',
         '            current = None')],
    # --- 7.5.35: 검증한 트리도 git 이 **쓰는** 모양이어야 한다 · 판정 diff 의 형식은 게이트가 정한다 ---
    # `_verified_go_entries`
    "AK1_any_mode_is_walked": [
        ('                if mode not in CANONICAL_TREE_MODES:', '                if False:')],
    "AK2_the_executable_mode_is_refused": [
        ('CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", b"120000", GITLINK_MODE)',
         'CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"120000", GITLINK_MODE)')],
    "AK3_a_symlink_is_refused": [
        ('CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", b"120000", GITLINK_MODE)',
         'CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", GITLINK_MODE)')],
    "AK4_a_gitlink_is_refused": [
        ('CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", b"120000", GITLINK_MODE)',
         'CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", b"120000")')],
    "AK5_names_are_not_checked": [
        ('                elif not name or b"/" in name or name in (b".", b".."):', '                elif False:')],
    "AK6_an_empty_name_passes": [
        ('                elif not name or b"/" in name or name in (b".", b".."):',
         '                elif b"/" in name or name in (b".", b".."):')],
    "AK7_a_slash_in_a_name_passes": [
        ('                elif not name or b"/" in name or name in (b".", b".."):',
         '                elif not name or name in (b".", b".."):')],
    "AK8_dot_passes": [
        ('                elif not name or b"/" in name or name in (b".", b".."):',
         '                elif not name or b"/" in name or name in (b"..",):')],
    "AK9_dot_dot_passes": [
        ('                elif not name or b"/" in name or name in (b".", b".."):',
         '                elif not name or b"/" in name or name in (b".",):')],
    "AK10_duplicates_pass": [
        ('                elif name in names:', '                elif False:')],
    "AK11_order_is_not_checked": [
        ('                elif key <= last:', '                elif False:')],
    "AK12_order_ignores_the_tree_slash": [
        ('                key = name + b"/" if mode == TREE_MODE else name', '                key = name')],
    "AK13_the_order_does_not_advance": [
        ('                names.add(name)\n                last = key\n', '                names.add(name)\n')],
    "AK14_names_are_not_remembered": [
        ('                names.add(name)\n', '')],
    "AK15_the_refusal_is_silent": [
        ('                if why:\n                    raise RuntimeError(NOT_CANONICAL',
         '                if False:\n                    raise RuntimeError(NOT_CANONICAL')],
    # 7.5.35 적대 재리뷰가 더한 셋: 초기 git 의 `100664`(P2-1) · 붙어 있지 않은 같은 이름(P2-2, 리뷰어 M11) · gitlink 의 정렬 열쇠(P3-1, M1)
    "AK21_early_git_mode_is_refused": [
        ('CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100664", b"100755", b"120000", GITLINK_MODE)',
         'CANONICAL_TREE_MODES = (TREE_MODE, b"100644", b"100755", b"120000", GITLINK_MODE)')],
    "AK22_duplicates_only_next_to_each_other": [
        ('                elif name in names:', '                elif key.rstrip(b"/") == last.rstrip(b"/"):')],
    "AK23_a_gitlink_sorts_like_a_tree": [
        ('                key = name + b"/" if mode == TREE_MODE else name',
         '                key = name + b"/" if mode in (TREE_MODE, GITLINK_MODE) else name')],
    # `_changed_existing_functions` · `_isolated_comparison`
    "AK16_the_source_prefix_is_the_users": [
        ('            "--src-prefix=a/",\n', '')],
    "AK17_the_destination_prefix_is_the_users": [
        ('            "--dst-prefix=b/",\n', '')],
    "AK18_the_colour_is_the_users": [
        ('            "--no-color",\n', '')],
    "AK19_hunks_are_merged_as_the_user_says": [
        ('            "--inter-hunk-context=0",\n', '')],
    "AK20_git_diff_opts_is_passed_on": [
        ('if key not in ("GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_DIFF_OPTS")}',
         'if key not in ("GIT_ALTERNATE_OBJECT_DIRECTORIES",)}')],
    "AD6_the_guard_ignores_renames": [
        ('"--no-ext-diff", "--no-textconv", "--find-renames",',
         '"--no-ext-diff", "--no-textconv",')],
    "AC5_the_hunk_reads_the_base_side_twice": [
        ('''                int(match.group(3)),
                int(match.group(4) or 1),''',
         '''                int(match.group(1)),
                int(match.group(2) or 1),''')],
    # --- task 6.4 (b)(c)(d) — 창의 시작을 잠근다 · 아카이브 날짜는 ASCII 숫자 · 해소기의 실패는 판정이다 ---
    "AL1_archive_digits_are_unicode": [
        ('ARCHIVED_CHANGE = re.compile(r"\\d{4}-\\d{2}-\\d{2}-(?P<change>.+)", re.ASCII)',
         'ARCHIVED_CHANGE = re.compile(r"\\d{4}-\\d{2}-\\d{2}-(?P<change>.+)")')],
    "AL2_the_base_shape_is_not_checked": [
        ('    if not FULL_SHA.fullmatch(candidate):\n        # 이름(`HEAD`', '    if False:\n        # 이름(`HEAD`')],
    # AL3 · AL4 앵커는 6.4 보수(P0-A)에서 대조가 `held` 순회로 바뀌어 다시 걸었다 — 뜻은 같다.
    "AL3_an_uncommitted_edit_moves_the_base": [
        ('        if shown == candidate:\n            continue', '        if True:\n            continue')],
    "AL4_a_new_base_must_be_committed": [
        ('    for place, value in held:', '    for place, value in held or [(relative, b"")]:')],
    "AL5_a_tag_id_is_a_base": [
        ('    if persisted != candidate:\n        # `^{commit}`', '    if False:\n        # `^{commit}`')],
    "AL6_the_base_is_read_at_the_live_head": [
        ('    committed = _committed_bytes(root, head, relative)', '    committed = _committed_bytes(root, "HEAD", relative)')],
    "AL7_the_judged_target_swallows_the_resolver": [
        ('        change_dir = resolve_referenced_change(root, change)\n    except ValueError as exc:\n        return [str(exc)], False',
         '        change_dir = resolve_referenced_change(root, change)\n    except ValueError as exc:\n'
         '        change_dir = root / "openspec" / "changes" / change')],
    # --- task 6.4 보수 (독립 리뷰 둘) — 이동이 대조를 벗지 못한다 · 모양은 전체 일치 ---
    "AL9_the_base_shape_is_a_prefix_match": [
        ('    if not FULL_SHA.fullmatch(candidate):\n        # 이름(`HEAD`', '    if not FULL_SHA.match(candidate):\n        # 이름(`HEAD`')],
    "AL10_no_search_elsewhere": [
        ('else _committed_elsewhere(root, head, change_id, relative)', 'else []')],
    "AL11_a_moved_mismatch_is_accepted": [
        ('        if shown == candidate:\n            continue',
         '        if shown == candidate or place != relative:\n            continue')],
    "AL12_the_head_archive_is_not_listed": [
        ('        if _archived_change_id(entry.removeprefix(ARCHIVE_PREFIX)) == change_id',
         '        if False')],
    "AL13_the_open_place_is_not_asked": [
        ('    places = [f"openspec/changes/{change_id}/base-commit.txt"] + [', '    places = [] + [')],
    "AL14_a_suffix_names_the_same_change": [
        ('        if _archived_change_id(entry.removeprefix(ARCHIVE_PREFIX)) == change_id',
         '        if entry.endswith(change_id)')],
    "AL15_every_mismatch_is_called_a_move": [
        ('        if place == relative:\n            raise ValueError(', '        if False:\n            raise ValueError(')],
    "AL16_a_failed_listing_is_an_empty_archive": [
        ('    if listed.returncode:\n        raise RuntimeError(_first_line(listed.stderr, "git ls-tree failed"))',
         '    if False:\n        raise RuntimeError(_first_line(listed.stderr, "git ls-tree failed"))')],
    "AL8_the_record_path_swallows_the_resolver": [
        ('        change_dir = resolve_referenced_change(root, change)\n    except ValueError as exc:\n        return 1, [str(exc)]',
         '        change_dir = resolve_referenced_change(root, change)\n    except ValueError as exc:\n'
         '        change_dir = root / "openspec" / "changes" / change')],
    # --- task 7.5.9 — 시험 인용은 추적 파일로만 충족된다 ---
    "AM1_the_index_reads_the_disk": [
        ('    index = TestIndex(frozenset(path for path in _tracked(root) if path.name.endswith("_test.go")))',
         '    index = TestIndex(frozenset(root.rglob("*_test.go")))')],
    "AM2_a_qualified_path_ignores_tracking": [
        ('        return qualified if qualified in files and _kind(qualified) == "reg" else None',
         '        return qualified if _kind(qualified) == "reg" else None')],
    "AM3_the_package_ignores_tracking": [
        ('    if local in files and _kind(local) == "reg":', '    if _kind(local) == "reg":')],
    "AM4_a_bare_name_walks_the_disk": [
        ('    matches = [path for path in files if path.name == cited]',
         '    matches = [path for path in root.rglob(cited) if ".git" not in path.parts]')],
    "AM5_a_failed_listing_is_read_as_empty": [
        ('    if process.returncode:\n        return f"rc:{process.returncode}", RuntimeError(',
         '    if False:\n        return f"rc:{process.returncode}", RuntimeError(')],
    "AM6_committed_files_only": [
        ('["git", *SNAPSHOT_PINS, "ls-files", "-z"]', '["git", *SNAPSHOT_PINS, "ls-tree", "-r", "-z", "--name-only", "HEAD"]')],
    "AM7_the_recheck_names_a_path": [
        ("'the tracked file list' if kind == 'tracked' else _shown(root, key)", "_shown(root, key)")],
    "AM8_the_message_says_the_tree": [
        ('                f"function in any tracked file"', '                f"function anywhere in the tree"')],
    "AM9_a_failed_listing_is_swallowed": [
        ('    if isinstance(value, BaseException):\n        raise value\n    return value',
         '    if isinstance(value, BaseException):\n        return []\n    return value')],
    # --- task 7.5.10 — 지문은 단사다 ---
    "AN1_the_listing_joins_with_newlines": [
        ('    encoded = b"".join(\n'
         '        len(raw).to_bytes(8, "big") + raw + (b"d" if is_dir else b"f")\n'
         '        for raw, is_dir in ((name.encode("utf-8"), is_dir) for name, is_dir in entries)\n'
         '    )',
         '    encoded = "\\n".join(f"{name}\\t{\'d\' if is_dir else \'f\'}" for name, is_dir in entries).encode("utf-8")')],
    "AN2_no_length_prefix": [
        ('        len(raw).to_bytes(8, "big") + raw + (b"d" if is_dir else b"f")',
         '        raw + (b"d" if is_dir else b"f")')],
    "AN3_the_tracked_fingerprint_rejoins_paths": [
        ('    return "tracked:" + hashlib.sha256(process.stdout).hexdigest(), listed',
         '    return "tracked:" + hashlib.sha256("\\n".join(map(str, listed)).encode("utf-8", "surrogateescape"))'
         '.hexdigest(), listed')],
    # --- task 7.5.11 — 종류를 못 물은 항목은 이름 댄 결함이다 ---
    "AO1_the_kind_through_is_dir": [
        ('            try:\n                mode = os.stat(child).st_mode\n            except OSError as exc:\n'
         '                raise UnstatableEntry(\n'
         '                    exc.errno, f"cannot tell what `{child.name}` is: {exc.strerror}", str(child)) from exc\n'
         '            entries.append((child.name, stat.S_ISDIR(mode)))',
         '            entries.append((child.name, child.is_dir()))')],
    "AO2_an_unknown_kind_is_a_file": [
        ('            except OSError as exc:\n                raise UnstatableEntry(\n'
         '                    exc.errno, f"cannot tell what `{child.name}` is: {exc.strerror}", str(child)) from exc',
         '            except OSError as exc:\n                mode = 0')],
    "AO3_the_raw_error_escapes": [
        ('                raise UnstatableEntry(\n'
         '                    exc.errno, f"cannot tell what `{child.name}` is: {exc.strerror}", str(child)) from exc',
         '                raise')],
    # --- task 7.5.15 — 판정 경로의 자식 프로세스는 시한을 받는다 (execution_baseline.py 쪽은 7515_mut.py) ---
    "AP1_the_guard_diff_has_no_timeout": [
        ('        # 시한은 판정 diff(`_changed_existing_functions`)와 같은 값이다 (task 7.5.15) — 같은 두 트리를 견준다.\n'
         '        timeout=30,\n', '')],
    # --- task 7.5.16 — 기록 직전: 역사 → 원장 → 거절 ---
    "AQ1_the_refusal_speaks_before_the_ledger": [
        ('    return _judged_state_moved(root, head, book) or refusal',
         '    return refusal or _judged_state_moved(root, head, book)')],
    "AQ2_the_ledger_is_not_asked_before_writing": [
        ('    return _judged_state_moved(root, head, book) or refusal', '    return refusal')],
    # --- task 7.5.9~7.5.16 보수 (독립 리뷰 둘) ---
    "MZ1_a_failed_start_is_an_empty_listing": [
        ('        return f"{type(exc).__name__}", exc',
         '        return "tracked:" + hashlib.sha256(b"").hexdigest(), []')],
    "MR1_the_qualified_path_is_not_resolved": [
        ('            qualified = root / Path(os.path.realpath(root / cited)).relative_to(os.path.realpath(root))',
         '            qualified = root / cited')],
    "MR2_every_archive_entry_is_asked_its_kind": [
        ('    for name in names:\n        if _archived_change_id(name) != change:\n            continue\n'
         '        kind = _kind(archive / name)\n',
         '    for name in names:\n        kind = _kind(archive / name)\n'
         '        if kind and _archived_change_id(name) != change:\n            continue\n')],
    "MR3_the_tracked_listing_is_not_pinned": [
        ('["git", *SNAPSHOT_PINS, "ls-files", "-z"]', '["git", "ls-files", "-z"]')],
    "MR4_the_ledger_forgets_the_archive_names": [
        ('    _remember("names", str(path), outcome)', '    pass')],
    "MR5_an_unknown_own_entry_is_not_archived": [
        ('        if not kind:\n            # 이 change 의 id 를 단 항목인데', '        if False:\n            # 이 change 의 id 를 단 항목인데')],
    # (첫 판에 적은 "밖의 경로를 `root / cited` 로 둔다" 는 동등 변이다 — 추적 목록은 전부 뿌리 아래라 어차피 None.
    #  돌리기 전에 그렇게 판단해 바꿨다: 밖의 경로가 **결함**으로 올라가는 모양이 이 갈래가 막는 것이다.)
    "MR6_a_path_outside_raises": [
        ('        except ValueError:\n            return None\n        return qualified if qualified in files',
         '        except ValueError:\n            raise\n        return qualified if qualified in files')],
}


def setup() -> Path:
    if WORK.exists():
        shutil.rmtree(WORK)
    shutil.copytree(REPO / "tools" / "logic-map", WORK / "logic-map",
                    ignore=shutil.ignore_patterns("__pycache__"))
    (WORK / "sdd").mkdir(parents=True)
    shutil.copy(REPO / "tools" / "sdd" / "sdd_doctor.py", WORK / "sdd")
    # 두 해소기를 한 표로 도는 시험(task 6.4(f))이 `check_analysis.py` 옆 디렉터리의 `gate.sh` 를 부른다 — 사본에 없으면
    # 무변이 대조군부터 빨갛다.
    shutil.copy(REPO / "tools" / "gate.sh", WORK / "gate.sh")
    return WORK / "logic-map" / "check_analysis.py"


# 스위트는 임시 저장소마다 `go run ./tools/logic-map` 을 부르고, go 의 캐시 키에 **디렉터리 경로**가 들어가서
# 한 판이 공유 캐시(`~/.cache/go-build`)를 약 375 MB 씩 불렸다 — 판 170 개면 60 GB 가 넘고, 2026-09-24 에 디스크가
# 두 번 0 이 됐다 (task 7.5.30). `-trimpath` 는 경로를 키에서 빼고(둘째 판 +1 MB), 캐시는 이 하네스 전용으로
# `_work/` 아래(무시되는 디렉터리)에 둔다 — 사람의 공유 캐시를 건드리지 않는다.
GO_ENV = {"GOFLAGS": "-trimpath", "GOCACHE": str(SP / "gocache")}


def run(names: list[str]) -> tuple[int, str]:
    process = subprocess.run([sys.executable, "-m", "unittest", *names],
                             cwd=WORK / "logic-map", capture_output=True, text=True, timeout=5400,
                             env={**os.environ, **GO_ENV})
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
