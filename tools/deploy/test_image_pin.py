"""image_pin 의 판단을 고정한다.

이 테스트가 지키는 것은 하나다: 빌드가 직전 이미지를 이름 없는 상태로 만드는
경우를 빌드 *전에* 알아채는 것. 2026-08-11·08-12·08-13 세 번 같은 방식으로
이미지를 잃었고, 세 번 다 사람이 문서의 한 줄을 기억하지 못해서였다.
"""

import unittest

from image_pin import PinNameError, orphan_verdict, pin_name


class OrphanVerdict(unittest.TestCase):
    """지금 tossos:local 인 이미지가 이 빌드에 이름을 잃는가."""

    def test_local_only_would_be_orphaned(self):
        # 이것이 2026-08-13 11:03 직전의 상태다 — 새 이미지가 local 만 달고 있었다.
        v = orphan_verdict(["tossos:local"])
        self.assertTrue(v.would_orphan)
        self.assertIn("tossos:local", v.reason)

    def test_a_durable_tag_makes_the_build_safe(self):
        # 2026-08-12 배포 뒤의 상태 — local 과 a101 핀을 함께 달고 있었다.
        v = orphan_verdict(["tossos:a101-2202ed9a", "tossos:local"])
        self.assertFalse(v.would_orphan)

    def test_no_image_yet_is_not_an_orphan(self):
        # 첫 빌드에는 잃을 것이 없다.
        self.assertFalse(orphan_verdict([]).would_orphan)

    def test_a_dangling_image_is_already_orphaned(self):
        # local 태그조차 없는 이미지는 이미 이름이 없다. 빌드가 새로 만드는
        # 손해는 아니지만 사실대로 보고한다.
        v = orphan_verdict(["<none>:<none>"])
        self.assertTrue(v.would_orphan)

    def test_a_foreign_tag_counts_as_durable(self):
        # 이름이 tossos: 로 시작하지 않아도 이름은 이름이다. 되찾을 수 있으면 된다.
        self.assertFalse(orphan_verdict(["registry.example/tossos:keep"]).would_orphan)


class PinName(unittest.TestCase):
    """핀 이름은 무엇을 되돌리는지 말해야 한다."""

    def test_change_and_commit(self):
        self.assertEqual(pin_name("a103", "ee8eca83"), "tossos:a103-ee8eca83")

    def test_change_is_required(self):
        # CHANGE 없이 만든 핀은 무엇을 되돌리는 것인지 말하지 않는다.
        with self.assertRaises(PinNameError):
            pin_name("", "ee8eca83")

    def test_commit_is_required(self):
        with self.assertRaises(PinNameError):
            pin_name("a103", "")

    def test_unknown_commit_is_refused(self):
        # Makefile 의 COMMIT 은 git 이 없으면 "unknown" 으로 떨어진다. 그 값으로
        # 만든 핀은 두 빌드가 같은 이름을 갖게 한다 — 핀이 아니라 덮어쓰기다.
        with self.assertRaises(PinNameError):
            pin_name("a103", "unknown")

    def test_docker_tag_charset_is_enforced(self):
        # 태그에 못 쓰는 문자가 change id 로 들어오면 docker 가 거부하기 전에
        # 우리가 거부한다 — 실패 지점이 빌드 뒤가 아니라 앞이어야 한다.
        with self.assertRaises(PinNameError):
            pin_name("a103 boot/starve", "ee8eca83")


if __name__ == "__main__":
    unittest.main()
