"""image_pin 의 판단을 고정한다.

이 테스트가 지키는 것은 하나다: 빌드가 롤백 대상을 없애는 경우를 빌드 *전에*
알아채는 것. 2026-08-11 · 08-12 · 08-13 세 번 같은 방식으로 이미지를 잃었고,
세 번 다 사람이 문서의 한 줄을 기억하지 못해서였다.

독립 리뷰(2026-08-13)가 초판의 구멍 셋을 찾았고 그 셋이 여기 고정돼 있다.
  - docker 를 못 본 것이 "잃을 게 없다"로 통과했다 (blind-guard)
  - 같은 커밋에서 다시 빌드하면 핀이 직전 이미지에서 *옮겨진다* (pin-overwrite)
  - dangling 판정이 docker 가 내지 않는 모양(`<none>:<none>`)을 검사했다
"""

import unittest

from image_pin import (
    ABSENT,
    PRESENT,
    UNKNOWN,
    PinNameError,
    orphan_verdict,
    pin_name,
)


class OrphanVerdict(unittest.TestCase):
    """지금 tossos:local 인 이미지가 이 빌드에 롤백 자격을 잃는가."""

    def test_local_only_would_be_orphaned(self):
        # 2026-08-13 11:03 직전의 상태 — 새 이미지가 local 만 달고 있었다.
        v = orphan_verdict(PRESENT, ["tossos:local"])
        self.assertTrue(v.refuse)
        self.assertIn("tossos:local", v.reason)

    def test_a_durable_tag_makes_the_build_safe(self):
        v = orphan_verdict(PRESENT, ["tossos:a101-2202ed9a", "tossos:local"])
        self.assertFalse(v.refuse)

    def test_no_image_yet_is_not_an_orphan(self):
        # 첫 빌드에는 잃을 것이 없다.
        self.assertFalse(orphan_verdict(ABSENT, []).refuse)

    def test_an_untagged_image_is_already_orphaned(self):
        # docker 는 태그 없는 이미지의 RepoTags 를 빈 배열로 낸다. `<none>:<none>`
        # 은 `docker images` 표의 표기이지 inspect 의 값이 아니다 — 초판 테스트는
        # docker 가 내지 않는 모양을 검사하고 있었다.
        v = orphan_verdict(PRESENT, [])
        self.assertTrue(v.refuse)
        self.assertIn("이름이 없다", v.reason)

    def test_a_foreign_tag_counts_as_durable(self):
        self.assertFalse(orphan_verdict(PRESENT, ["registry.example/tossos:keep"]).refuse)

    def test_docker_that_cannot_be_read_refuses(self):
        # 핵심. 검사가 *못 본* 것을 *안전하다* 로 보고하면 그것은 가드가 아니다.
        # 초판은 `2>/dev/null || true` 로 모든 실패를 빈 목록으로 접었고, 그래서
        # 데몬이 죽었을 때 08-11 유실 형태가 그대로 다시 열렸다.
        v = orphan_verdict(UNKNOWN, [])
        self.assertTrue(v.refuse)
        self.assertIn("볼 수 없다", v.reason)

    def test_unknown_state_ignores_any_tags_it_was_handed(self):
        # 못 본 상태에서 딸려 온 값은 증거가 아니다.
        self.assertTrue(orphan_verdict(UNKNOWN, ["tossos:a101-2202ed9a"]).refuse)

    def test_an_unrecognised_state_refuses(self):
        # 호출자가 새 상태를 추가하고 여기를 안 고치면 fail-closed 여야 한다.
        self.assertTrue(orphan_verdict("weather", []).refuse)


class PinCollision(unittest.TestCase):
    """핀 이름은 재사용되면 안 된다 — docker tag 는 이름을 *옮긴다*."""

    def test_a_taken_pin_refuses(self):
        # 같은 커밋에서 다시 빌드하면(더러운 트리, --amend, 실패 후 재시도)
        # `docker tag` 가 그 이름을 직전 이미지에서 떼어 새 이미지에 붙인다.
        # 가드가 방금 "이름이 있으니 안전하다"고 인정한 바로 그 이름을 빼앗는다.
        v = orphan_verdict(PRESENT, ["tossos:a103-ee8eca83", "tossos:local"],
                           pin_taken=True)
        self.assertTrue(v.refuse)
        self.assertIn("이미 있다", v.reason)

    def test_a_free_pin_passes(self):
        v = orphan_verdict(PRESENT, ["tossos:a101-2202ed9a", "tossos:local"],
                           pin_taken=False)
        self.assertFalse(v.refuse)


class PinName(unittest.TestCase):
    """핀 이름은 무엇을 되돌리는지 말해야 한다."""

    def test_change_and_commit(self):
        self.assertEqual(pin_name("a103", "ee8eca83"), "tossos:a103-ee8eca83")

    def test_change_is_required(self):
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
        with self.assertRaises(PinNameError):
            pin_name("a103 boot/starve", "ee8eca83")

    def test_a_shell_metacharacter_is_refused(self):
        # CHANGE 는 사람이 명령줄에 친다. 값이 셸을 빠져나가지 못하게 하는 것은
        # Makefile 쪽(환경변수 전달)이고, 여기서는 그런 값이 핀 이름이 되지
        # 않는다는 것만 고정한다.
        for bad in ['a103"; touch /tmp/x; echo "', "a103;ls", "a103$(id)", "a103`id`"]:
            with self.assertRaises(PinNameError):
                pin_name(bad, "ee8eca83")

    def test_docker_tag_length_limit(self):
        # docker 태그 성분은 128자까지다. 넘으면 빌드가 끝난 *뒤에* docker 가
        # 거부한다 — 실패 지점이 빌드 앞이어야 한다.
        with self.assertRaises(PinNameError):
            pin_name("a" * 130, "ee8eca83")

    def test_a_long_but_legal_name_passes(self):
        name = "a" * 100
        self.assertEqual(pin_name(name, "ee8eca83"), f"tossos:{name}-ee8eca83")


if __name__ == "__main__":
    unittest.main()
