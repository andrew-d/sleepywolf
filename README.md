# Sleepywolf

## What's Is It?

It's a test project that I hacked together to learn about Go code generation.
Essentially, it takes an input file with some structs in the format:

```go
type FooResource struct {
  // Fields
}

func (f *FooResource) BeforeAll(c web.C, w http.ResponseWriter, r *http.Request) bool {
  // This executes before every handler.
  return true
}

func (f *FooResource) GetMany(c web.C, w http.ResponseWriter, r *http.Request) bool {
  // This is mapped to "GET /api/foo"
}

// ...
```

And outputs a registration function that will handle registering the given
routes at appropriate URLs and calling any present "Before" functions.

The generated code assumes you're using the [goji](https://github.com/zenazn/goji)
web framework.

## What's With The Name?

A goji berry is also known as a wolfberry.  "REST" can also mean to sleep.
Together, "sleepywolf".  Yes, it's kind of stupid.

## Should I Use It?

Probably not.  It might be a useful reference for anyone wanting to do code
generation in Go, though.

## How Does It Work?

Inspired by the way [ffjson](https://github.com/pquerna/ffjson) did it, the
process is broken down into three stages:

1. sleepywolf inspects the given source file to find all structures listed.  It
   then generates a Go file that calls the "gather" subpackage with those
   structures.
2. The gather code will use runtime reflection to get the methods defined on
   the given structs.  This is written to stdout as JSON, which the main process
   reads and deserializes.
3. sleepywolf uses the information about the defined methods to generate the
   final registration code.

## License

Apache v2
