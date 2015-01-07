package base

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	// "net/http"
	// "net/http/httptest"
	"testing"
)

func TestCacheHeaderTmpl(t *testing.T) {

	assert.Contains(t, headCache, "<title>", "It should contain head template.")

}

func TestCacheBodyTmpl(t *testing.T) {

	assert.Contains(t, bodyCache, "<body>", "It should contain body template.")

}

func TestCacheLayoutTmpl(t *testing.T) {

	assert.Contains(t, layoutCache, "</html>", "It should contain html template.")

}
