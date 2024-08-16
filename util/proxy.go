package util

import (
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/FChannel0/FChannel-Server/config"
)

func GetPathProxyType(path string) string {
	if config.TorProxy != "" {
		re := regexp.MustCompile(`(http://|http://)?(www.)?(\w+\.onion|\w+\.loki|\w+\.i2p)`)
		onion := re.MatchString(path)

		if onion {
			return "tor"
		}
	}

	return "clearnet"
}

func MediaProxy(url string) string {
	re := regexp.MustCompile("(.+)?" + config.Domain + "(.+)?")
	if re.MatchString(url) {
		return url
	}

	re = regexp.MustCompile("(.+)?(\\.onion(.+)|\\.loki(.+)|\\.i2p(.+))?")
	if re.MatchString(url) {
		return url
	}

	config.MediaHashs[HashMedia(url)] = url

	return "/api/media?hash=" + HashMedia(url)
}

func RouteProxy(req *http.Request) (*http.Response, error) {
	var proxyType = GetPathProxyType(req.URL.Host)

	req.Header.Set("User-Agent", "FChannel/"+config.InstanceName)

	if proxyType == "tor" {
		proxyUrl, err := url.Parse(config.TorProxy)

		if err != nil {
			return nil, MakeError(err, "RouteProxy")
		}

		proxyTransport := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		client := &http.Client{Transport: proxyTransport, Timeout: time.Second * 30}

		return client.Do(req)
	}

	return http.DefaultClient.Do(req)
}
