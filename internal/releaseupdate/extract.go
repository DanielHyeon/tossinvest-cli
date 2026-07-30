package releaseupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

type archiveLimits struct {
	maxArchive  int64
	maxBinary   int64
	maxExpanded int64
	maxEntries  int
}

func extractTossctl(compressed []byte, limits archiveLimits) ([]byte, error) {
	if limits.maxArchive <= 0 || limits.maxBinary <= 0 ||
		limits.maxExpanded <= 0 || limits.maxEntries <= 0 {
		return nil, errors.New("releaseupdate: invalid archive limits")
	}
	if int64(len(compressed)) > limits.maxArchive {
		return nil, fmt.Errorf("releaseupdate: archive exceeds %d bytes", limits.maxArchive)
	}

	source := bytes.NewReader(compressed)
	gz, err := gzip.NewReader(source)
	if err != nil {
		return nil, fmt.Errorf("releaseupdate: opening gzip archive: %w", err)
	}
	gz.Multistream(false)
	tr := tar.NewReader(gz)

	seen := make(map[string]struct{})
	var binary []byte
	var expanded int64
	entries := 0
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = gz.Close()
			return nil, fmt.Errorf("releaseupdate: reading tar header: %w", err)
		}
		entries++
		if entries > limits.maxEntries {
			_ = gz.Close()
			return nil, fmt.Errorf("releaseupdate: archive has more than %d entries", limits.maxEntries)
		}
		if header.Size < 0 || header.Size > limits.maxExpanded-expanded {
			_ = gz.Close()
			return nil, fmt.Errorf("releaseupdate: archive expands beyond %d bytes", limits.maxExpanded)
		}
		expanded += header.Size
		if len(header.PAXRecords) != 0 ||
			header.Format == tar.FormatPAX ||
			header.Format == tar.FormatGNU {
			_ = gz.Close()
			return nil, fmt.Errorf(
				"releaseupdate: PAX/GNU metadata is not accepted for %q", header.Name)
		}

		name, err := canonicalArchiveName(header.Name, header.Typeflag == tar.TypeDir)
		if err != nil {
			_ = gz.Close()
			return nil, err
		}
		if _, duplicate := seen[name]; duplicate {
			_ = gz.Close()
			return nil, fmt.Errorf("releaseupdate: duplicate archive entry %q", name)
		}
		seen[name] = struct{}{}

		switch header.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			switch {
			case name == "tossctl":
				if header.Mode&0o111 == 0 {
					_ = gz.Close()
					return nil, errors.New("releaseupdate: tossctl archive entry is not executable")
				}
				if header.Size > limits.maxBinary {
					_ = gz.Close()
					return nil, fmt.Errorf("releaseupdate: tossctl exceeds %d bytes", limits.maxBinary)
				}
				if binary != nil {
					_ = gz.Close()
					return nil, errors.New("releaseupdate: duplicate tossctl entry")
				}
				binary, err = readExactBounded(tr, header.Size, limits.maxBinary)
				if err != nil {
					_ = gz.Close()
					return nil, fmt.Errorf("releaseupdate: reading tossctl: %w", err)
				}
			case name == "auth-helper" || strings.HasPrefix(name, "auth-helper/"):
				if _, err := io.CopyN(io.Discard, tr, header.Size); err != nil {
					_ = gz.Close()
					return nil, fmt.Errorf("releaseupdate: consuming %q: %w", name, err)
				}
			default:
				_ = gz.Close()
				return nil, fmt.Errorf("releaseupdate: unexpected regular archive entry %q", name)
			}
		case tar.TypeDir:
			if header.Size != 0 ||
				(name != "auth-helper" && !strings.HasPrefix(name, "auth-helper/")) {
				_ = gz.Close()
				return nil, fmt.Errorf("releaseupdate: unexpected directory entry %q", name)
			}
		default:
			_ = gz.Close()
			return nil, fmt.Errorf(
				"releaseupdate: archive entry %q has forbidden type %d", name, header.Typeflag)
		}
	}
	var trailing [1]byte
	if n, err := gz.Read(trailing[:]); n != 0 || !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing uncompressed payload")
		}
		_ = gz.Close()
		return nil, fmt.Errorf(
			"releaseupdate: tar end marker is not the end of the gzip payload: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("releaseupdate: closing gzip archive: %w", err)
	}
	if source.Len() != 0 {
		return nil, errors.New("releaseupdate: concatenated gzip members or trailing bytes are forbidden")
	}
	if binary == nil {
		return nil, errors.New("releaseupdate: archive has no regular root tossctl entry")
	}
	return binary, nil
}

func canonicalArchiveName(raw string, directory bool) (string, error) {
	if strings.ContainsRune(raw, '\x00') || strings.Contains(raw, `\`) ||
		strings.HasPrefix(raw, "/") {
		return "", fmt.Errorf("releaseupdate: unsafe archive path %q", raw)
	}
	name := raw
	if directory {
		name = strings.TrimSuffix(name, "/")
	}
	if name == "" || name == "." || path.Clean(name) != name {
		return "", fmt.Errorf("releaseupdate: unsafe archive path %q", raw)
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("releaseupdate: unsafe archive path %q", raw)
		}
	}
	return name, nil
}

func readExactBounded(reader io.Reader, size, max int64) ([]byte, error) {
	if size < 0 || size > max {
		return nil, fmt.Errorf("entry size %d exceeds %d", size, max)
	}
	var out bytes.Buffer
	out.Grow(int(size))
	if _, err := io.CopyN(&out, reader, size); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
