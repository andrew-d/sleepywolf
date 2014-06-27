package main

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type TodosResource struct {
	Foo string
}

func (t *TodosResource) BeforeAll(c web.C, w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("BeforeAll called")
	return true
}

func (t *TodosResource) GetMany(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetMany called")
	fmt.Fprint(w, "GetMany")
}

func (t TodosResource) GetOne(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetOne called")
	fmt.Fprint(w, "GetOne")
}


func main() {
	RegisterTodosResource(goji.DefaultMux)
	goji.Serve()
}
