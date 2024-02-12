package jwtUtils

import "testing"

var jwtString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.eYAVln5D77gv54Xvm5IdNyV4OZWEBSvZUnfuqsgtPVU"

func TestCreateJWT(t *testing.T) {
	expected := jwtString
	token, err := CreateJWT("test")
	if err != nil {
		t.Errorf("received error but expected a token")
	}
	if *token != expected {
		t.Errorf("expected: %v, received: %v", expected, *token)
	}
}

func TestValidateJWT(t *testing.T) {
	expected := "test"
	user, err := ValidateJWT(jwtString)
	if err != nil {
		t.Error(err)
	}
	if *user != expected {
		t.Errorf("expected: %v, received: %v", expected, *user)
	}

}
