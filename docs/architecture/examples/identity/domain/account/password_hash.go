package account

// maxPasswordHashLength is the upper bound of a stored hash: comfortably fits any
// encoded hash (bcrypt, argon2id, scrypt — a PHC/$2b$… string) while rejecting junk.
const maxPasswordHashLength = 255

// PasswordHash holds an already-hashed password as a value object: immutable and
// self-validating. The raw plaintext never enters the domain — the application's
// password-hasher (an outbound port) hashes on the way in and verifies on the way
// out. This VO only stores and structurally compares the hash; it is
// algorithm-agnostic (the domain neither knows nor checks the scheme).
type PasswordHash struct {
	value string
}

// NewPasswordHash is the only way to build a valid PasswordHash. It does not trim,
// normalize, or check the hash's format — hash encodings are exact, case-sensitive,
// and scheme-specific, and validating them would couple the domain to a hashing
// algorithm.
func NewPasswordHash(raw string) (PasswordHash, error) {
	hashIsMissing := raw == ""
	if hashIsMissing {
		return PasswordHash{}, NewEmptyPasswordHashError()
	}
	hashIsTooLong := len(raw) > maxPasswordHashLength
	if hashIsTooLong {
		return PasswordHash{}, NewPasswordHashTooLongError(len(raw))
	}
	return PasswordHash{value: raw}, nil
}

func (hash PasswordHash) Value() string { return hash.value }

// Equal is a plain structural comparison of two stored hashes (e.g. for
// change-detection); it is NOT password verification. Verifying a plaintext
// against the hash is the hasher port's job in the application, never here.
func (hash PasswordHash) Equal(other PasswordHash) bool { return hash.value == other.value }

func (hash PasswordHash) IsZero() bool { return hash == PasswordHash{} }
