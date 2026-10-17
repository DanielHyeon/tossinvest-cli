"""Codex 가 TossOS GBrain wrapper 를 정확히 하나의 등록으로만 띄우는지 고정하는 회귀 시험.

정적 시험(저장소 `.codex/config.toml`)과 호스트 로더 증거(`fixtures/codex_gbrain_registration.json`,
a119 `analysis/host-evidence.md` §3.1)를 구분함. 정적 개수가 호스트가 잰 개수·출처 층과 어긋나면
어느 한쪽이 낡은 것이므로 실패시킴. 다른 에이전트 소유 등록(`.mcp.json`)은 읽지도 고치지도 않음.
"""

import json
import tomllib
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CODEX_CONFIG = ROOT / ".codex" / "config.toml"
FIXTURE = Path(__file__).resolve().parent / "fixtures" / "codex_gbrain_registration.json"
WRAPPER = "tools/sdd/gbrain_project.py"


def launches_wrapper(server: dict) -> bool:
    """서버 정의가 프로젝트 wrapper 의 serve 를 실행하는지 판정함."""
    argv = [server.get("command", ""), *server.get("args", [])]
    return any(Path(part).as_posix().endswith(WRAPPER) for part in argv) and "serve" in argv


def launches_raw_gbrain(server: dict) -> bool:
    """wrapper 를 거치지 않고 gbrain 실행 파일을 직접 띄우는지 판정함(프로젝트 flock 우회)."""
    return Path(server.get("command", "")).name == "gbrain"


def enabled_wrapper_servers(config: dict) -> list[str]:
    """활성 상태로 wrapper 를 띄우는 MCP 서버 이름 목록을 반환함."""
    return sorted(
        name
        for name, server in config.get("mcp_servers", {}).items()
        if server.get("enabled", True) and launches_wrapper(server)
    )


class CodexGbrainRegistrationTests(unittest.TestCase):
    def setUp(self):
        self.config = tomllib.loads(CODEX_CONFIG.read_text(encoding="utf-8"))
        self.fixture = json.loads(FIXTURE.read_text(encoding="utf-8"))

    def test_codex_layer_registers_the_wrapper_exactly_once(self):
        self.assertEqual(enabled_wrapper_servers(self.config), ["gbrain"])
        server = self.config["mcp_servers"]["gbrain"]
        self.assertEqual(server["command"], "python3")
        self.assertEqual(server["args"], [WRAPPER, "serve"])

    def test_no_codex_server_bypasses_the_project_lock(self):
        # raw `gbrain serve` 는 wrapper 의 kernel flock·exit 75 busy 계약을 건너뜀.
        bypassing = [
            name
            for name, server in self.config.get("mcp_servers", {}).items()
            if launches_raw_gbrain(server)
        ]
        self.assertEqual(bypassing, [])

    def test_static_count_matches_recorded_host_loader_evidence(self):
        # 이 비교는 표류 고정(drift pin)이며 호스트 런타임 로딩의 검증이 아님.
        self.assertEqual(self.fixture["source_layer"], ".codex/config.toml")
        self.assertTrue(self.fixture["source_layer_basis"].startswith("inferred_by_elimination"))
        self.assertEqual(self.fixture["effective_wrapper_registrations"], 1)
        self.assertEqual(
            len(enabled_wrapper_servers(self.config)),
            self.fixture["effective_wrapper_registrations"],
        )

    def test_duplicate_registration_is_rejected(self):
        # 음성 대조: 같은 wrapper 를 이름만 달리해 두 번 등록한 설정은 둘로 세어져야 함.
        duplicate = tomllib.loads(
            CODEX_CONFIG.read_text(encoding="utf-8")
            + '\n[mcp_servers.gbrain_copy]\ncommand = "python3"\n'
            + f'args = ["./{WRAPPER}", "serve"]\n'
        )
        self.assertEqual(enabled_wrapper_servers(duplicate), ["gbrain", "gbrain_copy"])

    def test_disabled_copy_and_raw_launch_are_classified(self):
        disabled = tomllib.loads(
            '[mcp_servers.old]\ncommand = "python3"\n'
            f'args = ["{WRAPPER}", "serve"]\nenabled = false\n'
            '[mcp_servers.raw]\ncommand = "/usr/local/bin/gbrain"\nargs = ["serve"]\n'
        )
        self.assertEqual(enabled_wrapper_servers(disabled), [])
        self.assertTrue(launches_raw_gbrain(disabled["mcp_servers"]["raw"]))


if __name__ == "__main__":
    unittest.main()
