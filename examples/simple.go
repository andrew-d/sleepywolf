package simple

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

type TodosResource struct {
	Foo string
}

func (t *TodosResource) BeforeAll(c web.C, w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (t *TodosResource) GetMany(c web.C, w http.ResponseWriter, r *http.Request) {
}

func (t TodosResource) GetOne(c web.C, w http.ResponseWriter, r *http.Request) {
}
