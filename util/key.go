package util

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"math/rand"
	"os"
	"strings"

	"github.com/anomalous69/fchannel/config"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
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

// FIXME: appeers cookiekey is not not userd anywhere
func GetCookieKey() (string, error) {
	if config.C.CookieKey == "" {
		var file *os.File
		var err error

		if file, err = os.OpenFile("config/config-init", os.O_APPEND|os.O_WRONLY, 0644); err != nil {
			return "", MakeError(err, "GetCookieKey")
		}

		defer file.Close()

		config.C.CookieKey = encryptcookie.GenerateKey()
		file.WriteString("\ncookiekey:" + config.C.CookieKey)
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
