package operation

type FileChangedValue struct {
	Reason Reason
}

func NewFileOperation(relativePath string) Operation {
	return Operation{Operand: Create, RelativePath: relativePath}
}
