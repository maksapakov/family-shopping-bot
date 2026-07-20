package callback

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Signer struct {
	secret []byte
}

func NewSigner(secret string) (*Signer, error) {
	if secret == "" {
		return nil, fmt.Errorf("callback: secret is empty")
	}
	return &Signer{secret: []byte(secret)}, nil
}

func (s *Signer) SignToggle(chatID string, itemID string) string {
	payload := fmt.Sprintf("toggle\n%s\n%s", chatID, itemID)
	return s.sign(payload)
}

func (s *Signer) VerifyToggle(chatID string, itemID string, sig string) bool {
	return s.verify(fmt.Sprintf("toggle\n%s\n%s", chatID, itemID), sig)
}

func (s *Signer) SignUndo(chatID string) string {
	return s.sign(fmt.Sprintf("undo\n%s", chatID))
}

func (s *Signer) VerifyUndo(chatID string, sig string) bool {
	return s.verify(fmt.Sprintf("undo\n%s", chatID), sig)
}

func (s *Signer) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Signer) verify(payload string, sig string) bool {
	expected, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return hmac.Equal(mac.Sum(nil), expected)
}
