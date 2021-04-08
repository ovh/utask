package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/ovh/utask/engine/functions"
	"github.com/ovh/utask/pkg/metadata"
)

type listFunctionsIn struct {
	PageSize uint64  `query:"page_size"`
	Last     *string `query:"last"`
}

// ListFunctions returns a list of available functions in simplified format (steps not included)
func ListFunctions(c *gin.Context, in *listFunctionsIn) ([]*functions.Function, error) {
	var ret = []*functions.Function{}

	in.PageSize = normalizePageSize(in.PageSize)

	list := functions.List()

	// Get the offset on which get the functions
	from := 0
	if in.Last != nil && *in.Last != "" {
		var value string
		for from, value = range list {
			if value == *in.Last {
				from++
				break
			}
		}
	}

	for i := from; i < len(list) && i < from+int(in.PageSize); i++ {
		f, _ := functions.Get(list[i])
		ret = append(ret, f)
	}

	if uint64(len(ret)) == in.PageSize {
		last := ret[len(ret)-1].Name
		c.Header(
			linkHeader,
			buildFunctionNextLink(in.PageSize, last),
		)
	}

	c.Header(pageSizeHeader, fmt.Sprintf("%v", in.PageSize))

	return ret, nil
}

type getFunctionIn struct {
	Name string `path:"name, required"`
}

// GetFunction returns the full representation of a function, steps included
func GetFunction(c *gin.Context, in *getFunctionIn) (*functions.Function, error) {
	metadata.AddActionMetadata(c, metadata.FunctionName, in.Name)

	function, exists := functions.Get(in.Name)
	if !exists {
		return nil, errors.NewNotFound(nil, "function not found")
	}
	return function, nil

}
