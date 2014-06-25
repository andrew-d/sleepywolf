package sleepywolf

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/zenazn/goji/web"
)

type DefaultResource struct {
}

type wrapperFuncs struct {
	BeforeOne  interface{}
	BeforeMany interface{}
	BeforeAll  interface{}
}

type funcType int

const (
	funcTypeOne  funcType = 0
	funcTypeMany funcType = 1
)

func (w *wrapperFuncs) WrapHandler(f interface{}, ty funcType) interface{} {
	// We need to wrap from most specific to least specific.
	if funcTypeOne == ty && w.BeforeOne != nil {
		log.Println("wrapping one")
	} else if funcTypeMany == ty && w.BeforeMany != nil {
		log.Println("wrapping many")
	}

	if w.BeforeAll != nil {
		log.Println("wrapping all")
	}

	return f
}

// Checks whether the given function is a valid handler function for Goji.
// Will return nil if it is, otherwise an error specifying why not.  In order
// to test member functions, pass "true" as the second argument.
func CheckValidHandler(f interface{}, skipReceiver bool) error {
	ty := reflect.TypeOf(f)

	// The function should be a function...
	if ty.Kind() != reflect.Func {
		return fmt.Errorf("not a function")
	}

	// ... should return nothing ...
	if ty.NumOut() != 0 {
		return fmt.Errorf("should return nothing")
	}

	// ... and be of the form:
	//     func(c web.C, w http.ResponseWriter, r *http.Request)
	// or
	//     func(w http.ResponseWriter, r *http.Request)
	idx := 0
	numParams := ty.NumIn()

	// If this a method on a type (i.e. func (f Foo) DoThing(...)), then the
	// first param is the receiver, and we ignore it.
	if skipReceiver {
		idx += 1
		numParams -= 1
	}

	if numParams == 3 {
		if ty.In(idx) != reflect.TypeOf(web.C{}) {
			return fmt.Errorf("param 1 (for 3-argument function) should be web.C, not %s", ty.In(idx).String())
		}
		idx += 1
	} else if numParams != 2 {
		// Wrong # of parameters
		return fmt.Errorf("wrong number of parameters: %d", numParams)
	}

	if ty.In(idx+0) != reflect.TypeOf((*http.ResponseWriter)(nil)).Elem() {
		return fmt.Errorf("param %d should be http.ResponseWriter, not %s", idx+1, ty.In(idx+0).String())
	}
	if ty.In(idx+1) != reflect.TypeOf(&http.Request{}) {
		return fmt.Errorf("param %d should be *http.Request, not %s", idx+1+1, ty.In(idx+1).String())
	}

	return nil
}

func RegisterOn(mux *web.Mux, resource interface{}) (err error) {
	ty := reflect.TypeOf(resource)

	// Get the "before" functions.
	wrap := &wrapperFuncs{}

	var ok bool
	if wrap.BeforeOne, ok = ty.MethodByName("BeforeOne"); ok {
		err = CheckValidHandler(wrap.BeforeOne, true)
		if err != nil {
			return
		}
	} else {
		wrap.BeforeOne = nil
	}
	if wrap.BeforeMany, ok = ty.MethodByName("BeforeMany"); ok {
		err = CheckValidHandler(wrap.BeforeMany, true)
		if err != nil {
			return
		}
	} else {
		wrap.BeforeMany = nil
	}
	if wrap.BeforeAll, ok = ty.MethodByName("Before"); ok {
		err = CheckValidHandler(wrap.BeforeAll, true)
		if err != nil {
			return
		}
	} else {
		wrap.BeforeAll = nil
	}

	// For each of the given methods, if they exist, we register them.
	if m, ok := ty.MethodByName("GetMany"); ok {
		log.Println("found GetMany", m)
	}

	return nil
}
