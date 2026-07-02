package session

// maxSecretHashLength is the upper bound of a stored hash: comfortably fits any
// encoded hash (SHA-256, HMAC, argon2id) while rejecting junk.
const maxSecretHashLength = 255

// SecretHash holds the already-hashed session secret as a value object: immutable
// and self-validating. The raw secret never enters the domain — the application's
// secret-hasher (an outbound port) hashes the freshly issued secret on the way in
// and verifies a presented secret on the way out. This VO only stores and
// structurally compares the hash; it is algorithm-agnostic (the domain neither
// knows nor checks the scheme).
type SecretHash struct {
	value string
}

// NewSecretHash is the only way to build a valid SecretHash. It does not trim,
// normalize, or check the hash's format — hash encodings are exact,
// case-sensitive, and scheme-specific, and validating them would couple the domain
// to a hashing algorithm.
func NewSecretHash(raw string) (SecretHash, error) {
	hashIsMissing := raw == ""
	if hashIsMissing {
		return SecretHash{}, NewEmptySecretHashError()
	}
	hashIsTooLong := len(raw) > maxSecretHashLength
	if hashIsTooLong {
		return SecretHash{}, NewSecretHashTooLongError(len(raw))
	}
	return SecretHash{value: raw}, nil
}

func (hash SecretHash) Value() string { return hash.value }

// Equal is a plain structural comparison of two stored hashes (e.g.
// change-detection during rotation); it is NOT secret verification. Verifying a
// presented plaintext secret against the hash is the hasher port's job in the
// application, never here.
func (hash SecretHash) Equal(other SecretHash) bool { return hash.value == other.value }

func (hash SecretHash) IsZero() bool { return hash == SecretHash{} }
