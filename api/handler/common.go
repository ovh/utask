package handler

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/ovh/utask"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/utils"
)

const (
	pageSizeHeader  = "X-Paging-PageSize"
	linkHeader      = "link"
	obfuscatedValue = "**__SECRET__**"
)

func obfuscateInput(defs []input.Input, inputs map[string]interface{}) map[string]interface{} {
	for _, i := range defs {
		if i.Type == input.InputTypePassword && inputs[i.Name] != nil {
			inputs[i.Name] = obfuscatedValue
		}
	}
	return inputs
}

func deobfuscateNewInput(old, new map[string]interface{}) map[string]interface{} {
	for k, v := range new {
		if s, ok := v.(string); ok && s == obfuscatedValue {
			new[k] = old[k]
		}
	}
	return new
}

func buildTemplateNextLink(pageSize uint64, last string) string {
	values := &url.Values{}
	values.Add("page_size", strconv.FormatUint(pageSize, 10))
	values.Add("last", last)
	return buildLink("next", "/template", values.Encode())
}

func buildTaskNextLink(typ string, state, batch *string, pageSize uint64, last string) string {
	values := &url.Values{}
	values.Add("type", typ)
	if state != nil {
		values.Add("state", *state)
	}
	if batch != nil {
		values.Add("batch", *batch)
	}
	values.Add("page_size", strconv.FormatUint(pageSize, 10))
	values.Add("last", last)
	return buildLink("next", "/task", values.Encode())
}

func buildResolutionNextLink(typ string, state *string, instID *uint64, pageSize uint64, last string) string {
	values := &url.Values{}
	values.Add("type", typ)
	if state != nil {
		values.Add("state", *state)
	}
	if instID != nil {
		values.Add("instance_id", strconv.FormatUint(*instID, 10))
	}
	values.Add("page_size", strconv.FormatUint(pageSize, 10))
	values.Add("last", last)
	return buildLink("next", "/resolution", values.Encode())
}

func buildLink(label, path, query string) string {
	u := &url.URL{
		Path:     path,
		RawQuery: query,
	}
	return fmt.Sprintf("<%s>; rel=%s", u.String(), label)
}

func normalizePageSize(pageSize uint64) uint64 {
	switch {
	case pageSize == 0:
		return utask.DefaultPageSize
	case pageSize < utask.MinPageSize:
		return utask.MinPageSize
	case pageSize > utask.MaxPageSize:
		return utask.MaxPageSize
	}
	return pageSize
}

func isRequester(t *task.Task, c *gin.Context) error {
	callUser := auth.GetIdentity(c)
	if t.RequesterUsername != callUser {
		return errors.Forbiddenf("Not requester")
	}
	return nil
}

func isResolver(t *task.Task, c *gin.Context) error {
	callUser := auth.GetIdentity(c)
	if t.ResolverUsername != nil && *t.ResolverUsername != callUser {
		return errors.Forbiddenf("Not resolver")
	}
	return nil
}

func isWatcher(t *task.Task, c *gin.Context) error {
	callUser := auth.GetIdentity(c)
	if !utils.ListContainsString(t.WatcherUsernames, callUser) {
		return errors.Forbiddenf("Not watcher")
	}
	return nil
}
