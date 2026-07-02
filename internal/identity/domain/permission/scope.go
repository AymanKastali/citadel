package permission

import (
	"regexp"
	"strings"
)

// maxScopeLength is the upper bound of a valid scope: comfortably fits any
// realistic resource:action (or deeper) scope while rejecting junk.
const maxScopeLength = 100

// scopePattern matches one or more colon-separated segments, each a non-empty run
// of lowercase alphanumerics, with at least two segments (at least one colon) — so
// a scope is always a resource:action pair (e.g. "users:read"), and deeper scopes
// like "orgs:members:invite" are allowed. Compiled once at package init.
var scopePattern = regexp.MustCompile(`^[a-z0-9]+(:[a-z0-9]+)+$`)

// Scope is the permission string itself as a value object: immutable and
// self-validating. It is the resource/action pair a role is authorized for. It is
// trimmed and lower-cased, then required to match the scope format, so equality and
// per-role uniqueness are case-insensitive and stable.
type Scope struct {
	value string
}

// NewScope is the only way to build a valid Scope.
func NewScope(raw string) (Scope, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	scopeIsMissing := normalized == ""
	if scopeIsMissing {
		return Scope{}, NewEmptyScopeError()
	}
	scopeIsTooLong := len(normalized) > maxScopeLength
	if scopeIsTooLong {
		return Scope{}, NewScopeTooLongError(len(normalized))
	}
	scopeIsMalformed := !scopePattern.MatchString(normalized)
	if scopeIsMalformed {
		return Scope{}, NewMalformedScopeError(normalized)
	}
	return Scope{value: normalized}, nil
}

func (scope Scope) Value() string { return scope.value }

func (scope Scope) Equal(other Scope) bool { return scope.value == other.value }

func (scope Scope) IsZero() bool { return scope == Scope{} }
