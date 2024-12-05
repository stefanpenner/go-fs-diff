package operation

import (
	"fmt"
)

type (
	Operand string
)

const (
	Create       Operand = "Create"
	ChangeFile   Operand = "ChangeFile"
	ChangeFolder Operand = "ChangeDir"
)

const (
	Rmdir Operand = "Rmdir"
	Mkdir Operand = "Mkdir"
)

type Reason struct {
	Before interface{}
	After  interface{}
	Type   ReasonType
}

// TODDO: expand reason from enum, to struct (before / after)
type ReasonType string

const (
	TypeChanged    ReasonType = "Type Changed"
	ModeChanged    ReasonType = "Mode Changed"
	ContentChanged ReasonType = "Content Changed"
	Missing        ReasonType = "Missing"
	Because        ReasonType = "because"
)

type Operation struct {
	RelativePath string
	Value        interface{}
	Operand      Operand
}

func (op Operation) String() string {
	result := fmt.Sprintf("%s %s", op.Operand, op.RelativePath)
	// TODO: implement printing of value, which includes nesting
	return result
}
