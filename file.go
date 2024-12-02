package fsdt

import (
	"bytes"
	"os"

	op "github.com/stefanpenner/go-fsdt/operation"
)

type File struct {
	// TODO: FileInfo might be handy
	content []byte
	mode    os.FileMode
}

type FileOptions struct {
	Content []byte
	Mode    os.FileMode
}

var DEFAULT_FILE_MODE = os.FileMode(0644)

func NewFileString(content string) *File {
	return NewFile(FileOptions{Content: []byte(content)})
}

func NewFile(content ...FileOptions) *File {
	if len(content) == 0 {
		return &File{
			content: []byte{},
			mode:    DEFAULT_FILE_MODE,
		}
	}

	mode := content[0].Mode
	if mode <= 0 {
		mode = DEFAULT_FILE_MODE
	}
	return &File{
		content: content[0].Content,
		mode:    mode,
	}
}

func (f *File) Strings(prefix string) []string {
	return []string{prefix}
}

func (f *File) Clone() FolderEntry {
	return &File{
		content: append([]byte(nil), f.content...),
		mode:    f.mode,
	}
}

func (f *File) RemoveOperation(relativePath string) op.Operation {
	return op.NewUnlink(relativePath)
}

func (f *File) CreateOperation(relativePath string) op.Operation {
	return op.NewFileOperation(relativePath)
}

func (f *File) CreateChangeOperation(relativePath string, reason op.Reason) op.Operation {
	return op.FileChangedOperation(relativePath, reason)
}

func (f *File) Type() FolderEntryType {
	return FILE
}

func (f *File) WriteTo(location string) error {
	file, err := os.OpenFile(location, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.mode)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(f.content)
	return err
}

func (f *File) Content() []byte {
	return f.content
}

func (f *File) Mode() os.FileMode {
	return f.mode
}

func (f *File) ContentString() string {
	return string(f.content)
}

func (f *File) Equal(entry FolderEntry) bool {
	equal, _ := f.EqualWithReason(entry)
	return equal
}

func (f *File) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	file, isFile := entry.(*File)

	if isFile {
		if f.mode != file.mode {
			return false, op.Reason{
				Type:   op.ModeChanged,
				Before: f.Mode(),
				After:  file.Mode(),
			}
		}

		if bytes.Equal(f.content, file.content) {
			return true, op.Reason{}
		} else {
			// TODO: maybe should show offset and first char difference
			return false, op.Reason{
				Type:   op.ContentChanged,
				Before: f.content,
				After:  file.content,
			}
		}
	}
	return false, op.Reason{
		Type:   op.TypeChanged,
		Before: f.Type(),
		After:  entry.Type(),
	}
}

func (f *File) HasContent() bool {
	return true
}
