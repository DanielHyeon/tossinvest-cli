package journal

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestDataDirResolutionOrder pins the precedence the order-execution spec fixes:
// $TOSSOS_DATA_DIR > $XDG_DATA_HOME/tossos > ~/.local/share/tossos. The journal
// must never land inside the repository or the config directory.
func TestDataDirResolutionOrder(t *testing.T) {
	cases := []struct {
		name    string
		dataDir string
		xdg     string
		home    string
		want    string
		wantErr error
	}{
		{
			name:    "TOSSOS_DATA_DIR wins over everything",
			dataDir: "/srv/tossos-data", xdg: "/home/u/.local/share", home: "/home/u",
			want: "/srv/tossos-data",
		},
		{
			name: "XDG_DATA_HOME gets the app subdirectory",
			xdg:  "/home/u/.local/share", home: "/home/u",
			want: "/home/u/.local/share/tossos",
		},
		{
			name: "home fallback",
			home: "/home/u",
			want: "/home/u/.local/share/tossos",
		},
		{
			name:    "surrounding whitespace is trimmed",
			dataDir: "  /srv/tossos-data  ", home: "/home/u",
			want: "/srv/tossos-data",
		},
		{
			name: "blank XDG_DATA_HOME falls through to home",
			xdg:  "   ", home: "/home/u",
			want: "/home/u/.local/share/tossos",
		},
		{
			// XDG basedir spec: a relative path in the variable is invalid and
			// must be ignored.
			name: "relative XDG_DATA_HOME is ignored",
			xdg:  "relative/share", home: "/home/u",
			want: "/home/u/.local/share/tossos",
		},
		{
			// An explicit operator override is different: a relative path there
			// is a configuration error we refuse rather than reinterpret.
			name:    "relative TOSSOS_DATA_DIR is refused",
			dataDir: "relative/data", home: "/home/u",
			wantErr: ErrRelativeDataDir,
		},
		{
			name:    "no home and no overrides fails",
			home:    "",
			wantErr: ErrNoDataDir,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvDataDir, tc.dataDir)
			t.Setenv(EnvXDGDataHome, tc.xdg)
			t.Setenv("HOME", tc.home)

			got, err := DataDir()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("DataDir() = (%q, %v), want error %v", got, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DataDir(): %v", err)
			}
			if got != tc.want {
				t.Fatalf("DataDir() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestDefaultPath verifies the journal file name inside the resolved data dir.
func TestDefaultPath(t *testing.T) {
	t.Setenv(EnvDataDir, "/srv/tossos-data")
	t.Setenv(EnvXDGDataHome, "")
	t.Setenv("HOME", "/home/u")

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath(): %v", err)
	}
	if want := filepath.Join("/srv/tossos-data", DBFileName); got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}
