package util

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"

	"github.com/anomalous69/fchannel/config"
)

const domain = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func CreateKey(len int) (string, error) {
	// TODO: provided that CreateTripCode still uses sha512, the max len can be 128 at most.
	if len > 128 {
		return "", MakeError(errors.New("len is greater than 128"), "CreateKey")
	}

	str := CreateTripCode(RandomID(len))
	return str[:len], nil
}

func CreateTripCode(input string) string {
	out := sha512.Sum512([]byte(input))

	return hex.EncodeToString(out[:])
}

func GetCookieKey() (string, error) {
	if config.C.CookieKey == "" {
		panic("cookie_key in not set in the fchan.yaml file\n Run openssl rand -base64 32 to generate")
	}

	return config.C.CookieKey, nil
}

func RandomID(size int) string {
	rng := size
	newID := strings.Builder{}

	for i := 0; i < rng; i++ {
		newID.WriteByte(domain[rand.Intn(len(domain))])
	}

	return newID.String()
}
