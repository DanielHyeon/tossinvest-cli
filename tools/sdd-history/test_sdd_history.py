import io
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import agent_save_hook
import collect_events
import context_graph_sync
import git_commit_event
import ingest_typedb
import refresh_indexes


class EventTests(unittest.TestCase):
    def test_collect_masks_sensitive_values(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            ok = collect_events.collect_event(
                {
                    "type": "agent-save",
                    "summary": "Bearer abcdefghijklmnopqrstuvwxyz0123456789 user@example.com",
                    "token": "do-not-store",
                },
                root,
            )
            self.assertTrue(ok)
            stored = json.loads((root / "events.jsonl").read_text(encoding="utf-8"))
            self.assertNotIn("do-not-store", json.dumps(stored))
            self.assertNotIn("user@example.com", json.dumps(stored))

    def test_collect_masks_short_secret_assignment(self):
        value = collect_events.mask_string("password=abc123 api_key: short")
        self.assertNotIn("abc123", value)
        self.assertNotIn("short", value)

    def test_invalid_event_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            self.assertFalse(collect_events.collect_event({}, Path(tmp)))

    def test_typedb_loader_skips_broken_json(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "events.jsonl"
            path.write_text('{"type":"commit","summary":"ok"}\nnot-json\n', encoding="utf-8")
            self.assertEqual(len(ingest_typedb.load_events(path)), 1)

    def test_typedb_pending_events_are_incremental(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "events.jsonl"
            path.write_text('{"type":"commit","summary":"one"}\n', encoding="utf-8")
            first, offsets = ingest_typedb.pending_events(path, {})
            second, _ = ingest_typedb.pending_events(path, offsets)
            self.assertEqual(len(first), 1)
            self.assertEqual(second, [])

    def test_context_graph_source_is_part_of_event_identity(self):
        requests = []

        class Response(io.BytesIO):
            def __enter__(self):
                return self

            def __exit__(self, *_args):
                self.close()

        def open_request(request, timeout):
            self.assertEqual(timeout, 10)
            requests.append(json.loads(request.data))
            return Response(b'{"results":[],"errors":[]}')

        event = {"type": "commit", "summary": "same"}
        environment = {
            "NEO4J_USERNAME": "user",
            "NEO4J_PASSWORD": "pass",
            "NEO4J_HTTP_URI": "http://localhost:7474",
        }
        with mock.patch.dict(os.environ, environment, clear=False), mock.patch(
            "context_graph_sync.urllib.request.urlopen",
            side_effect=open_request,
        ):
            context_graph_sync.ingest_events_http("commits", [event])
            context_graph_sync.ingest_events_http("agent-saves", [event])
        first = requests[0]["statements"][0]["parameters"]
        second = requests[1]["statements"][0]["parameters"]
        self.assertNotEqual(first["hash"], second["hash"])
        self.assertEqual(first["source"], "tossos-commits")
        self.assertEqual(second["source"], "tossos-agent-saves")

    def test_context_graph_remasks_legacy_untrusted_jsonl(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "events.jsonl"
            path.write_text(
                '{"type":"commit","summary":"password=short user@example.com"}\n',
                encoding="utf-8",
            )
            events, _ = context_graph_sync.read_new_events(path, 0)
            encoded = json.dumps(events)
            self.assertNotIn("short", encoded)
            self.assertNotIn("user@example.com", encoded)

    def test_context_graph_uses_private_ephemeral_credential_free_workspace(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            events = root / "events.jsonl"
            events.write_text('{"type":"commit","summary":"ok"}\n', encoding="utf-8")

            def fake_run(command, **_kwargs):
                output = Path(command[command.index("--output-dir") + 1])
                output.mkdir(parents=True, exist_ok=True)
                (output / ".env").write_text("NEO4J_PASSWORD=secret\n", encoding="utf-8")
                self.assertIn("--dry-run", command)
                self.assertNotIn("--neo4j-aura-env", command)
                self.assertNotIn("pass", " ".join(command))
                return mock.Mock(returncode=0)

            environment = {
                "NEO4J_URI": "bolt://localhost:7687",
                "NEO4J_USERNAME": "user",
                "NEO4J_PASSWORD": "pass",
            }
            with mock.patch.object(context_graph_sync, "ROOT", root), mock.patch.object(
                context_graph_sync,
                "CHECKPOINTS",
                root / "checkpoints",
            ), mock.patch(
                "context_graph_sync.shutil.which",
                return_value="/bin/create-context-graph",
            ), mock.patch(
                "context_graph_sync.subprocess.run",
                side_effect=fake_run,
            ), mock.patch(
                "context_graph_sync.ingest_events_http",
            ), mock.patch.dict(
                os.environ,
                environment,
                clear=False,
            ):
                self.assertEqual(context_graph_sync.sync("commits", events), 0)
            self.assertEqual(list(root.rglob(".env")), [])
            self.assertEqual(
                list((root / "tmp" / "create-context-graph" / "runs").iterdir()),
                [],
            )

    def test_context_graph_failed_adapter_keeps_separate_retry_offset(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            events = root / "events.jsonl"
            events.write_text('{"type":"commit","summary":"ok"}\n', encoding="utf-8")
            environment = {
                "NEO4J_USERNAME": "user",
                "NEO4J_PASSWORD": "pass",
            }
            failed = mock.Mock(returncode=1)
            with mock.patch.object(context_graph_sync, "ROOT", root), mock.patch.object(
                context_graph_sync,
                "CHECKPOINTS",
                root / "checkpoints",
            ), mock.patch(
                "context_graph_sync.shutil.which",
                return_value="/bin/create-context-graph",
            ), mock.patch(
                "context_graph_sync.subprocess.run",
                return_value=failed,
            ), mock.patch(
                "context_graph_sync.ingest_events_http",
            ), mock.patch.dict(
                os.environ,
                environment,
                clear=False,
            ):
                self.assertEqual(context_graph_sync.sync("commits", events), 0)
                self.assertEqual(context_graph_sync.load_offset("commits", "ccg"), 0)
                self.assertGreater(context_graph_sync.load_offset("commits", "http"), 0)

    def test_agent_save_hook_keeps_local_event_when_refresh_fails(self):
        with tempfile.TemporaryDirectory() as tmp, mock.patch(
            "agent_save_hook.refresh_indexes.schedule",
            side_effect=RuntimeError("offline"),
        ):
            root = Path(tmp)
            result = agent_save_hook.run(
                {
                    "agent": "password=short",
                    "tool_input": {"file_path": str(agent_save_hook.ROOT / "CLAUDE.md")},
                },
                root,
            )
            self.assertEqual(result, 0)
            stored = (root / "agent-saves.jsonl").read_text(encoding="utf-8")
            self.assertNotIn("password=short", stored)

    def test_commit_hook_keeps_local_event_when_refresh_fails(self):
        with tempfile.TemporaryDirectory() as tmp, mock.patch(
            "git_commit_event.refresh_indexes.schedule",
            side_effect=RuntimeError("offline"),
        ), mock.patch(
            "git_commit_event.paths",
            return_value=["CLAUDE.md"],
        ):
            root = Path(tmp)
            result = git_commit_event.run(
                ("a" * 40, "codex", "docs: safe", ""),
                root,
            )
            self.assertEqual(result, 0)
            stored = json.loads((root / "commits.jsonl").read_text(encoding="utf-8"))
            self.assertEqual(stored["paths"], ["CLAUDE.md"])

    def test_commit_metadata_failure_warns_without_silent_success(self):
        with mock.patch("git_commit_event.head", return_value=None), mock.patch(
            "sys.stderr",
            new_callable=io.StringIO,
        ) as stderr:
            self.assertEqual(git_commit_event.run(), 1)
            self.assertIn("cannot read", stderr.getvalue())

    def test_refresh_requests_are_coalesced_behind_one_worker(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            with mock.patch.object(refresh_indexes, "STATE", state), mock.patch.object(
                refresh_indexes,
                "QUEUE",
                state / "queue",
            ), mock.patch.object(
                refresh_indexes,
                "LOCK",
                state / "lock",
            ), mock.patch(
                "refresh_indexes.subprocess.Popen",
            ) as popen:
                popen.return_value.pid = os.getpid()
                refresh_indexes.schedule("agent-saves")
                refresh_indexes.schedule("agent-saves")
            self.assertEqual(popen.call_count, 1)
            self.assertEqual(len(list((state / "queue").glob("*.json"))), 2)

    def test_refresh_lock_is_not_stolen_from_live_owner(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            with mock.patch.object(refresh_indexes, "STATE", state), mock.patch.object(
                refresh_indexes,
                "LOCK",
                state / "lock",
            ):
                refresh_indexes.LOCK.write_text(
                    json.dumps(
                        {
                            "pid": os.getpid(),
                            "start": refresh_indexes.process_start_identity(os.getpid()),
                            "token": "owner",
                        }
                    ),
                    encoding="utf-8",
                )
                old = 1
                os.utime(refresh_indexes.LOCK, (old, old))
                self.assertIsNone(refresh_indexes.claim())
                self.assertEqual(refresh_indexes.lock_owner()["token"], "owner")

    def test_tokenless_worker_cannot_remove_another_lock(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            with mock.patch.object(refresh_indexes, "STATE", state), mock.patch.object(
                refresh_indexes,
                "LOCK",
                state / "lock",
            ):
                refresh_indexes.LOCK.write_text(
                    json.dumps({"pid": os.getpid(), "token": "owner"}),
                    encoding="utf-8",
                )
                self.assertEqual(refresh_indexes.run_worker(), 2)
                self.assertTrue(refresh_indexes.LOCK.exists())


if __name__ == "__main__":
    unittest.main()
