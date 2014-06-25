package main

import (
	"fmt"
	"net/http"

	"github.com/andrew-d/sleepywolf"
	"github.com/zenazn/goji/web"
)

type TodosResource struct {
}

func (t *TodosResource) GetMany(c web.C, w http.ResponseWriter, r *http.Request) {
}

func main() {
	fmt.Println("Sleepywolf example")

	mux := web.New()

	sleepywolf.RegisterOn(mux, 1234)
	sleepywolf.RegisterOn(mux, &TodosResource{})

	sleepywolf.CheckValidHandler((*TodosResource).GetMany, true)
}
