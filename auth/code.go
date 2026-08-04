package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
)

func GenerateChallengeID() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate challenge ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

func GenerateCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate authentication code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func HashCode(secret []byte, challengeID, code string) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(challengeID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil)
}
