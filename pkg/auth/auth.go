package auth

import (
	"context"

	"github.com/juju/errors"
	"github.com/ovh/configstore"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/utils"
)

// IdentityProviderCtxKey is the key used to store/retrieve identity data from Context
const IdentityProviderCtxKey = "__identity_provider_key"

var (
	adminUsers []string
)

// WithIdentity adds identity data to a context
func WithIdentity(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, IdentityProviderCtxKey, id)
}

// Init reads authorization from configstore, bootstraps values
// used to handle authorization
func Init(store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}
	if len(cfg.AdminUsernames) < 1 {
		return errors.New("Admin user list can't be empty")
	}
	adminUsers = cfg.AdminUsernames
	return nil
}

// GetIdentity returns identity data stored in context
func GetIdentity(ctx context.Context) string {
	id := ctx.Value(IdentityProviderCtxKey)
	if id != nil {
		return id.(string)
	}
	return ""
}

// IsAdmin asserts that identity data found in context represents an admin user
func IsAdmin(ctx context.Context) error {
	id := GetIdentity(ctx)
	if !utils.ListContainsString(adminUsers, id) {
		return errors.Forbiddenf("Not an admin user")
	}
	return nil
}

// IsRequester asserts that identity data found in context represents
// the requester of the given task
func IsRequester(ctx context.Context, t *task.Task) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)
	if t.RequesterUsername != id {
		return errors.Forbiddenf("User is not requester of this task")
	}
	return nil
}

// IsWatcher asserts that identity data found in context represents
// a watcher of the given task
func IsWatcher(ctx context.Context, t *task.Task) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)
	if !utils.ListContainsString(t.WatcherUsernames, id) {
		return errors.Forbiddenf("User is not watcher of this task")
	}
	return nil
}

// IsResolver asserts that identity data found in context is the actual resolver of a given resolution
func IsResolver(ctx context.Context, r *resolution.Resolution) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)

	if id != r.ResolverUsername {
		return errors.Forbiddenf("User not authorized on this resolution")
	}

	return nil
}

// IsResolutionManager asserts that identity data found in context is either:
// - a template owner (allowed_resolver_usernames)
// - a task resolver (resolver_usernames)
// - this task resolver (resolver_username)
func IsResolutionManager(ctx context.Context, tt *tasktemplate.TaskTemplate, t *task.Task, r *resolution.Resolution) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)

	if t == nil {
		return errors.New("nil task")
	}

	if err := IsTemplateOwner(ctx, tt); err == nil {
		return nil
	}

	if utils.ListContainsString(t.ResolverUsernames, id) {
		return nil
	}

	if r != nil && r.ResolverUsername == id {
		return nil
	}

	return errors.Forbiddenf("User not authorized on this resolution")
}

// IsTemplateOwner asserts that identity data found in context is a template allowed_resolver_usernames
func IsTemplateOwner(ctx context.Context, tt *tasktemplate.TaskTemplate) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)

	if tt == nil {
		return errors.New("nil tasktemplate")
	}

	if utils.ListContainsString(tt.AllowedResolverUsernames, id) {
		return nil
	}

	return errors.Forbiddenf("User not authorized on this resolution")
}
