package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

func Sha256AfterSha256(data []byte) [32]byte {
	hash256 := sha256.Sum256(data)
	hash256 = sha256.Sum256(hash256[:])
	return hash256
}

func Ripemd160AfterSha256(data []byte) ([]byte, error) {
	hash256 := sha256.Sum256(data)
	ripemd160Obj := ripemd160.New()
	if _, err := ripemd160Obj.Write(hash256[:]); err != nil {
		return nil, err
	}
	return ripemd160Obj.Sum(nil), nil
}

func Keccak256Hash(data []byte) []byte {
	keccak256Hash2 := sha3.NewLegacyKeccak256()
	keccak256Hash2.Write(data)
	return keccak256Hash2.Sum(nil)
}

func HMACWithSHA512(seed []byte, key []byte) []byte {
	hmac512 := hmac.New(sha512.New, key)
	hmac512.Write(seed)
	return hmac512.Sum(nil)
}
