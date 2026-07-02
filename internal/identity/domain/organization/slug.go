package organization

import (
	"regexp"
	"strings"
)

// maxSlugLength is the maximum length of a single DNS label, so a slug can always
// be used as a sub-domain.
const maxSlugLength = 63

// slugPattern matches lowercase alphanumeric labels joined by single hyphens, with
// no leading, trailing, or consecutive hyphens. Compiled once at package init.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Slug is an organization's stable, URL-safe identifier as a value object:
// immutable and self-validating. It is trimmed and lower-cased, then required to
// be a DNS-label-style token so it is safe in URLs and sub-domains and stable for
// equality and uniqueness. The constructor rejects anything that is not already a
// clean slug rather than silently rewriting it — producing a slug from a name is
// an application/UI concern, not the value object's.
type Slug struct {
	value string
}

func NewSlug(raw string) (Slug, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	slugIsMissing := normalized == ""
	if slugIsMissing {
		return Slug{}, NewEmptySlugError()
	}
	slugIsTooLong := len(normalized) > maxSlugLength
	if slugIsTooLong {
		return Slug{}, NewSlugTooLongError(len(normalized))
	}
	slugIsMalformed := !slugPattern.MatchString(normalized)
	if slugIsMalformed {
		return Slug{}, NewMalformedSlugError(normalized)
	}
	return Slug{value: normalized}, nil
}

func (slug Slug) Value() string { return slug.value }

func (slug Slug) Equal(other Slug) bool { return slug.value == other.value }

func (slug Slug) IsZero() bool { return slug == Slug{} }
