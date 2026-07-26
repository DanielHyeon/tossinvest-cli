from __future__ import annotations

import unittest

from gbrain_project import GBRAIN_HOME, project_environment


class GBrainProjectTest(unittest.TestCase):
    def test_project_environment_overrides_global_brain_home(self) -> None:
        env = project_environment()
        self.assertEqual(env["GBRAIN_HOME"], str(GBRAIN_HOME))
        self.assertEqual(
            GBRAIN_HOME,
            GBRAIN_HOME.parents[1] / ".sdd" / "gbrain-home",
        )


if __name__ == "__main__":
    unittest.main()
