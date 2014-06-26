package common

type HandlerInfo struct {
	Name   string
	Params int
}

type StructInfo struct {
	StructName    string
	Handlers      []HandlerInfo
	HasBeforeOne  bool
	HasBeforeMany bool
	HasBeforeAll  bool
	Warnings      []string
}
