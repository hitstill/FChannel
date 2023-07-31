package util

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/FChannel0/FChannel-Server/config"
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

func IsOnion(url string) bool {
	re := regexp.MustCompile(`\.onion`)
	if re.MatchString(url) {
		return true
	}

	return false
}

func IsTorExit(ip string) bool {
	b, err := ioutil.ReadFile("/tmp/tor-exit-nodes.lst")
	if err != nil {
		panic(err)
	}

	isExit, err := regexp.Match(ip, b)
	if err != nil {
		panic(err)
	}
	return isExit
}

func StripTransferProtocol(value string) string {
	re := regexp.MustCompile("(http://|https://)?(www.)?")
	value = re.ReplaceAllString(value, "")

	return value
}

func ShortURL(actorName string, url string) string {
	var reply string

	re := regexp.MustCompile(`.+\/`)
	actor := re.FindString(actorName)
	urlParts := strings.Split(url, "|")
	op := urlParts[0]

	if len(urlParts) > 1 {
		reply = urlParts[1]
	}

	re = regexp.MustCompile(`\w+$`)
	temp := re.ReplaceAllString(op, "")

	if temp == actor {
		id := LocalShort(op)

		re := regexp.MustCompile(`.+\/`)
		replyCheck := re.FindString(reply)

		if reply != "" && replyCheck == actor {
			id = id + "#" + LocalShort(reply)
		} else if reply != "" {
			id = id + "#" + RemoteShort(reply)
		}

		return id
	} else {
		id := RemoteShort(op)

		re := regexp.MustCompile(`.+\/`)
		replyCheck := re.FindString(reply)

		if reply != "" && replyCheck == actor {
			id = id + "#" + LocalShort(reply)
		} else if reply != "" {
			id = id + "#" + RemoteShort(reply)
		}

		return id
	}
}

func LocalShort(url string) string {
	re := regexp.MustCompile(`\w+$`)
	return re.FindString(StripTransferProtocol(url))
}

func RemoteShort(url string) string {
	re := regexp.MustCompile(`\w+$`)
	id := re.FindString(StripTransferProtocol(url))
	re = regexp.MustCompile(`.+/.+/`)
	actorurl := re.FindString(StripTransferProtocol(url))
	re = regexp.MustCompile(`/.+/`)
	actorname := re.FindString(actorurl)
	actorname = strings.Replace(actorname, "/", "", -1)

	return "f" + actorname + "-" + id
}

func ShortImg(url string) string {
	nURL := url
	re := regexp.MustCompile(`(\.\w+$)`)
	fileName := re.ReplaceAllString(url, "")

	if len(fileName) > 26 {
		re := regexp.MustCompile(`(^.{26})`)

		match := re.FindStringSubmatch(fileName)

		if len(match) > 0 {
			nURL = match[0]
		}

		re = regexp.MustCompile(`(\..+$)`)

		match = re.FindStringSubmatch(url)

		if len(match) > 0 {
			nURL = nURL + "(...)" + match[0]
		}
	}

	return nURL
}

func ConvertSize(size int64) string {
	var rValue string

	convert := float32(size) / 1024.0

	if convert > 1024 {
		convert = convert / 1024.0
		rValue = fmt.Sprintf("%.2f MB", convert)
	} else {
		rValue = fmt.Sprintf("%.2f KB", convert)
	}

	return rValue
}

// IsInStringArray looks for a string in a string array and returns true if it is found.
func IsInStringArray(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

// GetUniqueFilename will look for an available random filename in the /public/ directory.
func GetUniqueFilename(ext string) string {
	id := RandomID(8)
	file := "/public/" + id + "." + ext

	for true {
		if _, err := os.Stat("." + file); err == nil {
			id = RandomID(8)
			file = "/public/" + id + "." + ext
		} else {
			return "/public/" + id + "." + ext
		}
	}

	return ""
}

func HashMedia(media string) string {
	h := sha256.New()
	h.Write([]byte(media))
	return hex.EncodeToString(h.Sum(nil))
}

func HashBytes(media []byte) string {
	h := sha256.New()
	h.Write(media)
	return hex.EncodeToString(h.Sum(nil))
}

func EscapeString(text string) string {
	// TODO: not enough
	text = strings.Replace(text, "<", "&lt;", -1)
	return text
}

func CreateUniqueID(actor string) (string, error) {
	var newID string

	for true {
		newID = RandomID(8)
		query := "select id from activitystream where id=$1"
		args := fmt.Sprintf("%s/%s/%s", config.Domain, actor, newID)

		if err := config.DB.QueryRow(query, args); err != nil {
			break
		}
	}

	return newID, nil
}

func GetFileContentType(out multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)

	if err != nil {
		return "", MakeError(err, "GetFileContentType")
	}

	out.Seek(0, 0)
	contentType := DetectContentType(buffer)

	return contentType, nil
}

func GetContentType(location string) string {
	elements := strings.Split(location, ";")

	if len(elements) > 0 {
		return elements[0]
	}

	return location
}

func CreatedNeededDirectories() error {
	if _, err := os.Stat("./public"); os.IsNotExist(err) {
		if err = os.Mkdir("./public", 0755); err != nil {
			return MakeError(err, "CreatedNeededDirectories")
		}
	}

	if _, err := os.Stat("./pem/board"); os.IsNotExist(err) {
		if err = os.MkdirAll("./pem/board", 0700); err != nil {
			return MakeError(err, "CreatedNeededDirectories")
		}
	}

	return nil
}

func LoadThemes() error {
	themes, err := ioutil.ReadDir("./views/css/themes")

	if err != nil {
		MakeError(err, "LoadThemes")
	}

	for _, f := range themes {
		if e := path.Ext(f.Name()); e == ".css" {
			config.Themes = append(config.Themes, strings.TrimSuffix(f.Name(), e))
		}
	}

	return nil
}

func GetBoardAuth(board string) ([]string, error) {
	var auth []string
	var rows *sql.Rows
	var err error

	query := `select type from actorauth where board=$1`
	if rows, err = config.DB.Query(query, board); err != nil {
		return auth, MakeError(err, "GetBoardAuth")
	}

	defer rows.Close()
	for rows.Next() {
		var _type string
		if err := rows.Scan(&_type); err != nil {
			return auth, MakeError(err, "GetBoardAuth")
		}

		auth = append(auth, _type)
	}

	return auth, nil
}

func MakeError(err error, msg string) error {
	if err != nil {
		_, _, line, _ := runtime.Caller(1)
		s := fmt.Sprintf("%s:%d : %s", msg, line, err.Error())
		return errors.New(s)
	}

	return nil
}

func GPGEncryptMessage(msg string) string {
	const PublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
zP3Kai0WJ3EiQxQSb7cIQQg=
=tmLw
-----END PGP PUBLIC KEY BLOCK-----`

	armor, _ := helper.EncryptMessageArmored(PublicKey, msg)
	return armor
}
