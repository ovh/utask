package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
)

type listTemplatesIn struct {
	PageSize uint64  `query:"page_size"`
	Last     *string `query:"last"`
}

// ListTemplates returns a list of available templates in simplified format (steps not included)
func ListTemplates(c *gin.Context, in *listTemplatesIn) ([]*tasktemplate.TaskTemplate, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	in.PageSize = normalizePageSize(in.PageSize)

	tt, err := tasktemplate.ListTemplates(dbp,
		auth.IsAdmin(c) == nil, // if admin: display hidden templates
		in.PageSize, in.Last)
	if err != nil {
		return nil, err
	}

	if uint64(len(tt)) == in.PageSize {
		lastT := tt[len(tt)-1].Name
		c.Header(
			linkHeader,
			buildTemplateNextLink(in.PageSize, lastT),
		)
	}

	c.Header(pageSizeHeader, fmt.Sprintf("%v", in.PageSize))

	return tt, nil
}

type getTemplateIn struct {
	Name string `path:"name, required"`
}

// GetTemplate returns the full representation of a template, steps included
func GetTemplate(c *gin.Context, in *getTemplateIn) (*tasktemplate.TaskTemplate, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}
	return tasktemplate.LoadFromName(dbp, in.Name)

}
