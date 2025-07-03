# bfwalk

### Breadth-First Walk

`bfwalk` is a Go package that provides a **breadth-first** variant of `filepath.WalkDir` for traversing file trees using the `fs.FS` interface.

Unlike the standard depth-first traversal, `bfwalk.WalkDir` visits all entries at the current depth level before descending into subdirectories. This can be useful when you need to process higher-level directories before their contents.

## Installation

```sh
go get github.com/eriicafes/bfwalk
```

## Usage

`bfwalk.WalkDir` behave exactly like `filepath.WalkDir` but differes only in the **order** in which it traverses the file tree.

It supports special error return values such as `fs.SkipDir` and `fs.SkipAll` to control traversal.

```go
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/eriicafes/bfwalk"
)

func main() {
	err := bfwalk.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```
