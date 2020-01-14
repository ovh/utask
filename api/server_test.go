package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateBaseHref(t *testing.T) {
	assert.Equal(t, "/ui/dashboard/", generateBaseHref("", "/ui/dashboard/"))
	assert.Equal(t, "/ui/dashboard/", generateBaseHref("/", "/ui/dashboard/"))
	assert.Equal(t, "/ui/dashboard/", generateBaseHref("", "/ui/dashboard"))
	assert.Equal(t, "/toto/ui/dashboard/", generateBaseHref("/toto", "/ui/dashboard/"))
	assert.Equal(t, "/toto/ui/dashboard/", generateBaseHref("/toto/", "/ui/dashboard/"))
}

func Test_generateAPIPathPrefix(t *testing.T) {
	assert.Equal(t, "/", generatePathPrefixAPI(""))
	assert.Equal(t, "/", generatePathPrefixAPI("./"))
	assert.Equal(t, "/", generatePathPrefixAPI("."))
	assert.Equal(t, "/toto/", generatePathPrefixAPI("/toto"))
	assert.Equal(t, "/toto/", generatePathPrefixAPI("/toto/"))

}
