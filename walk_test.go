package bfwalk

import (
	"io/fs"
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
