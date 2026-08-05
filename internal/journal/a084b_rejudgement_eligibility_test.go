package journal

// a084 개정 2 · D7 — 재판정 자격은 사유와 방향을 모두 본다.
//
// 개정 1은 자격을 "각인이 현재 개정과 다르다"로만 정의했다. 독립 리뷰가 그 정의의
// 두 구멍을 열었다 (issues.md B3·B4):
//
//   - 방향: 개정이 *낮아져도* 재판정한다. 더 새로운 선택기가 거부한 판단을 더 오래된
//     바이너리가 뒤집고, 거부하면 자기 개정으로 다시 쓴다. 이 배포는 프로세스가
//     셋이므로 개정이 섞이면 주기당 한 행씩 핑퐁한다.
//   - 사유: NeedsReJudgement는 *복구 선택기*에 관한 사실인데 모든 격리 사유에
//     적용됐다. stored_snapshot_corrupt는 SelectRecoverySnapshot이 만들지 않는데,
//     그 행이 "the generation was re-judged"라는 근거와 함께 SELECTOR_REVISED로
//     닫히고 다음 성공 판정에 자동 해제된다 — a084 이전에는 운영자 전용이었다.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// TestARollbackDoesNotLetAnOlderSelectorOverturnANewerRefusal: 개정 3이 각인한
// 거부를 개정 2 바이너리가 재판정하면, 더 새로운 선택기가 위험하다고 판단한 후보쌍을
// 더 오래된 선택기가 받아들여 격리를 푼다. §0.9의 반대 방향이다.
func TestARollbackDoesNotLetAnOlderSelectorOverturnANewerRefusal(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	// A newer binary owned this row; the operator then rolled back to this build.
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_snapshot_quarantines SET selector_revision=? WHERE position_id=?`,
		exitpolicy.RecoverySelectorRevision+1, position.ID); err != nil {
		t.Fatalf("seeding a newer revision: %v", err)
	}

	active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
	if err != nil || !ok {
		t.Fatalf("ActiveExitSnapshotQuarantine: %v ok=%v", err, ok)
	}
	if active.NeedsReJudgement() {
		t.Fatal("a rolled-back build re-judged a quarantine a newer selector wrote: " +
			"the older selector accepts what the newer one refused, releases the row, " +
			"and rolling forward re-quarantines it — the two builds ping-pong the same position")
	}
}

// TestAnUnknownRevisionStillEarnsItsOneReJudgement: NULL은 0으로 읽히고 0은 현재보다
// 낮으므로, 방향 규칙이 backfill 없는 재판정 의도를 깨지 않아야 한다. 이 저장소에
// 살아 있는 세 행이 정확히 이 모양이다.
func TestAnUnknownRevisionStillEarnsItsOneReJudgement(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_snapshot_quarantines SET selector_revision=NULL WHERE position_id=?`,
		position.ID); err != nil {
		t.Fatalf("clearing the stamp: %v", err)
	}

	active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
	if err != nil || !ok {
		t.Fatalf("ActiveExitSnapshotQuarantine: %v ok=%v", err, ok)
	}
	if !active.NeedsReJudgement() {
		t.Fatal("a row written before the stamp existed must still earn its one re-judgement; " +
			"the three live rows a084 exists for are exactly this shape")
	}
}

// TestAQuarantineTheSelectorNeverDecidedIsNotReJudged: stored_snapshot_corrupt는
// 복구 선택기가 만들지 않는다. 선택기 개정이 바뀌었다는 사실은 그 행에 대해 아무것도
// 말하지 않는다.
func TestAQuarantineTheSelectorNeverDecidedIsNotReJudged(t *testing.T) {
	for _, reason := range []string{"stored_snapshot_corrupt", "legacy_policy_identity_unknown"} {
		t.Run(reason, func(t *testing.T) {
			j := exitFixture(t)
			ctx := context.Background()
			o, _ := openedPosition(t, j, "10")
			position := currentPosition(t, j, o)

			if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
				reason, "the stored snapshot could not be read"); err != nil {
				t.Fatalf("QuarantineExitSnapshot: %v", err)
			}
			// The v29 database leaves every pre-existing row unstamped.
			if _, err := j.db.ExecContext(ctx,
				`UPDATE exit_snapshot_quarantines SET selector_revision=NULL WHERE position_id=?`,
				position.ID); err != nil {
				t.Fatalf("clearing the stamp: %v", err)
			}

			active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
			if err != nil || !ok {
				t.Fatalf("ActiveExitSnapshotQuarantine: %v ok=%v", err, ok)
			}
			if active.NeedsReJudgement() {
				t.Fatalf("a %s quarantine was made eligible for re-judgement; "+
					"SelectRecoverySnapshot never decided it, so a selector revision says "+
					"nothing about it — and letting it through auto-releases on the next "+
					"successful judgement a row that used to need an operator", reason)
			}
		})
	}
}

// TestAReQuarantineDoesNotClaimAReJudgementThatDidNotHappen: 사유 게이트가 없으면
// 원장이 일어나지 않은 재판정을 기록한다. 원장의 해제 기록은 운영자가 "보호가 왜
// 재개됐는가"를 재구성하는 유일한 산출물이다.
func TestAReQuarantineDoesNotClaimAReJudgementThatDidNotHappen(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"stored_snapshot_corrupt", "the stored snapshot could not be read"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_snapshot_quarantines SET selector_revision=NULL WHERE position_id=?`,
		position.ID); err != nil {
		t.Fatalf("clearing the stamp: %v", err)
	}

	// The corruption is still there on the next cycle, so the same call repeats.
	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"stored_snapshot_corrupt", "the stored snapshot could not be read"); err != nil {
		t.Fatalf("second QuarantineExitSnapshot: %v", err)
	}

	var revised int
	if err := j.db.QueryRowContext(ctx,
		`SELECT count(*) FROM exit_snapshot_quarantines WHERE position_id=? AND release_kind=?`,
		position.ID, QuarantineReleaseSelectorRevised).Scan(&revised); err != nil {
		t.Fatalf("counting releases: %v", err)
	}
	if revised != 0 {
		t.Fatalf("%d row(s) closed on SELECTOR_REVISED evidence, want 0: the recovery "+
			"selector never ran for a stored_snapshot_corrupt row, so the ledger is "+
			"asserting a re-judgement that did not happen", revised)
	}
}

// TestTheOperatorReleasePathCannotStampMachineEvidence: SELECTOR_REVISED는 판정
// 트랜잭션 안에서만 쓰인다. 외부에서 공급된 kind로 도달할 수 있으면, 사람을 지명해야
// 하는 해제가 기계 결정으로 기록될 수 있다.
func TestTheOperatorReleasePathCannotStampMachineEvidence(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	err = j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, q.Version,
		QuarantineReleaseSelectorRevised, "an operator supplied the machine's own kind")
	if err == nil {
		t.Fatal("the externally-supplied release path accepted SELECTOR_REVISED; " +
			"that kind is written only from inside the judgement transaction, and " +
			"reaching it from outside lets a release that must name a person be " +
			"recorded as a machine decision")
	}
}

// TestAJudgementThatIsNotAReJudgementReleasesNothing (개정 3).
//
// 세 번째 독립 리뷰가 실행 재현으로 잡았다. 개정 2는 재판정 여부를 활성 행의 *사유*로
// 추론했고, 그래서 지금 도는 선택기가 방금 쓴 격리를 다음 성공 판정이 "the recovery
// selector changed … the re-judgement chose one verified candidate"라는 근거와 함께
// 닫았다 — 일어나지 않은 재판정이다. 운영자만 풀 수 있어야 하는 행이 기계적으로 풀린다.
//
// 추론 대신 사실로 게이트한다. 각인은 판별자가 될 수 없다 — judge가 이 트랜잭션보다
// 먼저 각인하므로 어느 쪽이든 현재 개정이 찍혀 있다.
func TestAJudgementThatIsNotAReJudgementReleasesNothing(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		QuarantineReasonAmbiguousRecovery, "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if q.SelectorRevision != exitpolicy.RecoverySelectorRevision {
		t.Fatalf("precondition: the row must carry the running selector, got %d", q.SelectorRevision)
	}

	if err := releaseReJudgedQuarantineTxForTest(ctx, j, 0, position.ID, position.InstanceSeq); err != nil {
		t.Fatalf("release helper: %v", err)
	}

	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if !active {
		t.Fatal("a judgement that was not a re-judgement closed the quarantine. It re-decided " +
			"nothing, so the release names evidence that does not exist, and a row only an " +
			"operator could lift was lifted by a machine")
	}

	// And the genuine re-judgement still releases.
	if err := releaseReJudgedQuarantineTxForTest(ctx, j, q.Version, position.ID, position.InstanceSeq); err != nil {
		t.Fatalf("release helper: %v", err)
	}
	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if active {
		t.Fatal("the re-judgement's own release stopped working")
	}
}

// TestAReJudgementReleasesOnlyTheRowThatEarnedIt (개정 4).
//
// 네 번째 독립 리뷰가 실행 재현으로 잡았다. judge()의 각인과 이 트랜잭션 사이에서
// 운영자가 각인된 행을 풀고 병행 관측이 새 행을 열 수 있다. 버전을 보지 않는 해제는
// 그 새 행 — 지금 도는 선택기가 스스로 쓴 행 — 을 SELECTOR_REVISED로 닫는다.
// 개정 3이 없앤 것과 같은 종류의 거짓 근거이고, 한 버전만큼 좁아졌을 뿐이다.
func TestAReJudgementReleasesOnlyTheRowThatEarnedIt(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	earned, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		QuarantineReasonAmbiguousRecovery, "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	// The operator lifts the row the re-judgement was earned against, and another
	// observer's failed selection opens a replacement — both inside the window
	// between the stamp and the judgement's commit.
	if err := j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, earned.Version,
		QuarantineReleaseHumanRepair, "operator repaired the stored snapshot by hand"); err != nil {
		t.Fatalf("operator release: %v", err)
	}
	replacement, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		QuarantineReasonAmbiguousRecovery, "exitpolicy: recovery candidate identity mismatch, again")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if replacement.Version == earned.Version {
		t.Fatalf("precondition: the replacement must be a new version, got %d", replacement.Version)
	}

	if err := releaseReJudgedQuarantineTxForTest(ctx, j, earned.Version, position.ID, position.InstanceSeq); err != nil {
		t.Fatalf("release helper: %v", err)
	}

	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if !active {
		t.Fatal("a re-judgement earned against an older quarantine closed the row that replaced " +
			"it. The selector running now wrote that replacement, so SELECTOR_REVISED describes " +
			"a comparison nobody ran, and the generation this selector just refused is judgeable again")
	}
}

// releaseReJudgedQuarantineTxForTest reaches the transaction-scoped helper the
// judgement path uses, which has no exported entry point by design.
func releaseReJudgedQuarantineTxForTest(ctx context.Context, j *Journal, reJudgingVersion int64,
	id string, generation int64) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := releaseReJudgedQuarantineTx(ctx, tx, reJudgingVersion, id, generation, j.nowString()); err != nil {
		return err
	}
	return tx.Commit()
}

// TestAnOperatorOnlyQuarantineSurvivesAGenuineReJudgement (개정 4).
//
// 다섯 번째 독립 리뷰가 커버리지로 잡았다. `active.Reason != ambiguous_recovery` 가드는
// map의 안전 결론이 근거로 삼는 조건인데 어떤 테스트도 그 줄을 실행하지 않았다.
//
// 이 가드는 선택기에 관한 사실이 선택기가 만들지 않은 행에 적용되지 않게 한다.
// stored_snapshot_corrupt는 선택기가 상의되기 전에 결정되므로, 선택기 개정은 그것에
// 대해 아무 말도 하지 않는다. 재판정 버전이 정확히 맞아도 이 경로는 풀 수 없다.
func TestAnOperatorOnlyQuarantineSurvivesAGenuineReJudgement(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"stored_snapshot_corrupt", "journal: the stored exit snapshot did not decode")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	// The exact version, so only the reason can refuse.
	if err := releaseReJudgedQuarantineTxForTest(ctx, j, q.Version, position.ID, position.InstanceSeq); err != nil {
		t.Fatalf("release helper: %v", err)
	}

	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if !active {
		t.Fatal("a re-judgement closed a quarantine the recovery selector never decided. " +
			"SELECTOR_REVISED describes a comparison that was never consulted for this row, " +
			"and a row only an operator could lift was lifted by a machine")
	}
}
