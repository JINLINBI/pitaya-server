package auth

import "testing"

func TestJWTRoundTrip(t *testing.T) {
	j := NewJWT("test-secret", 0)
	token, _, err := j.Issue("42", "alice")
	if err != nil {
		t.Fatal(err)
	}
	id, err := j.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != "42" {
		t.Fatalf("want 42 got %q", id)
	}
	claims, err := j.VerifyClaims(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Username != "alice" {
		t.Fatalf("want alice got %q", claims.Username)
	}
}