package atlas

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"ariga.io/atlas/sql/migrate"
)

type EmbedDir struct {
	fs.FS
}

func NewEmbedDir(fsys fs.FS) EmbedDir {
	return EmbedDir{withComputedSum(fsys)}
}

func (e EmbedDir) WriteFile(_ string, _ []byte) error {
	return errors.New("not implemented")
}

func (e EmbedDir) Files() ([]migrate.File, error) {
	names, err := fs.Glob(e, "*.sql")
	if err != nil {
		return nil, err
	}
	// Sort files lexicographically.
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	files := make([]migrate.File, 0, len(names))
	for _, n := range names {
		b, err := fs.ReadFile(e, n)
		if err != nil {
			return nil, fmt.Errorf("sql/migrate: read file %q: %w", n, err)
		}
		files = append(files, migrate.NewLocalFile(n, b))
	}
	return files, nil
}

func (e EmbedDir) Checksum() (migrate.HashFile, error) {
	files, err := e.Files()
	if err != nil {
		return nil, err
	}
	return migrate.NewHashFile(files)
}

// withComputedSum wraps fsys so that "atlas.sum" always returns a checksum
// computed from the actual *.sql files. This keeps the directory valid even
// when multiple embed.FS are merged together (each with its own atlas.sum).
func withComputedSum(fsys fs.FS) fs.FS {
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil || len(names) == 0 {
		return fsys
	}
	sort.Strings(names)
	files := make([]migrate.File, 0, len(names))
	for _, n := range names {
		b, err := fs.ReadFile(fsys, n)
		if err != nil {
			return fsys
		}
		files = append(files, migrate.NewLocalFile(n, b))
	}
	hf, err := migrate.NewHashFile(files)
	if err != nil {
		return fsys
	}
	sumBytes, err := hf.MarshalText()
	if err != nil {
		return fsys
	}
	return &sumOverrideFS{FS: fsys, sum: sumBytes}
}

// sumOverrideFS wraps an fs.FS and overrides the "atlas.sum" file content.
type sumOverrideFS struct {
	fs.FS
	sum []byte
}

func (s *sumOverrideFS) Open(name string) (fs.File, error) {
	if name == "atlas.sum" {
		return &memFile{Reader: bytes.NewReader(s.sum), size: int64(len(s.sum))}, nil
	}
	return s.FS.Open(name)
}

// memFile is a minimal fs.File backed by in-memory bytes.
type memFile struct {
	*bytes.Reader
	size int64
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return &memFileInfo{size: f.size}, nil
}

func (f *memFile) Close() error { return nil }

type memFileInfo struct{ size int64 }

func (i *memFileInfo) Name() string       { return "atlas.sum" }
func (i *memFileInfo) Size() int64        { return i.size }
func (i *memFileInfo) Mode() fs.FileMode  { return 0o444 }
func (i *memFileInfo) ModTime() time.Time { return time.Time{} }
func (i *memFileInfo) IsDir() bool        { return false }
func (i *memFileInfo) Sys() any           { return nil }

func (s *sumOverrideFS) ReadFile(name string) ([]byte, error) {
	if name == "atlas.sum" {
		return s.sum, nil
	}
	if rf, ok := s.FS.(fs.ReadFileFS); ok {
		return rf.ReadFile(name)
	}
	return fs.ReadFile(s.FS, name)
}
