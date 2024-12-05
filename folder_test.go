package fsdt

// let's ensure we handle images, text encodings, binary etc.
// let's ensure we handle symlinks, hardlinks and stuff
import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/assert"
)

func JSON[T any](value T) string {
	s, _ := json.MarshalIndent(value, "", "\t")
	s = append(s, '\n')
	return string(s)
}

// TODO: these are handy, but should likely be part of some nice
// file/folder/content based assertion helpers
func dir(root string) []string {
	directory, err := os.ReadDir(root)
	if err != nil {
		log.Panic(err)
	}
	names := []string{}

	for _, entry := range directory {
		names = append(names, entry.Name())
	}
	return names
}

func read(root string) []byte {
	content, err := os.ReadFile(root)
	if err != nil {
		log.Panic(err)
	}

	return content
}

func readString(root string) string {
	return string(read(root))
}

func TestFolderMode(t *testing.T) {
	assert := assert.New(t)

	folder := NewFolder()

	assert.Equal(folder.Mode(), os.FileMode(os.ModeDir|0755))
	assert.Equal(folder.Mode().Perm(), os.FileMode(0755))
	tempdir := t.TempDir()

	folder.WriteTo(tempdir + "/foo/")

	stat, err := os.Stat(tempdir + "/foo")
	if err != nil {
		panic(err)
	}

	assert.Equal(os.FileMode(os.ModeDir|0755), stat.Mode())
	assert.Equal(os.FileMode(0755), stat.Mode().Perm())
}

func TestFolderStrings(t *testing.T) {
	assert := assert.New(t)
	folder := NewFolder()
	folder.FileString("README.md", "## HI\n")
	folder.Folder("lib", func(f *Folder) {
		f.Folder("foo", func(f *Folder) {
			f.Folder("bar", func(f *Folder) {
				f.FileString("baz.go", "package baz\n")
			})
		})
		f.FileString("lib.go", "package lib\n")
	})

	assert.Equal([]string{"README.md", "lib/", "lib/foo/", "lib/foo/bar/", "lib/foo/bar/baz.go", "lib/lib.go"}, folder.Strings(""))
	assert.Equal([]string{"root/", "root/README.md", "root/lib/", "root/lib/foo/", "root/lib/foo/bar/", "root/lib/foo/bar/baz.go", "root/lib/lib.go"}, folder.Strings("root"))
	assert.Equal([]string{"/root/", "/root/README.md", "/root/lib/", "/root/lib/foo/", "/root/lib/foo/bar/", "/root/lib/foo/bar/baz.go", "/root/lib/lib.go"}, folder.Strings("/root"))
}

func TestFolderToDirAndBack(t *testing.T) {
	folder := NewFolder()
	folder.FileString("README.md", "## HI\n")
	folder.Folder("lib", func(f *Folder) {
		f.Folder("foo", func(f *Folder) {
			f.Folder("bar", func(f *Folder) {
				f.FileString("baz.go", "package baz\n")
			})
		})
		f.FileString("lib.go", "package lib\n")
	})

	tempDir := t.TempDir()
	folderRoot := filepath.Join(tempDir, "folder")
	assert.Equal(t, []string{}, dir(tempDir), "has no files")

	err := folder.WriteTo(folderRoot)
	if err != nil {
		log.Panic(err)
	}

	t.Run("check the raw disk", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal([]string{"folder"}, dir(tempDir), "has files")
		assert.Equal([]string{"README.md", "lib"}, dir(folderRoot))
		assert.Equal([]string{"foo", "lib.go"}, dir(folderRoot+"/lib"))
		assert.Equal([]string{"bar"}, dir(folderRoot+"/lib/foo"))
		assert.Equal([]string{"baz.go"}, dir(folderRoot+"/lib/foo/bar"))

		assert.Equal("## HI\n", readString(folderRoot+"/README.md"))
		assert.Equal("package lib\n", readString(folderRoot+"/lib/lib.go"))
		assert.Equal("package baz\n", readString(folderRoot+"/lib/foo/bar/baz.go"))
	})

	t.Run("check if we can populate a folder from disk", func(t *testing.T) {
		assert := assert.New(t)
		newFolder := NewFolder()
		assert.Equal([]string{}, newFolder.Entries(), "has no files")
		newFolder.ReadFrom(folderRoot)
		assert.Equal([]string{"README.md", "lib"}, newFolder.Entries())

		assert.Equal(folder.Get("README.md").ContentString(), "## HI\n")
	})
}

func TestBasicFolderStuff(t *testing.T) {
	folder := NewFolder()
	readme := folder.FileString("README.md", "## HI\n")
	lib := folder.Folder("lib", func(f *Folder) {
		f.Folder("foo", func(f *Folder) {
			f.Folder("bar", func(f *Folder) {
				f.FileString("baz.go", "package bar\n")
			})
		})
		f.FileString("lib.go", "package lib\n")
	})

	foo := lib.Get("foo").(*Folder)
	bar := foo.Get("bar").(*Folder)

	t.Run("folder scenarios", func(t *testing.T) {
		assert := assert.New(t)
		folderScenarios := []struct {
			description string
			expected    []string
			actual      []string
		}{
			{"folder.Entries()", []string{"README.md", "lib"}, folder.Entries()},
			{"lib.Entries()", []string{"foo", "lib.go"}, lib.Entries()},
			{"foo.Entries()", []string{"bar"}, foo.Entries()},
			{"bar.Entries()", []string{"baz.go"}, bar.Entries()},
		}
		for _, s := range folderScenarios {
			assert.Equal(s.expected, s.actual, s.description)
		}
	})

	t.Run("file Scenarios", func(t *testing.T) {
		assert := assert.New(t)
		baz := bar.Get("baz.go").(*File)
		fileScenerios := []struct {
			description string
			expected    string
			actual      string
		}{
			{description: "baz.go", expected: "package bar\n", actual: string(baz.Content())},
			{description: "README.md", expected: "## HI\n", actual: string(readme.Content())},
		}
		for _, s := range fileScenerios {
			assert.Equal(s.expected, s.actual, s.description)
		}
	})
}

func BenchmarkNewFolder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFolder()
	}
}

func BenchmarkNewFileString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFileString("some content")
	}
}

func BenchmarkFolderStructure(b *testing.B) {
	folder := NewFolder()

	for i := 0; i < b.N; i++ {
		readme := folder.FileString("README.md", "## HI\n")
		lib := folder.Folder("lib", func(f *Folder) {
			f.Folder("foo", func(f *Folder) {
				f.Folder("bar", func(f *Folder) {
					f.FileString("baz.go", "package bar\n")
				})
			})
			f.FileString("lib.go", "package lib\n")
		})
		_ = readme
		_ = lib
	}
}

func BenchmarkFolderEntries(b *testing.B) {
	folder := NewFolder()
	folder.FileString("README.md", "## HI\n")
	folder.Folder("lib", func(f *Folder) {
		f.Folder("foo", func(f *Folder) {
			f.Folder("bar", func(f *Folder) {
				f.FileString("baz.go", "package bar\n")
			})
		})
		f.FileString("lib.go", "package lib\n")
	})

	for i := 0; i < b.N; i++ {
		_ = folder.Entries()
	}
}

func TestFolderEntryType(t *testing.T) {
	assert := assert.New(t)
	a := NewFolder()
	b := NewFile()

	assert.Equal(FOLDER, a.Type())
	assert.Equal(FILE, b.Type())
}

func TestCreateChildOperation(t *testing.T) {
	assert := assert.New(t)
	a := NewFolder()
	a.FileString("README.md", "## HI\n")
	a.Folder("a", func(f *Folder) {
		f.Folder("b", func(f *Folder) {
			f.FileString("c", "## HI\n")
		})
	})

	assert.Equal(op.NewFileOperation("README.md"), a.CreateChildOperation("README.md", op.Reason{}))
	assert.Equal(op.NewMkdirOperation("a",
		op.NewMkdirOperation("b",
			op.NewFileOperation("c"),
		),
	), a.CreateChildOperation("a", op.Reason{}))
}

func TestRemoveChildOperation(t *testing.T) {
	assert := assert.New(t)
	a := NewFolder()
	a.FileString("README.md", "## HI\n")
	a.Folder("a", func(f *Folder) {
		f.Folder("b", func(f *Folder) {
			f.FileString("c", "## HI\n")
		})
	})

	assert.Equal(op.NewUnlink("README.md"), a.RemoveChildOperation("README.md", op.Reason{}))
	assert.Equal(op.NewRmdir("a",
		op.NewRmdir("b",
			op.NewUnlink("c"),
		),
	), a.RemoveChildOperation("a", op.Reason{}))
}

func TestReadmeExample(t *testing.T) {
	assert := assert.New(t)
	tempDir := t.TempDir()

	// create a new folder, in memory
	folder := NewFolder()

	// add a file, with the content of a string
	readme_one := folder.FileString("README_one.md", "## h1\n")

	// add a file, with the content of a []byte{} and other more advanced options
	readme_two := folder.File("README_two.md", FileOptions{
		Content: []byte("## h1 \n"),
	})

	// create lib/Empty.txt
	lib := folder.Folder("lib", func(lib *Folder) {
		// create another file, this time with no content
		lib.File("Empty.txt")
	})

	// empty folder
	tests := folder.Folder("tests")

	_, _, _, _ = readme_one, readme_two, lib, tests

	// write that folder to disk
	err := folder.WriteTo(tempDir + "/folder/")
	if err != nil {
		panic(err)
	}

	// now /path/to/somewhere has the contents of folder

	// let's create a second folder
	newFolder := NewFolder()

	// populate it from the path we just wrote
	err = newFolder.ReadFrom(tempDir + "/folder/")
	if err != nil {
		panic(err)
	}

	// comparision to self should yield nothing
	assert.Equal([]op.Operation{}, folder.Diff(folder))

	// let's compare the two, but they are the same so it's boring
	assert.ElementsMatch([]op.Operation{}, newFolder.Diff(folder))

	// // we can also create a new folder via clone
	clone := folder.Clone().(*Folder)

	// let's compare the two, but they are the same so it's boring
	assert.Equal([]op.Operation{}, folder.Diff(clone))
	assert.Equal([]op.Operation{}, newFolder.Diff(clone))

	// let's remove a folder that contains a file
	err = folder.Remove("lib")
	if err != nil {
		panic(err)
	}
	// now there should be a diff
	assert.Equal([]op.Operation{
		op.NewMkdirOperation("lib",
			op.NewFileOperation("Empty.txt"),
		),
	}, folder.Diff(clone))

	// now there should be a diff (Reverse)
	assert.Equal([]op.Operation{
		op.NewRmdir("lib",
			op.NewUnlink("Empty.txt"),
		),
	}, clone.Diff(folder))
}
