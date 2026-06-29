package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"
)

const (
	digitLength = 6
	timeStep    = 30
	secretSize  = 20
)

func GenerateSecret() (string, error) {
	secret := make([]byte, secretSize)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

func GenerateCode(secret string, t time.Time) (string, error) {
	counter := uint64(t.Unix() / timeStep)
	code, err := hotp(secret, counter)
	if err != nil {
		return "", err
	}
	return code, nil
}

func ValidateCode(secret, code string, t time.Time, skew int) bool {
	counter := uint64(t.Unix() / timeStep)
	for d := -skew; d <= skew; d++ {
		adj := uint64(int64(counter) + int64(d))
		expected, err := hotp(secret, adj)
		if err != nil {
			continue
		}
		if expected == code {
			return true
		}
	}
	return false
}

func GenerateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 4)
		rand.Read(b)
		codes[i] = fmt.Sprintf("%08x", b)
	}
	return codes
}

func GetQRCodeURL(secret, issuer, accountName string) string {
	issuerEnc := url.QueryEscape(issuer)
	accountEnc := url.QueryEscape(accountName)
	secretEnc := url.QueryEscape(secret)
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuerEnc, accountEnc, secretEnc, issuerEnc, digitLength, timeStep)
}

func hotp(secret string, counter uint64) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(counter & 0xff)
		counter >>= 8
	}

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0xf
	binary := (uint32(hash[offset]&0x7f) << 24) |
		(uint32(hash[offset+1]) << 16) |
		(uint32(hash[offset+2]) << 8) |
		uint32(hash[offset+3])

	code := binary % uint32(math.Pow10(digitLength))
	return fmt.Sprintf("%0*d", digitLength, code), nil
}
