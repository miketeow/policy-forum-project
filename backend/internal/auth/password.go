package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrInvalidHash is returned when the encoded hash is not in the correct format.
var ErrInvalidHash = errors.New("the encoded hash is not in the correct format")

// Define the security parameter, which match the standard OWASP recommendations.
type argonParams struct {
	memory      uint32
	iteration   uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultParams = argonParams{
	memory:      64 * 1024, // 64 MB of RAM per hash
	iteration:   3,         // CPU time cost
	parallelism: 2,         // Number of threads
	saltLength:  16,        // 16 bytes salt
	keyLength:   32,        // 32 bytes resulting hash
}

// HashPassword generates a cryptographically secure hash Argon2ID hash in PHC string format
func HashPassword(password string) (string, error) {
	// generate a cryptographically secure random salt
	salt := make([]byte, defaultParams.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// generate the argon2id hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultParams.iteration,
		defaultParams.memory,
		defaultParams.parallelism,
		defaultParams.keyLength,
	)

	// base64 encode the salt and the hashed password
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// build the PHC format string
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultParams.memory,
		defaultParams.iteration,
		defaultParams.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil

}

// ComparePasswordAndHash extracts the parameters from the PHC string, rehashes the provided password, and securely compare the two.
func ComparePasswordAndHash(password, encodedHash string) (bool, error) {
	// split string into parts: "", "argon2id", "v=19", "m=65536", "t=3", "p=2", "salt", "hash"
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, ErrInvalidHash
	}

	// Extract the paramaters to ensure we hash the attempted password with exactly the same parameters
	var memory, iterations uint32
	var parrallelism uint8
	_, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parrallelism)
	if err != nil {
		return false, err
	}

	// decode the base64 salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	// decode the base64 salt and hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	keyLength := uint32(len(decodedHash))

	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parrallelism,
		keyLength,
	)

	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil

}
