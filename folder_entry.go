package fsdt

import op "github.com/stefanpenner/go-fsdt/operation"

// Define a custom type for FolderEntryType
type FolderEntryType string

// Define constants for FolderEntryType
const (
	FOLDER   FolderEntryType = "folder"
	FILE     FolderEntryType = "file"
	SYMLINK  FolderEntryType = "symlink"
	HARDLINK FolderEntryType = "hardlink"
)

type FolderEntry interface {
	WriteTo(location string) error
	Clone() FolderEntry
	Strings(prefix string) []string
	// TODO: consider all operations taking optional reason
	RemoveOperation(relativePath string, reason op.Reason) op.Operation
	CreateOperation(relativePath string, reason op.Reason) op.Operation
	ChangeOperation(relativePath string, reason op.Reason, operations ...op.Operation) op.Operation
	Type() FolderEntryType
	Equal(FolderEntry) bool
	EqualWithReason(FolderEntry) (bool, op.Reason)
	HasContent() bool
	Content() []byte
	ContentString() string
}
