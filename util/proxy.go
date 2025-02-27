package util

import (
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/anomalous69/fchannel/config"
)

func GetPathProxyType(path string) string {
	if config.C.Proxy != "" {
		re := regexp.MustCompile(`(http://|http://)?(www.)?(\w+\.onion|\w+\.loki|\w+\.i2p)`)
		onion := re.MatchString(path)

		if onion {
			return "tor"
		}
	}

	return "clearnet"
}

func MediaProxy(url string) string {
	re := regexp.MustCompile("(.+)?" + config.C.Instance.Domain + "(.+)?")
	if re.MatchString(url) {
		return url
	}

	re = regexp.MustCompile(`(.+)?(\.onion(.+)|\.loki(.+)|\.i2p(.+))?`)
	if re.MatchString(url) {
		return url
	}

	config.MediaHashs[HashMedia(url)] = url

	return "/api/media?hash=" + HashMedia(url)
}

func RouteProxy(req *http.Request) (*http.Response, error) {
	var proxyType = GetPathProxyType(req.URL.Host)

	req.Header.Set("User-Agent", "FChannel/"+config.C.Instance.Name)

	if proxyType == "tor" {
		proxyUrl, err := url.Parse(config.C.Proxy)

		if err != nil {
			return nil, MakeError(err, "RouteProxy")
		}

		proxyTransport := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		client := &http.Client{Transport: proxyTransport, Timeout: time.Second * 60}

		return client.Do(req)
	}

	return http.DefaultClient.Do(req)
}
