package password

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultConfig = Config{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func Hash(password string) (string, error) {
	return HashWithConfig(password, DefaultConfig)
}

func HashWithConfig(password string, cfg Config) (string, error) {
	salt := make([]byte, cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, cfg.Iterations, cfg.Memory, cfg.Parallelism, cfg.KeyLength)

	b64Salt := hex.EncodeToString(salt)
	b64Hash := hex.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, cfg.Memory, cfg.Iterations, cfg.Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func Verify(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("parse params: %w", err)
	}

	salt, err := hex.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := hex.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	return hex.EncodeToString(computed) == hex.EncodeToString(hash), nil
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateAPIKey() (rawKey, keyHash, keyPrefix string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", fmt.Errorf("generate key: %w", err)
	}
	rawKey = "nxc_" + hex.EncodeToString(bytes)
	keyPrefix = rawKey[:11]

	hash := sha256.Sum256([]byte(rawKey))
	keyHash = hex.EncodeToString(hash[:])

	return rawKey, keyHash, keyPrefix, nil
}

func HashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
