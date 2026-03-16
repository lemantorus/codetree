package model

type EntityType int

const (
	EntityFunction EntityType = iota
	EntityClass
	EntityMethod
	EntityConstant
	EntityVariable
	EntityInterface
	EntityStruct
	EntityEnum
)

func (e EntityType) String() string {
	switch e {
	case EntityFunction:
		return "func"
	case EntityClass:
		return "class"
	case EntityMethod:
		return "method"
	case EntityConstant:
		return "const"
	case EntityVariable:
		return "var"
	case EntityInterface:
		return "interface"
	case EntityStruct:
		return "struct"
	case EntityEnum:
		return "enum"
	default:
		return "unknown"
	}
}

type CodeEntity struct {
	Name      string
	Type      EntityType
	Signature string
	Docstring string
	LineStart int
	LineEnd   int
	Children  []CodeEntity
}

type FileResult struct {
	Path     string
	Entities []CodeEntity
}

type DirNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*DirNode
	Entities []CodeEntity
}
