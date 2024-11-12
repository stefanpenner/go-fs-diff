package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Folder struct {
	// FileInfo might be handy
	_entries map[string]FolderEntry
	mode     os.FileMode
}

func NewFolder() *Folder {
	return &Folder{
		_entries: map[string]FolderEntry{},
		mode:     0777,
	}
}

func (f *Folder) Entries() []string {
	entries := []string{}
	for name := range f._entries {
		entries = append(entries, name)
	}
	// TODO: sort on demand, then cache until a mutation
	// hopefully a lexigraphic sort
	sort.Strings(entries)
	return entries
}

func (f *Folder) RemoveOperation(relativePath string) Operation {
	operations := []Operation{}
	for _, relativePath := range f.Entries() {
		entry := f._entries[relativePath]
		operations = append(operations, entry.RemoveOperation(relativePath))
	}
	return NewRmdir(relativePath, operations...)
}

func (f *Folder) Get(relativePath string) FolderEntry {
	entry, ok := f._entries[relativePath]
	if ok {
		return entry
	} else {
		panic(fmt.Sprintf("Entry: %s not found in: %v", relativePath, f.Entries()))
	}
}

func (f *Folder) Remove(relativePath string) error {
	_, ok := f._entries[relativePath]
	if ok {
		delete(f._entries, relativePath)
		return nil
	} else {
		return fmt.Errorf("Remove Error: %s not found in: %v", relativePath, f.Entries())
	}
}

func (f *Folder) RemoveChildOperation(relativePath string) Operation {
	return f.Get(relativePath).RemoveOperation(relativePath)
}

func (f *Folder) CreateOperation(relativePath string) Operation {
	operations := []Operation{}

	for _, relativePath := range f.Entries() {
		entry := f._entries[relativePath]
		operations = append(operations, entry.CreateOperation(relativePath))
	}

	return NewMkdir(relativePath, operations...)
}

func (f *Folder) CreateChildOperation(relativePath string) Operation {
	return f.Get(relativePath).CreateOperation(relativePath)
}

func (f *Folder) File(name string, content ...FileOptions) *File {
	file := NewFile(content...)
	f._entries[name] = file
	return file
}

func (f *Folder) FileString(name string, content string) *File {
	file := NewFile(FileOptions{
		Content: []byte(content),
		Mode:    0666,
	})
	f._entries[name] = file
	return file
}

func (f *Folder) Folder(name string, cb ...func(*Folder)) *Folder {
	folder := NewFolder()
	f._entries[name] = folder
	for _, cb := range cb {
		cb(folder)
	}
	return folder
}

func (f *Folder) Clone() FolderEntry {
	clone := NewFolder()
	for name, entry := range f._entries {
		clone._entries[name] = entry.Clone()
	}
	return clone
}

func (f *Folder) Strings(prefix string) []string {
	entries := []string{}

	if prefix != "" {
		// TODO: decide if a non-prefix empty root folder should show up as "/" or just be dropped
		entries = append(entries, prefix+string(os.PathSeparator))
	}

	for _, name := range f.Entries() {
		entry := f._entries[name]
		fullpath := filepath.Join(prefix, name)
		entries = append(entries, entry.Strings(fullpath)...)
	}
	return entries
}

func (f *Folder) WriteTo(location string) error {
	if f.mode == 0 {
		panic("OMG")
	}
	err := os.Mkdir(location, f.mode) // TODO: mode
	if err != nil {
		return err
	}
	for _, relativePath := range f.Entries() {
		err := f.Get(relativePath).WriteTo(filepath.Join(location, relativePath))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Folder) ReadFrom(path string) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		if entry.IsDir() {
			folder := f.Folder(entry.Name(), func(f *Folder) {})
			err = folder.ReadFrom(path + "/" + entry.Name())
			if err != nil {
				return err
			}
		} else {
			content, err := os.ReadFile(path + "/" + entry.Name())
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			f.File(entry.Name(), FileOptions{Content: content, Mode: info.Mode()})
		}
	}

	return nil
}

func (f *Folder) Type() FolderEntryType {
	return FOLDER
}

func (f *Folder) Equal(entry FolderEntry) bool {
	return false
}

func (f *Folder) EqualWithReason(entry FolderEntry) (bool, Reason) {
	// TODO: deal with MODE
	return false, Reason{}
}

func (f *Folder) HasContent() bool {
	return false
}

func (f *Folder) Content() []byte {
	return []byte{}
}

func (f *Folder) Diff(b *Folder) []Operation {
	return Diff(f, b, true)
}

func (f *Folder) CaseInsensitiveDiff(b *Folder) []Operation {
	return Diff(f, b, false)
}

func (f *Folder) ContentString() string {
	panic("fsdt:folder.go(Folder does not implement contentString)")
}
