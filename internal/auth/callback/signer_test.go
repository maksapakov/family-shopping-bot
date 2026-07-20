package callback

import "testing"

func TestSigner_toggleRoundTrip(t *testing.T) {
	s, err := NewSigner("test-secret")
	if err != nil {
		t.Fatal("failed to create signer", err)
	}
	sig := s.SignToggle("demo-chat", "item-1")
	if !s.VerifyToggle("demo-chat", "item-1", sig) {
		t.Fatal("failed to verify toggle, expected valid signature")
	}
	if s.VerifyToggle("demo-chat", "item-1", "deadbeef") {
		t.Fatal("failed to verify toggle, expected invalid signature")
	}
}
