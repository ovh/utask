package auth

import (
	"context"

	"github.com/juju/errors"
	"github.com/ovh/configstore"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/utils"
)

// CtxKey is a custom type based on string type
// used to fix golint when IdentityProviderCtxKey
// is set
type CtxKey string

// IdentityProviderCtxKey is the key used to store/retrieve identity data from Context
const IdentityProviderCtxKey = "__identity_provider_key"

var (
	adminUsers    []string
	resolverUsers []string
)

// WithIdentity adds identity data to a context
func WithIdentity(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxKey(IdentityProviderCtxKey), id)
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
	resolverUsers = cfg.ResolverUsernames
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

// IsAllowedResolver asserts that identity data found in context represents
// an authorized resolver user, for a given template and task
func IsAllowedResolver(ctx context.Context, tt *tasktemplate.TaskTemplate, extendedResolverUsernames []string) error {
	if err := IsAdmin(ctx); err == nil {
		return nil
	}

	id := GetIdentity(ctx)
	if !utils.ListContainsString(
		append(
			append(
				tt.AllowedResolverUsernames,
				resolverUsers...),
			extendedResolverUsernames...,
		), id) && !tt.AllowAllResolverUsernames {
		return errors.Forbiddenf("User cannot resolve this task")
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

// IsGlobalResolverUser asserts that a given user is universally allowed to resolve any task
func IsGlobalResolverUser(u string) bool {
	return utils.ListContainsString(resolverUsers, u)
}

// from a list of tokens (names and/or ids), return a list of token IDs
// translation only available if the token name is found in config
func translatedTokens(tokenNames map[string]string, tks []string) []string {
	translated := make([]string, 0)
	for _, tk := range tks {
		translation, ok := tokenNames[tk]
		if ok {
			translated = append(translated, translation)
		} else {
			translated = append(translated, tk)
		}
	}
	return translated
}
