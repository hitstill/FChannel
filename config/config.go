package config

import (
	"strconv"

	"github.com/spf13/viper"
)

var Port = ":" + GetConfigValue("instance.port", "3000")
var TP = GetConfigValue("instance.tp", "")
var Domain = TP + "" + GetConfigValue("instance.domain", "")
var InstanceName = GetConfigValue("instance.name", "")
var InstanceSummary = GetConfigValue("instance.summary", "")
var Salt = GetConfigValue("instance.salt", "")
var SiteEmail = GetConfigValue("email.address", "")
var SiteEmailUsername = GetConfigValue("email.user", "")
var SiteEmailPassword = GetConfigValue("email.pass", "")
var SiteEmailServer = GetConfigValue("email.server", "")
var SiteEmailPort = GetConfigValue("email.port", "")
var SiteEmailNotifyTo = GetConfigValue("email.notify", "")
var NtfyURL = GetConfigValue("ntfy.url", "")
var NtfyAuth = GetConfigValue("ntfy.auth", "")
var TorProxy = GetConfigValue("torproxy", "")
var DBHost = GetConfigValue("db.host", "localhost")
var DBPort, _ = strconv.Atoi(GetConfigValue("db.port", "5432"))
var DBUser = GetConfigValue("db.user", "postgres")
var DBPassword = GetConfigValue("db.pass", "password")
var DBName = GetConfigValue("db.name", "server")
var CookieKey = GetConfigValue("cookiekey", "")
var ActivityStreams = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
var AuthReq = []string{"captcha", "email", "passphrase"}
var PostCountPerPage = 10
var SupportedFiles = []string{"image/avif", "image/gif", "image/jpeg", "image/jxl", "image/png", "image/webp", "image/apng", "video/mp4", "video/ogg", "video/webm", "audio/mpeg", "audio/ogg", "audio/wav", "audio/wave", "audio/x-wav", "application/x-shockwave-flash"}
var MediaHashs = make(map[string]string)
var Key = GetConfigValue("modkey", "")
var MinPostDelete = GetConfigValue("minpostdelete", "60")
var MaxPostDelete = GetConfigValue("maxpostdelete", "1800")

// TODO: this is bad but I don't feel like doing a new config system yet, and I can't into computers
var MaxAttachmentSize, _ = strconv.Atoi(GetConfigValue("maxattachsize", "7340032"))
var MaxMindDB = GetConfigValue("maxminddb", "")
var TorExitList = GetConfigValue("torexitlist", "")
var ProxyHeader = GetConfigValue("proxyheader", "")
var CaptchaFont = GetConfigValue("captchafont", "") // TODO: should probably check user not passing anything weird to exec.Command
var Themes []string

var Version string
var Debug = viper.GetBool("debug")

func GetConfigValue(value string, ifnone string) string {
	viper.SetDefault(value, ifnone)
	return viper.GetString(value)
}

func ReadConfig() error {
	viper.SetConfigName("fchan") // name of config file (without extension)
	viper.SetConfigType("yaml")  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")     // optionally look for config in the working directory
	return viper.ReadInConfig()  // Find and read the config file
}
