package common

type FuncInfo struct {
	Name   string
	Params int
}

type StructInfo struct {
	StructName string
	Handlers   []FuncInfo
	BeforeOne  *FuncInfo
	BeforeMany *FuncInfo
	BeforeAll  *FuncInfo
	Warnings   []string
}
