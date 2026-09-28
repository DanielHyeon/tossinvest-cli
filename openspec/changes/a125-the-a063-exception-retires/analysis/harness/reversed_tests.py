
A063 = "a063-align-attestation-renewal-profile"


def _a063_fixture(*, deleted: bool = False) -> tuple[tempfile.TemporaryDirectory, Path, str, str]:
    """a063 의 옛 이관 모양을 **모듈 없이** 세운다 (a125). P(base) → E → S → H(이관 기록) → map, detached.

    기록(`execution-baseline.json`)은 a120 초안기가 쓰던 키 집합 그대로의 정적 JSON 이다 — 게이트가 그것을 **읽지
    않는다**는 것이 이 픽스처로 재는 사실이다(design D2). 번들은 일반 규칙이 요구하는 모양으로 둔다: 창은 P → 워킹트리
    이고, `Attest` 가 살아 있으면 오늘 소스의 `revision: current`, 지워졌으면 P 의 소스를 적은 `revision: base` 다.
    """
    raw = tempfile.TemporaryDirectory()
    root = _init_fixture(raw)
    change = root / "openspec" / "changes" / A063
    change.mkdir(parents=True)
    (change / "base-commit.txt").write_text("pending\n")
    target = root / "internal" / "soak" / "attest.go"
    target.parent.mkdir(parents=True)
    target.write_text("package soak\nfunc Attest() int { return 1 }\n")
    p = _commit_all(root, "P")
    at_base = check_analysis.go_functions(target, root)[0]
    (change / "base-commit.txt").write_text(p + "\n")
    target.write_text("package soak\nfunc Attest() int { return 2 }\n")
    e = _commit_all(root, "E")
    target.write_text("package soak\n" if deleted else "package soak\nfunc Attest() int { return 3 }\n")
    s = _commit_all(root, "S")
    record = {
        "adversarial_review_path": f"openspec/changes/{A063}/analysis/adoption/adversarial-review.md",
        "adversarial_review_sha256": "0" * 64, "change": A063, "execution_base": e,
        "gstack_review_path": f"openspec/changes/{A063}/analysis/adoption/gstack-review.md",
        "gstack_review_sha256": "1" * 64,
        "inherited_history_disposition": "committed historical work; missing original analysis remains debt",
        "ledger_path": f"openspec/changes/{A063}/analysis/execution-baseline-ledger.json", "ledger_sha256": "2" * 64,
        "planning_base": p, "pre_edit_provenance": "retrospective-exception", "schema": 1, "source_commit": s,
        "source_tree": subprocess.check_output(["git", "rev-parse", f"{s}^{{tree}}"], cwd=root, text=True).strip(),
    }
    (change / "execution-baseline.json").write_text(json.dumps(record, sort_keys=True))
    _commit_all(root, "H: the leftover adoption record")
    value = dict(at_base) if deleted else check_analysis.go_functions(target, root)[0]
    value.update({"package": "soak", "signature": "Attest(params=0, results=1)", "branches": []})
    if deleted:
        value["revision"] = "base"
    bundle = change / "analysis" / "function-logic" / "internal-soak--attest"
    bundle.mkdir(parents=True)
    (bundle / "ast.json").write_text(json.dumps(value))
    (bundle / "function-logic-map.md").write_text(
        "# Function Logic Map: `Attest`\ninternal/soak/attest.go\n## Inputs and invariants\nevidence\n"
        "## Branches and early returns\nevidence\n## Calls and live bindings\nevidence\n"
        "## State mutations and fallbacks\nevidence\n## Safety conclusion\nevidence\n")
    (bundle / "branch-test-map.md").write_text("# Branch Test Map: `Attest`\n| B1 | leaf | test | yes | yes |\n")
    (bundle / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/soak/attest.go\n")
    _commit_all(root, "map")
    subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
    return raw, root, p, e


def _a063_cli(root: Path) -> tuple[int, str]:
    output = io.StringIO()
    with mock.patch.object(sys, "argv", ["check_analysis.py", "--change", A063, "--root", str(root)]), \
            redirect_stdout(output):
        code = check_analysis.main()
    return code, output.getvalue()


def _recommit_detached(root: Path, subject: str) -> None:
    _commit_all(root, subject)
    subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)


ADOPTION_WORDS = ("adoption", "audited source-commit", "execution-baseline")


class TheA063ExceptionIsRetired(unittest.TestCase):
    """a125 — a120 의 a063 전용 실행 기준선 이관 특례를 지웠다. 특례를 못 박던 시험은 지우지 않고 **반전**했다:
    같은 픽스처(옛 이관 모양)에서 이제 일반 규칙이 판정한다는 것을 못 박는다(design D3). 옛 판본에서는 전부 빨갛다 —
    기록이 있으면 이관 판정기가 먼저 불려 `invalid execution-baseline adoption` 으로 멈췄다."""

    def assertNoAdoptionWord(self, text: str) -> None:
        for word in ADOPTION_WORDS:
            self.assertNotIn(word, text)

    def test_main_prints_only_the_ordinary_success_line(self) -> None:
        # 반전: test_main_distinguishes_ordinary_adoption_and_invalid_results — 옛 문맥 키가 들어와도 이관 줄은 없다.
        output = io.StringIO()

        def fake_check(change: str, root: Path, context: dict[str, object]) -> list[str]:
            context["execution_baseline_adoption"] = True
            return []
        with mock.patch.object(sys, "argv", ["check_analysis.py", "--change", "fixture"]), \
                mock.patch("check_analysis.check", side_effect=fake_check), redirect_stdout(output):
            self.assertEqual(check_analysis.main(), 0)
        self.assertIn("evidence complete or diff-proven exempt", output.getvalue())
        self.assertNoAdoptionWord(output.getvalue())

    def test_a063_with_a_leftover_record_is_judged_from_its_base_commit(self) -> None:
        # 반전: test_valid_adoption_uses_e… · test_main_real_adoption_prints_exception_label… ·
        # test_the_adoption_window_ends_at_the_audited_source_commit · test_the_adoption_path_does_not_advise…
        raw, root, p, e = _a063_fixture()
        with raw:
            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check(A063, root, context), [])
            self.assertEqual(context.get("effective_base"), p)
            self.assertEqual(context.get("landing"), "")
            self.assertEqual(context.get("required_count"), 1)
            for key in ("execution_baseline_adoption", "adoption_source"):
                self.assertNotIn(key, context)
            code, printed = _a063_cli(root)
            self.assertEqual(code, 0, printed)
            self.assertIn(f"base {p[:12]} → working tree", printed)
            self.assertNoAdoptionWord(printed)

    def test_a063s_modified_function_is_still_required(self) -> None:
        raw, root, _, _ = _a063_fixture()
        with raw:
            shutil.rmtree(root / "openspec" / "changes" / A063 / "analysis" / "function-logic")
            _recommit_detached(root, "remove the evidence")
            errors = check_analysis.check(A063, root)
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("missing Function Logic Map for 1 function(s)", errors[0])
            self.assertIn("internal/soak/attest.go:Attest", errors[0])

    def test_a_deleted_function_needs_a_base_revision_bundle_from_the_base_commit(self) -> None:
        # 반전: test_real_adoption_deleted_function_requires_base_revision — 기준은 E 가 아니라 P 다.
        raw, root, _, _ = _a063_fixture(deleted=True)
        with raw:
            self.assertEqual(check_analysis.check(A063, root), [])
            path = root / "openspec" / "changes" / A063 / "analysis" / "function-logic" / "internal-soak--attest" / "ast.json"
            value = json.loads(path.read_text()); value["revision"] = "current"; path.write_text(json.dumps(value))
            _recommit_detached(root, "wrong deletion revision")
            self.assertTrue(any("AST revision must be base" in error for error in check_analysis.check(A063, root)))

    def test_sdd_base_ref_accepts_only_the_base_commit(self) -> None:
        # 반전: test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back.
        raw, root, p, e = _a063_fixture()
        with raw:
            with mock.patch.dict("os.environ", {"SDD_BASE_REF": p}, clear=False):
                self.assertEqual(check_analysis.check(A063, root), [])
            for bad in (e, "0" * 40, "HEAD"):
                with mock.patch.dict("os.environ", {"SDD_BASE_REF": bad}, clear=False):
                    self.assertTrue(any("cannot derive" in error for error in check_analysis.check(A063, root)), bad)
            (root / "openspec" / "changes" / A063 / "execution-baseline.json").write_text("{}")
            _recommit_detached(root, "a broken leftover record")
            with mock.patch.dict("os.environ", {"SDD_BASE_REF": p}, clear=False):
                self.assertEqual(check_analysis.check(A063, root), [])

    def test_a_stale_current_hash_fails(self) -> None:
        raw, root, _, _ = _a063_fixture()
        with raw:
            path = root / "openspec" / "changes" / A063 / "analysis" / "function-logic" / "internal-soak--attest" / "ast.json"
            value = json.loads(path.read_text()); value["source_sha256"] = "0" * 64; path.write_text(json.dumps(value))
            _recommit_detached(root, "stale map")
            errors = check_analysis.check(A063, root)
            self.assertTrue(any("AST source hash is stale" in e or "AST hash does not match" in e for e in errors), errors)

    def test_a_reference_beside_local_maps_is_refused(self) -> None:
        raw, root, _, _ = _a063_fixture()
        with raw:
            (root / "openspec" / "changes" / A063 / "analysis" / "function-logic-reference.txt").write_text("other\n")
            _recommit_detached(root, "conflicting reference")
            self.assertIn("function-logic reference cannot coexist with local function-logic evidence",
                          check_analysis.check(A063, root))

    def test_a063_records_and_is_judged_by_the_landing_rules(self) -> None:
        # 반전: test_a_landing_record_is_refused_in_the_adoption_path · test_the_recorder_refuses_the_adoption_path_too.
        raw, root, _, _ = _a063_fixture()
        with raw:
            code, lines = check_analysis.record_landing(A063, root)
            self.assertEqual(code, 0, lines)
            landing = root / "openspec" / "changes" / A063 / check_analysis.LANDING_FILE
            self.assertTrue(landing.is_file(), lines)
            _recommit_detached(root, "record the landing")
            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check(A063, root, context), [])
            self.assertEqual(context.get("landing"), landing.read_text().strip())
            self.assertNoAdoptionWord("\n".join(lines))

    def test_an_undecodable_landing_record_is_named_by_the_landing_rules(self) -> None:
        # 반전: test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised.
        raw, root, _, _ = _a063_fixture()
        with raw:
            (root / "openspec" / "changes" / A063 / check_analysis.LANDING_FILE).write_bytes(b"\xff\xfe not utf-8\n")
            _recommit_detached(root, "an undecodable landing record")
            errors = check_analysis.check(A063, root)
            self.assertTrue(any("landing point is not UTF-8" in error for error in errors), errors)
            code, printed = _a063_cli(root)
            self.assertEqual(code, 1, printed)
            self.assertNoAdoptionWord(printed)

    def test_an_archived_a063_is_rechecked_by_its_id_on_the_general_path(self) -> None:
        # 반전: test_an_archived_adoption_is_rechecked_by_its_id.
        raw, root, p, _ = _a063_fixture()
        with raw:
            archived = root / "openspec" / "changes" / "archive" / f"2026-09-11-{A063}"
            archived.parent.mkdir(parents=True)
            subprocess.run(["git", "mv", f"openspec/changes/{A063}", archived.relative_to(root).as_posix()],
                           cwd=root, check=True)
            _recommit_detached(root, "archive a063")
            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check(A063, root, context), [])
            self.assertEqual(context.get("effective_base"), p)
            self.assertNotIn("execution_baseline_adoption", context)

    def test_a_copied_record_changes_nothing_for_another_id(self) -> None:
        # 반전: test_a_copied_adoption_record_does_not_make_another_change_a063 — 이제 거절할 특례 자체가 없다.
        raw, root, _, _ = _a063_fixture()
        with raw:
            original = root / "openspec" / "changes" / A063
            shutil.copytree(original, root / "openspec" / "changes" / "a130-copy")
            shutil.copytree(original, root / "openspec" / "changes" / "archive" / "2026-09-11-a131-copy")
            _recommit_detached(root, "copy a063 whole under two other ids")
            for change in (A063, "a130-copy", "a131-copy"):
                self.assertEqual(check_analysis.check(change, root), [], change)

    def test_the_exception_module_and_its_suite_are_gone(self) -> None:
        # 반전: test_the_execution_baseline_suite_pins_itself — 판정 모듈이 그 모듈을 부르지도, 저장소에 두지도 않는다.
        here = Path(check_analysis.__file__).resolve().parent
        self.assertFalse((here / "execution_baseline.py").exists())
        self.assertFalse((here / "test_execution_baseline.py").exists())
        tree = ast.parse(Path(check_analysis.__file__).read_text(encoding="utf-8"))
        imported = {alias.name for node in ast.walk(tree) if isinstance(node, (ast.Import, ast.ImportFrom))
                    for alias in node.names} | {node.module for node in ast.walk(tree)
                                                if isinstance(node, ast.ImportFrom) and node.module}
        self.assertNotIn("execution_baseline", imported)
        names = {node.id for node in ast.walk(tree) if isinstance(node, ast.Name)} | {
            node.attr for node in ast.walk(tree) if isinstance(node, ast.Attribute)} | {
            node.name for node in ast.walk(tree) if isinstance(node, (ast.FunctionDef, ast.ClassDef))} | {
            target.id for node in ast.walk(tree) if isinstance(node, ast.Assign)
            for target in node.targets if isinstance(target, ast.Name)}
        self.assertEqual(sorted(name for name in names if "adopt" in name.lower() or "audited" in name.lower()), [])
        strings = [node.value for node in ast.walk(tree) if isinstance(node, ast.Constant) and isinstance(node.value, str)]
        self.assertEqual([text for text in strings if any(word in text for word in ADOPTION_WORDS)], [])

    def test_a_leftover_record_changes_neither_verdict_nor_context_nor_output(self) -> None:
        # design D2 · 1.1 논증의 제거 뒤 핀: 같은 역사에서 기록만 지웠을 때 판정 줄 · 문맥 · 창 줄이 같다.
        raw, root, _, _ = _a063_fixture()
        with raw:
            def judged() -> tuple[list[str], dict[str, object], int, str]:
                context: dict[str, object] = {}
                verdict = check_analysis.check(A063, root, context)
                context.pop("head", None)
                head = check_analysis._head_commit(root)
                code, printed = _a063_cli(root)
                return verdict, context, code, printed.replace(head[:12], "<head>")
            with_record = judged()
            subprocess.run(["git", "rm", "-q", f"openspec/changes/{A063}/execution-baseline.json"], cwd=root, check=True)
            _recommit_detached(root, "the record is deleted")
            self.assertEqual(with_record, judged())
            self.assertEqual(with_record[0], [])

    def test_an_irregular_leftover_record_is_not_read_either(self) -> None:
        # freeze F9 — "안 읽음" 의 핀. 옛 판정기는 lstat 뒤 정규 파일이 아니면 거절했다. 읽는 코드가 없으면 모양이 무엇이든
        # 일반 판정이 나와야 한다: FIFO(열면 멎는다) · 디렉터리 · 저장소 밖 심링크 · 해독 불가 바이트.
        for shape in ("fifo", "directory", "symlink", "undecodable"):
            with self.subTest(shape=shape):
                raw, root, p, _ = _a063_fixture()
                with raw:
                    record = root / "openspec" / "changes" / A063 / "execution-baseline.json"
                    record.unlink()
                    if shape == "fifo":
                        os.mkfifo(record)
                    elif shape == "directory":
                        record.mkdir(); (record / "keep").write_text("x")
                    elif shape == "symlink":
                        record.symlink_to("/dev/zero")
                    else:
                        record.write_bytes(b"\xff\xfe{")
                    if shape != "fifo":
                        _recommit_detached(root, f"a {shape} record")
                    context: dict[str, object] = {}
                    self.assertEqual(check_analysis.check(A063, root, context), [])
                    self.assertEqual(context.get("effective_base"), p)
