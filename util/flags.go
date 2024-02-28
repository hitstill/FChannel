package util

import (
	"net"
	"strings"

	"github.com/FChannel0/FChannel-Server/config"
	"github.com/oschwald/maxminddb-golang"
)

// func GetFlag(ip string) (string) {
// 	if isTorExit(ip) || ip == "172.16.0.1" {
// 		return "🏴‍☠️ "
// }
// 	db, err := maxminddb.Open("/usr/share/GeoIP/GeoLite2-Country.mmdb")
// 	if err != nil {
// 		return "🏴 "
// 	}
// 	defer db.Close()

// 	var record struct {
// 		Country struct {
// 			ISOCode string `maxminddb:"iso_code"`
// 		} `maxminddb:"country"`
// 	}

// 	err = db.Lookup(net.ParseIP(ip), &record)
// 	if err != nil {
// 		return "🏴 "
// 	}
// 	code := record.Country.ISOCode
// 	//you probably want to check the returned isocode actually has a unicode character but eh
// 	return  string(0x1F1E6+rune(code[0])-'A') + string(0x1F1E6+rune(code[1])-'A') + " "
// }

func GetCC(ip string) string {
	if IsTorExit(ip) || ip == "172.16.0.1" {
		return "xp"
	}
	
	db, err := maxminddb.Open(config.MaxMindDB)
	if err != nil {
		return "xx"
	}
	defer db.Close()

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	err = db.Lookup(net.ParseIP(ip), &record)
	if err != nil {
		return "xx"
	}
	code := strings.ToLower(record.Country.ISOCode)
	return code
}
