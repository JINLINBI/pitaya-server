package env

import "testing"

func TestNeedDB(t *testing.T) {
	cases := []struct {
		typ  string
		want bool
	}{
		{"monolith", true},
		{"standalone", true},
		{"account", true},
		{"connector", false},
		{"room", false},
	}
	for _, c := range cases {
		if got := NeedDB(c.typ); got != c.want {
			t.Fatalf("%s: want %v got %v", c.typ, c.want, got)
		}
	}
}

func TestValidateJWTSecret(t *testing.T) {
	dev := Settings{AuthJWTSecret: DevJWTSecret}
	if err := dev.Validate(true); err == nil {
		t.Fatal("cluster should reject dev secret")
	}
	if err := dev.Validate(false); err != nil {
		t.Fatalf("monolith dev secret: %v", err)
	}
	ok := Settings{AuthJWTSecret: "prod-secret"}
	if err := ok.Validate(true); err != nil {
		t.Fatalf("cluster custom secret: %v", err)
	}
}