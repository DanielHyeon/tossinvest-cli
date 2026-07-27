package journal

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recordingProber is the injection point that makes the allowlist logic testable
// on any host: the real statfs syscall is replaced by a canned answer.
type recordingProber struct {
	info   FSInfo
	err    error
	probed []string
}

func (p *recordingProber) Probe(dir string) (FSInfo, error) {
	p.probed = append(p.probed, dir)
	return p.info, p.err
}

// TestCheckFilesystemAllowlist is the guard's decision table. Journaling local
// filesystems pass; network, union, in-memory and FUSE-backed ones fail closed,
// because the journal's durability contract (fsync means fsync) does not hold on
// them.
func TestCheckFilesystemAllowlist(t *testing.T) {
	cases := []struct {
		name    string
		info    FSInfo
		wantErr error
	}{
		{name: "ext4 by magic", info: FSInfo{Magic: MagicExt}},
		{name: "xfs by magic", info: FSInfo{Magic: MagicXFS}},
		{name: "btrfs by magic", info: FSInfo{Magic: MagicBtrfs}},
		{name: "ext4 by name", info: FSInfo{Name: "ext4"}},
		{name: "xfs by name", info: FSInfo{Name: "XFS"}},
		{name: "btrfs by name", info: FSInfo{Name: "btrfs"}},

		{name: "fuseblk (this repository's mount)", info: FSInfo{Magic: MagicFUSE}, wantErr: ErrFilesystemNotAllowed},
		{name: "tmpfs", info: FSInfo{Magic: MagicTmpfs}, wantErr: ErrFilesystemNotAllowed},
		{name: "nfs", info: FSInfo{Magic: MagicNFS}, wantErr: ErrFilesystemNotAllowed},
		{name: "smb2", info: FSInfo{Magic: MagicSMB2}, wantErr: ErrFilesystemNotAllowed},
		{name: "overlayfs", info: FSInfo{Magic: MagicOverlay}, wantErr: ErrFilesystemNotAllowed},
		{name: "exfat", info: FSInfo{Magic: MagicExFAT}, wantErr: ErrFilesystemNotAllowed},
		{name: "vfat", info: FSInfo{Magic: MagicVFAT}, wantErr: ErrFilesystemNotAllowed},
		{name: "ntfs3", info: FSInfo{Magic: MagicNTFS3}, wantErr: ErrFilesystemNotAllowed},
		{name: "known name outside the allowlist", info: FSInfo{Name: "nfs4"}, wantErr: ErrFilesystemNotAllowed},

		{name: "unrecognised magic", info: FSInfo{Magic: 0x1234567}, wantErr: ErrFilesystemUnknown},
		{name: "empty probe result", info: FSInfo{}, wantErr: ErrFilesystemUnknown},
	}

	dir := t.TempDir()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &recordingProber{info: tc.info}
			got, err := CheckFilesystem(p, dir)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("CheckFilesystem = (%+v, %v), want error %v", got, err, tc.wantErr)
				}
				// The operator has to know which filesystem was rejected.
				if !strings.Contains(err.Error(), dir) {
					t.Errorf("error must name the inspected directory: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckFilesystem: %v", err)
			}
			if got.Name == "" {
				t.Errorf("allowed filesystem must be reported by name, got %+v", got)
			}
		})
	}
}

// TestCheckFilesystemProbesNearestExistingAncestor covers first-run behaviour: the
// data directory does not exist yet, so the guard has to inspect the closest
// existing parent instead of failing on ENOENT.
func TestCheckFilesystemProbesNearestExistingAncestor(t *testing.T) {
	base := t.TempDir()
	existing := filepath.Join(base, "share")
	if err := os.MkdirAll(existing, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(existing, "tossos", "nested", "deeper")

	p := &recordingProber{info: FSInfo{Magic: MagicExt}}
	if _, err := CheckFilesystem(p, target); err != nil {
		t.Fatalf("CheckFilesystem: %v", err)
	}
	if len(p.probed) != 1 || p.probed[0] != existing {
		t.Fatalf("probed %v, want exactly [%s]", p.probed, existing)
	}
}

// TestCheckFilesystemProbeErrorFailsClosed proves an unusable probe (unsupported
// platform, permission error) refuses the journal rather than assuming the
// filesystem is fine.
func TestCheckFilesystemProbeErrorFailsClosed(t *testing.T) {
	p := &recordingProber{err: ErrProbeUnsupported}
	if _, err := CheckFilesystem(p, t.TempDir()); !errors.Is(err, ErrProbeUnsupported) {
		t.Fatalf("want ErrProbeUnsupported, got %v", err)
	}

	p2 := &recordingProber{err: errors.New("permission denied")}
	if _, err := CheckFilesystem(p2, t.TempDir()); err == nil {
		t.Fatal("probe failure must not be treated as an allowed filesystem")
	}
}

// TestFilesystemNameMapping keeps the diagnostic names stable — they end up in the
// startup refusal message an operator has to act on.
func TestFilesystemNameMapping(t *testing.T) {
	cases := map[int64]string{
		MagicExt:     "ext2/ext3/ext4",
		MagicXFS:     "xfs",
		MagicBtrfs:   "btrfs",
		MagicFUSE:    "fuse/fuseblk",
		MagicTmpfs:   "tmpfs",
		MagicNFS:     "nfs",
		MagicOverlay: "overlayfs",
		0x1234567:    "",
	}
	for magic, want := range cases {
		if got := FilesystemName(magic); got != want {
			t.Errorf("FilesystemName(%#x) = %q, want %q", magic, got, want)
		}
	}
}

// TestAllowedFilesystemsList documents the allowlist the spec fixes.
func TestAllowedFilesystemsList(t *testing.T) {
	got := AllowedFilesystems()
	want := []string{"ext4", "xfs", "btrfs"}
	if len(got) != len(want) {
		t.Fatalf("AllowedFilesystems() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllowedFilesystems() = %v, want %v", got, want)
		}
	}
}
