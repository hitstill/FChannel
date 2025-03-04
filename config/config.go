package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	Instance struct {
		Port    int
		Scheme  string
		Name    string
		Domain  string
		Summary string
		Salt    string
	}
	Email struct {
		Address  string
		User     string
		Password string
		Server   string
		Port     int
		NotifyTo string `mapstructure:"notify_to"`
	}
	Db struct {
		Host     string
		Port     int
		User     string
		Password string
		Database string
	}
	Ntfy struct {
		Url  string
		Auth string
	}

	Posts struct {
		MaxAttachmentSize         int `mapstructure:"max_attachment_size"`
		RemovableNotBeforeSeconds int `mapstructure:"removable_not_before_seconds"`
		RemovableNotAfterSeconds  int `mapstructure:"removable_not_after_seconds"`
	}

	Proxy       string
	CookieKey   string `mapstructure:"cookie_key"`
	ModKey      string `mapstructure:"mod_key"`
	MaxMindDb   string `mapstructure:"maxmind_db"`
	TorExitList string `mapstructure:"tor_exit_list"`
	Salt        string
	ProxyHeader string `mapstructure:"proxy_header"`
	CaptchaFont string `mapstructure:"captcha_font"`
	Debug       bool
}

var C AppConfig

var ActivityStreams = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
var AuthReq = []string{"captcha", "email", "passphrase"}
var PostCountPerPage = 10
var SupportedFiles = []string{"image/avif", "image/gif", "image/jpeg", "image/jxl", "image/png", "image/webp", "image/apng", "video/mp4", "video/ogg", "video/webm", "audio/mpeg", "audio/ogg", "audio/wav", "audio/wave", "audio/x-wav", "application/x-shockwave-flash"}
var MediaHashs = make(map[string]string)
var Themes []string

var Version string

func ReadConfig() error {
	viper.SetConfigName("fchan.yaml") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // optionally look for config in the working directory
	setDefaults()
	err := viper.ReadInConfig() // Find and read the config file

	viper.Unmarshal(&C)

	return err
}

func setDefaults() {
	viper.SetDefault("debug", false)

	viper.SetDefault("instance.port", 3000)

	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "postgres")
	viper.SetDefault("db.password", "postgres")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.database", "server")

	viper.SetDefault("posts.removable_not_before_seconds", 60)
	viper.SetDefault("posts.removable_not_after_seconds", 1800)
	viper.SetDefault("posts.max_attach_size", 7340032)

	viper.SetDefault("email.port", 25)
}
