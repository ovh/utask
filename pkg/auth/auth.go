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

// GroupProviderCtxKey is the key used to store/retrieve group data from Context
const GroupProviderCtxKey = "__group_provider_key"

var (
	adminUsers  []string
	adminGroups []string
)

// WithIdentity adds identity data to a context
func WithIdentity(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, IdentityProviderCtxKey, id)
}

// WithIdentity adds identity data to a context
func WithGroups(ctx context.Context, groups []string) context.Context {
	return context.WithValue(ctx, GroupProviderCtxKey, groups)
}

// Init reads authorization from configstore, bootstraps values
// used to handle authorization
func Init(store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}
	if len(cfg.AdminUsernames) < 1 && len(cfg.AdminGroups) < 1 {
		return errors.New("Admin user list can't be empty")
	}
	adminUsers = cfg.AdminUsernames
	adminGroups = cfg.AdminGroups
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

// GetGroups returns group data stored in context
func GetGroups(ctx context.Context) []string {
	groups := ctx.Value(GroupProviderCtxKey)
	if groups != nil {
		return groups.([]string)
	}
	return []string{}
}

// IsAdmin asserts that identity data found in context represents an admin user
func IsAdmin(ctx context.Context) error {
	id := GetIdentity(ctx)
	if utils.ListContainsString(adminUsers, id) {
		return nil
	}

	groups := GetGroups(ctx)
	if utils.HasIntersection(adminGroups, groups) {
		return nil
	}

	return errors.Forbiddenf("Not an admin user")
}

// IsRequester asserts that identity data found in context represents
// the requester of the given task
func IsRequester(ctx context.Context, t *task.Task) error {
	id := GetIdentity(ctx)
	if t.RequesterUsername != id {
		return errors.Forbiddenf("User is not requester of this task")
	}
	return nil
}

// IsWatcher asserts that identity data found in context represents
// a watcher of the given task
func IsWatcher(ctx context.Context, t *task.Task) error {
	id := GetIdentity(ctx)
	if !utils.ListContainsString(t.WatcherUsernames, id) {
		return errors.Forbiddenf("User is not watcher of this task")
	}
	return nil
}

// IsResolutionManager asserts that identity data found in context is either:
// - a template owner (allowed_resolver_usernames or allowed_resolver_groups)
// - a task resolver (resolver_usernames or resolver_groups)
// - this task resolver (resolver_username)
func IsResolutionManager(ctx context.Context, tt *tasktemplate.TaskTemplate, t *task.Task, r *resolution.Resolution) error {
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

	groups := GetGroups(ctx)
	if utils.HasIntersection(t.ResolverGroups, groups) {
		return nil
	}

	return errors.Forbiddenf("User not authorized on this resolution")
}

// IsTemplateOwner asserts that:
// - identity data found in context is a template allowed_resolver_usernames
// - or group data found in context is a template allowed_resolver_groups
func IsTemplateOwner(ctx context.Context, tt *tasktemplate.TaskTemplate) error {
	id := GetIdentity(ctx)

	if tt == nil {
		return errors.New("nil tasktemplate")
	}

	if utils.ListContainsString(tt.AllowedResolverUsernames, id) {
		return nil
	}

	groups := GetGroups(ctx)
	if utils.HasIntersection(tt.AllowedResolverGroups, groups) {
		return nil
	}

	return errors.Forbiddenf("User not authorized on this resolution")
}
