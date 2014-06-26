package gather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/zenazn/goji/web"

	"github.com/andrew-d/sleepywolf/common"
)

type registeredStruct struct {
	Inst interface{}
	Name string
}

type InfoGatherer struct {
	registered []registeredStruct
}

func NewInfoGatherer() InfoGatherer {
	return InfoGatherer{
		registered: []registeredStruct{},
	}
}

// Checks whether the given function is a valid handler function for Goji.
// Will return nil if it is, otherwise an error specifying why not.  In order
// to test member functions, pass "true" as the second argument.
func CheckValidHandler(f interface{}, skipReceiver bool) error {
	ty := reflect.TypeOf(f)

	// The function should be a function...
	if ty.Kind() != reflect.Func {
		return fmt.Errorf("value is not a function: %s", ty.Kind().String())
	}

	// ... should return nothing ...
	if ty.NumOut() != 0 {
		return fmt.Errorf("function should have 0 return values")
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

func (i *InfoGatherer) Register(name string, s interface{}) {
	i.registered = append(i.registered, registeredStruct{
		Name: name,
		Inst: s,
	})
}

func (i *InfoGatherer) Run(w io.Writer) (err error) {
	output := []common.StructInfo{}
	checkMethods := []string{
		"Delete",
		"GetMany",
		"GetOne",
		"Patch",
		"Post",
		"Put",
	}

	for _, s := range i.registered {
		ty := reflect.TypeOf(s.Inst)
		curr := common.StructInfo{
			StructName: s.Name,
			Handlers:   []common.HandlerInfo{},
			Warnings:   []string{},
		}

		// Check for handler functions.
		for _, mname := range checkMethods {
			method, ok := ty.MethodByName(mname)
			if !ok {
				continue
			}

			miface := method.Func.Interface()
			valid := CheckValidHandler(miface, true)
			if valid != nil {
				curr.Warnings = append(curr.Warnings, fmt.Sprintf(
					"method '%s' is present but invalid: %s",
					mname, valid.Error(),
				))
				continue
			}

			// Note: the -1 is to account for the implicit "reciever" param,
			// which is the first parameter of the bare function.
			curr.Handlers = append(curr.Handlers, common.HandlerInfo{
				Name: mname,
				Params: reflect.TypeOf(miface).NumIn() - 1,
			})
		}

		// Check for 'Before' functions
		_, curr.HasBeforeOne = ty.MethodByName("BeforeOne")
		_, curr.HasBeforeMany = ty.MethodByName("BeforeMany")
		_, curr.HasBeforeAll = ty.MethodByName("BeforeAll")

		output = append(output, curr)
	}

	json.NewEncoder(w).Encode(output)
	return nil
}
