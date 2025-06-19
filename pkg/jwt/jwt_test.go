package jwt_test

import (
	"goadv/pkg/jwt"
	"testing"
)

func TestJWT(t *testing.T) {
	const email = "a@a.ru"

	jwtService := jwt.NewJWT("hnH1ly3sNC3ypDmNJXEMXUH6Na+CNzakcwVcGHYGY3Q=")
	token, err := jwtService.Create(jwt.JWTData{
		Email: email,
	})
	if err != nil {
		t.Fatal(err)
	}
	isValid, data := jwtService.Parse(token)
	if !isValid {
		t.Fatalf("Invalid token")
	}
	if data.Email != email {
		t.Fatalf("Expected %s got %s", email, data.Email)
	}
}
