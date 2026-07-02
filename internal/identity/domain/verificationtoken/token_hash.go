package verificationtoken

// maxTokenHashLength is the upper bound of a stored hash: comfortably fits any
// encoded hash (SHA-256, HMAC, argon2id) while rejecting junk.
const maxTokenHashLength = 255

// TokenHash holds an already-hashed secret as a value object: immutable and
// self-validating. The raw token secret never enters the domain — infrastructure
// generates it, delivers it to the user, hashes it for storage, and later hashes
// the presented secret to compare. This VO only stores and structurally compares
// the hash; it is algorithm-agnostic (the domain neither knows nor checks the
// scheme).
type TokenHash struct {
	value string
}

// NewTokenHash is the only way to build a valid TokenHash. It does not trim,
// normalize, or check the hash's format — hash encodings are exact,
// case-sensitive, and scheme-specific, and validating them would couple the
// domain to a hashing algorithm.
func NewTokenHash(raw string) (TokenHash, error) {
	hashIsMissing := raw == ""
	if hashIsMissing {
		return TokenHash{}, NewEmptyTokenHashError()
	}
	hashIsTooLong := len(raw) > maxTokenHashLength
	if hashIsTooLong {
		return TokenHash{}, NewTokenHashTooLongError(len(raw))
	}
	return TokenHash{value: raw}, nil
}

func (hash TokenHash) Value() string { return hash.value }

// Equal is a plain structural comparison of two stored hashes; it is NOT token
// verification. Comparing a presented secret against the stored hash is
// infrastructure's job (GetByHash looks a token up by its already-hashed value),
// never here.
func (hash TokenHash) Equal(other TokenHash) bool { return hash.value == other.value }

func (hash TokenHash) IsZero() bool { return hash == TokenHash{} }
