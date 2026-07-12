package identity

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Algorithm      = "argon2id"
	argon2Version        = 19
	maxPasswordBytes     = 1024
	maxArgon2MemoryKiB   = 524288 // 512 MiB
	maxArgon2Iterations  = 100
	maxArgon2Parallelism = 255
	maxSaltLength        = 64
	maxKeyLength         = 64
	minSaltLength        = 8
	minKeyLength         = 16
)

// ErrEmptyPassword indicates the plaintext password is empty.
var ErrEmptyPassword = errors.New("password must not be empty")

// ErrPasswordTooLong indicates the plaintext password exceeds the byte limit.
var ErrPasswordTooLong = errors.New("password exceeds maximum length")

// ErrInvalidArgon2Config indicates Argon2id configuration is invalid.
var ErrInvalidArgon2Config = errors.New("invalid argon2id configuration")

// Argon2idConfig holds approved Argon2id cost parameters.
// Memory is expressed in kibibytes (KiB), matching golang.org/x/crypto/argon2.
type Argon2idConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2idConfig is the production-oriented Argon2id parameter set (ADR-0004).
var DefaultArgon2idConfig = Argon2idConfig{
	Memory:      65536, // 64 MiB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// Validate checks Argon2id parameters for safe, non-zero bounds.
func (c Argon2idConfig) Validate() error {
	if c.Memory == 0 || c.Memory > maxArgon2MemoryKiB {
		return fmt.Errorf("%w: memory must be between 1 and %d KiB", ErrInvalidArgon2Config, maxArgon2MemoryKiB)
	}
	if c.Iterations == 0 || c.Iterations > maxArgon2Iterations {
		return fmt.Errorf("%w: iterations must be between 1 and %d", ErrInvalidArgon2Config, maxArgon2Iterations)
	}
	if c.Parallelism == 0 || c.Parallelism > maxArgon2Parallelism {
		return fmt.Errorf("%w: parallelism must be between 1 and %d", ErrInvalidArgon2Config, maxArgon2Parallelism)
	}
	if c.SaltLength < minSaltLength || c.SaltLength > maxSaltLength {
		return fmt.Errorf("%w: salt length must be between %d and %d bytes", ErrInvalidArgon2Config, minSaltLength, maxSaltLength)
	}
	if c.KeyLength < minKeyLength || c.KeyLength > maxKeyLength {
		return fmt.Errorf("%w: key length must be between %d and %d bytes", ErrInvalidArgon2Config, minKeyLength, maxKeyLength)
	}
	return nil
}

// PasswordHasher hashes and verifies passwords with Argon2id.
type PasswordHasher struct {
	config Argon2idConfig
}

// NewPasswordHasher validates config and returns a password hasher.
func NewPasswordHasher(config Argon2idConfig) (*PasswordHasher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &PasswordHasher{config: config}, nil
}

// HashPassword derives and encodes an Argon2id hash for plaintext.
func (h *PasswordHasher) HashPassword(ctx context.Context, plaintext string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := validatePlaintextPassword(plaintext); err != nil {
		return "", err
	}

	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(plaintext), salt, h.config.Iterations, h.config.Memory, h.config.Parallelism, h.config.KeyLength)
	return encodePasswordHash(h.config, salt, key), nil
}

// VerifyPassword checks plaintext against an encoded Argon2id hash.
// Wrong passwords and malformed encodings both yield valid=false with nil error.
func (h *PasswordHasher) VerifyPassword(ctx context.Context, plaintext, encodedHash string) (valid bool, needsRehash bool, err error) {
	if err := ctx.Err(); err != nil {
		return false, false, err
	}
	if err := validatePlaintextPassword(plaintext); err != nil {
		return false, false, err
	}

	parsed, ok := parseEncodedPasswordHash(encodedHash)
	if !ok {
		return false, false, nil
	}

	if err := ctx.Err(); err != nil {
		return false, false, err
	}

	derived := argon2.IDKey([]byte(plaintext), parsed.salt, parsed.params.Iterations, parsed.params.Memory, parsed.params.Parallelism, uint32(len(parsed.key)))
	if len(derived) != len(parsed.key) || subtle.ConstantTimeCompare(derived, parsed.key) != 1 {
		return false, false, nil
	}

	return true, h.configDiffers(parsed.params, parsed.saltLength, uint32(len(parsed.key))), nil
}

func validatePlaintextPassword(plaintext string) error {
	if plaintext == "" {
		return ErrEmptyPassword
	}
	if len(plaintext) > maxPasswordBytes {
		return ErrPasswordTooLong
	}
	return nil
}

type encodedHashParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
}

type parsedPasswordHash struct {
	params     encodedHashParams
	salt       []byte
	key        []byte
	saltLength uint32
}

func encodePasswordHash(config Argon2idConfig, salt, key []byte) string {
	params := fmt.Sprintf("m=%d,t=%d,p=%d", config.Memory, config.Iterations, config.Parallelism)
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	keyB64 := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$%s$v=%d$%s$%s$%s", argon2Algorithm, argon2Version, params, saltB64, keyB64)
}

func parseEncodedPasswordHash(encoded string) (parsedPasswordHash, bool) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" {
		return parsedPasswordHash{}, false
	}
	if parts[1] != argon2Algorithm {
		return parsedPasswordHash{}, false
	}

	version, ok := parseVersion(parts[2])
	if !ok {
		return parsedPasswordHash{}, false
	}
	if version != argon2Version {
		return parsedPasswordHash{}, false
	}

	params, ok := parseHashParams(parts[3])
	if !ok {
		return parsedPasswordHash{}, false
	}

	salt, ok := decodeHashComponent(parts[4], minSaltLength, maxSaltLength)
	if !ok {
		return parsedPasswordHash{}, false
	}
	key, ok := decodeHashComponent(parts[5], minKeyLength, maxKeyLength)
	if !ok {
		return parsedPasswordHash{}, false
	}

	return parsedPasswordHash{
		params:     params,
		salt:       salt,
		key:        key,
		saltLength: uint32(len(salt)),
	}, true
}

func parseVersion(segment string) (int, bool) {
	if !strings.HasPrefix(segment, "v=") {
		return 0, false
	}
	version, err := strconv.Atoi(segment[2:])
	if err != nil || version <= 0 {
		return 0, false
	}
	return version, true
}

func parseHashParams(segment string) (encodedHashParams, bool) {
	if segment == "" || strings.ContainsAny(segment, " \t\n\r\v\f") {
		return encodedHashParams{}, false
	}

	fields := strings.Split(segment, ",")
	if len(fields) != 3 {
		return encodedHashParams{}, false
	}

	expectedKeys := [3]string{"m", "t", "p"}
	var (
		memory      uint64
		iterations  uint64
		parallelism uint64
	)

	for i, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok || value == "" || key != expectedKeys[i] {
			return encodedHashParams{}, false
		}

		var err error
		switch key {
		case "m":
			memory, err = strconv.ParseUint(value, 10, 32)
		case "t":
			iterations, err = strconv.ParseUint(value, 10, 32)
		case "p":
			parallelism, err = strconv.ParseUint(value, 10, 8)
		default:
			return encodedHashParams{}, false
		}
		if err != nil {
			return encodedHashParams{}, false
		}
	}

	if memory == 0 || memory > maxArgon2MemoryKiB {
		return encodedHashParams{}, false
	}
	if iterations == 0 || iterations > maxArgon2Iterations {
		return encodedHashParams{}, false
	}
	if parallelism == 0 || parallelism > maxArgon2Parallelism {
		return encodedHashParams{}, false
	}

	return encodedHashParams{
		Memory:      uint32(memory),
		Iterations:  uint32(iterations),
		Parallelism: uint8(parallelism),
	}, true
}

func decodeHashComponent(segment string, minLen, maxLen int) ([]byte, bool) {
	if segment == "" {
		return nil, false
	}
	decoded, err := base64.RawStdEncoding.DecodeString(segment)
	if err != nil {
		return nil, false
	}
	if len(decoded) < minLen || len(decoded) > maxLen {
		return nil, false
	}
	return decoded, true
}

func (h *PasswordHasher) configDiffers(params encodedHashParams, saltLength, keyLength uint32) bool {
	return params.Memory != h.config.Memory ||
		params.Iterations != h.config.Iterations ||
		params.Parallelism != h.config.Parallelism ||
		saltLength != h.config.SaltLength ||
		keyLength != h.config.KeyLength
}
