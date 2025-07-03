package bfwalk

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
)

func TestWalkDir(t *testing.T) {
	memFS := fstest.MapFS{
		"root/file1.txt":          {Data: []byte("")},
		"root/dirA/file1.txt":     {Data: []byte("")},
		"root/dirB/file1.txt":     {Data: []byte("")},
		"root/dirB/sub/file1.txt": {Data: []byte("")},
	}

	var visited []string
	err := WalkDir(memFS, "root", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"root",
		"root/dirA",
		"root/dirB",
		"root/file1.txt",
		"root/dirA/file1.txt",
		"root/dirB/file1.txt",
		"root/dirB/sub",
		"root/dirB/sub/file1.txt",
	}
	if !slices.Equal(visited, expected) {
		t.Errorf("expected:\n  %v\ngot\n: %v", expected, visited)
	}
}

func TestWalkDirSkipAll(t *testing.T) {
	memFS := fstest.MapFS{
		"root/file1.txt":      {Data: []byte("")},
		"root/dirA/file1.txt": {Data: []byte("")},
	}

	var visited []string
	err := WalkDir(memFS, "root", func(path string, d fs.DirEntry, err error) error {
		visited = append(visited, path)
		return fs.SkipAll
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"root",
	}
	if !slices.Equal(visited, expected) {
		t.Errorf("expected:\n  %v\ngot\n: %v", expected, visited)
	}
}

func TestWalkDirSkipDir(t *testing.T) {
	memFS := fstest.MapFS{
		"root/file1.txt":          {Data: []byte("")},
		"root/dirA/file1.txt":     {Data: []byte("")},
		"root/dirB/afile1.txt":    {Data: []byte("")},
		"root/dirB/file1.txt":     {Data: []byte("")},
		"root/dirB/sub/file1.txt": {Data: []byte("")},
		"root/dirB/zfile1.txt":    {Data: []byte("")},
	}

	// skipped entire directory
	var visited []string
	err := WalkDir(memFS, "root", func(path string, d fs.DirEntry, err error) error {
		visited = append(visited, path)
		if path == "root/dirB" {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"root",
		"root/dirA",
		"root/dirB",
		"root/file1.txt",
		"root/dirA/file1.txt",
	}
	if !slices.Equal(visited, expected) {
		t.Errorf("expected:\n  %v\ngot\n: %v", expected, visited)
	}

	// skipped path appears before some
	visited = nil
	err = WalkDir(memFS, "root", func(path string, d fs.DirEntry, err error) error {
		visited = append(visited, path)
		if path == "root/dirB/afile1.txt" {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = []string{
		"root",
		"root/dirA",
		"root/dirB",
		"root/file1.txt",
		"root/dirA/file1.txt",
		"root/dirB/afile1.txt",
	}
	if !slices.Equal(visited, expected) {
		t.Errorf("expected:\n  %v\ngot\n: %v", expected, visited)
	}

	// skipped path appears after some
	visited = nil
	err = WalkDir(memFS, "root", func(path string, d fs.DirEntry, err error) error {
		visited = append(visited, path)
		if path == "root/dirB/zfile1.txt" {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = []string{
		"root",
		"root/dirA",
		"root/dirB",
		"root/file1.txt",
		"root/dirA/file1.txt",
		"root/dirB/afile1.txt",
		"root/dirB/file1.txt",
		"root/dirB/sub",
		"root/dirB/zfile1.txt",
	}
	if !slices.Equal(visited, expected) {
		t.Errorf("expected:\n  %v\ngot\n: %v", expected, visited)
	}
}

func BenchmarkWalk(b *testing.B) {
	smFsys := generateFS("data", 3, 2)
	lgFsys := generateFS("data", 100, 5)

	cases := []struct {
		name     string
		fsys     fs.FS
		walkFunc func(fsys fs.FS, root string, fn fs.WalkDirFunc) error
	}{
		{"Std Small", smFsys, fs.WalkDir},
		{"BreadthFirst Small", smFsys, WalkDir},
		{"Std Large", lgFsys, fs.WalkDir},
		{"BreadthFirst Large", lgFsys, WalkDir},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				visited := 0
				c.walkFunc(c.fsys, "data", func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					visited++
					return nil
				})
			}
		})
	}
}

func generateFS(root string, numDirs, nestingDepth int) fs.FS {
	var (
		filesPerDir = 10
		fileTypes   = []string{"html", "go", "ts", "css"}
		fsys        = fstest.MapFS{}
	)

	for i := range make([]struct{}, numDirs) {
		currentPath := root

		for d := range make([]struct{}, nestingDepth) {
			currentPath = filepath.Join(currentPath, fmt.Sprintf("dir%d_%d", i, d))

			for f := range make([]struct{}, filesPerDir) {
				ext := fileTypes[f%len(fileTypes)]
				isLayout := f%5 == 0

				filename := fmt.Sprintf("file%d.%s", f, ext)
				if isLayout {
					filename = fmt.Sprintf("layout.%s", ext)
				}

				filePath := filepath.Join(currentPath, filename)
				content := fmt.Appendf(nil, "// dummy content for %s\n", filePath)
				fsys[filePath] = &fstest.MapFile{Data: content, Mode: 0644}
			}
		}
	}
	return fsys
}
