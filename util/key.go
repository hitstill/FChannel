package util

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/anomalous69/fchannel/config"
)

const domain = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func CreateKey(length int) (string, error) {
	if length > 128 {
		return "", MakeError(errors.New("len is greater than 128"), "CreateKey")
	}

	str := CreateTripCode(RandomID(length))
	return str[:length], nil
}

func CreateTripCode(input string) string {
	out := sha512.Sum512([]byte(input))

	return hex.EncodeToString(out[:])
}

func GetCookieKey() (string, error) {
	if config.C.CookieKey == "" {
		return "", MakeError(errors.New("cookie_key in not set in the fchan.yaml file\n Run openssl rand -base64 32 to generate"), "GetCookieKey")
	}

	return config.C.CookieKey, nil
}

func RandomID(size int) string {
	newID := strings.Builder{}
	sizeAsBig := big.NewInt(int64(len(domain)))

	for range size {
		randIndex, _ := rand.Int(rand.Reader, sizeAsBig)

		newID.WriteByte(domain[randIndex.Int64()])
	}
	hashed := sha512.Sum512([]byte(newID.String()))
	return hex.EncodeToString(hashed[:])
}
