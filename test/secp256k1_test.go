package test

import (
	"fmt"
	"github.com/w3liu/go-common/crypto/secp256k1"
	"testing"
)

func TestGenkey(t *testing.T) {
	key, err := secp256k1.NewPrivateKey(secp256k1.S256())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(key)
}
