package sleepywolf

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenazn/goji/web"
)

func TestCheckValidHandler(t *testing.T) {
	var err error

	err = CheckValidHandler(1234, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "not a function")
	}

	err = CheckValidHandler(func() int { return 1 }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "should return nothing")
	}

	err = CheckValidHandler(func(a int) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "wrong number of parameters: 1")
	}

	err = CheckValidHandler(func(a, b, c, d int) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "wrong number of parameters: 4")
	}

	err = CheckValidHandler(func(a, b int) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "param 1 should be http.ResponseWriter, not int")
	}

	err = CheckValidHandler(func(a http.ResponseWriter, b int) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "param 2 should be *http.Request, not int")
	}

	err = CheckValidHandler(func(a http.ResponseWriter, b http.Request) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "param 2 should be *http.Request, not http.Request")
	}

	err = CheckValidHandler(func(a int, b http.ResponseWriter, c http.Request) { return }, false)
	if assert.Error(t, err, "an error was expected") {
		assert.Equal(t, err.Error(), "param 1 (for 3-argument function) should be web.C, not int")
	}

	err = CheckValidHandler(func(a http.ResponseWriter, b *http.Request) { return }, false)
	assert.NoError(t, err)

	err = CheckValidHandler(func(a web.C, b http.ResponseWriter, c *http.Request) { return }, false)
	assert.NoError(t, err)

	// Test that the first param is skipped
	err = CheckValidHandler(func(r int, a http.ResponseWriter, b *http.Request) { return }, true)
	assert.NoError(t, err)

	err = CheckValidHandler(func(r int, a http.ResponseWriter, b *http.Request) { return }, true)
	assert.NoError(t, err)
}
