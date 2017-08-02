package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/acoshift/cachestatic"
	"github.com/acoshift/header"
	"github.com/acoshift/middleware"
	"gopkg.in/yaml.v2"
)

type config struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Hosts []struct {
		Host   string `yaml:"host"`
		Target string `yaml:"target"`
	} `yaml:"hosts"`
}

var (
	configFile = flag.String("config", "config.yaml", "Config file")
)

type hostMux map[string]http.Handler

func (x hostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}

	host := r.Host
	if len(host) == 0 {
		host = r.Header.Get(header.XForwardedHost)
	}
	h, ok := x[host]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("wpcache host not found"))
		return
	}
	h.ServeHTTP(w, r)
}

func modifyResponse(resp *http.Response) error {
	resp.Header.Add(header.Via, "acoshift/wpcache")

	if resp.Request != nil {
		// remove cookie if not admin section
		// u := resp.Request.URL
		// if !isAdminSection(u) {
		// 	resp.Header.Del(header.SetCookie)
		// }
	}
	return nil
}

var staticExt = makeMapStringStruct(
	".js", ".css",
	".jpg", ".jpeg", ".gif", ".png", ".ico", ".svg",
	".eot", ".otf", ".woff", ".ttf",
	".mp3", ".mp4",
)

func makeMapStringStruct(ss ...string) map[string]struct{} {
	r := make(map[string]struct{})
	for _, s := range ss {
		r[s] = struct{}{}
	}
	return r
}

func isStatic(u *url.URL) bool {
	ext := filepath.Ext(u.Path)
	_, ok := staticExt[ext]
	return ok
}

func isAdminSection(u *url.URL) bool {
	return strings.HasPrefix(u.Path, "/wp-login")
}

func cacheSkipper(r *http.Request) bool {
	u := r.URL
	if isStatic(u) {
		return false
	}
	// TODO: cache more page
	return true
}

func main() {
	flag.Parse()

	// laod config
	var c config
	{
		bs, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("load config error; %v", err)
		}
		err = yaml.Unmarshal(bs, &c)
		if err != nil {
			log.Fatalf("load config error; %v", err)
		}
	}
	mux := make(hostMux)
	for _, host := range c.Hosts {
		if _, ok := mux[host.Host]; ok {
			log.Fatalf("load config error; host %s duplicated", host.Host)
		}
		target, err := url.Parse(host.Target)
		if err != nil {
			log.Fatalf("load config; parse url error; %v", err)
		}
		h := httputil.NewSingleHostReverseProxy(target)
		h.ModifyResponse = modifyResponse
		mux[host.Host] = middleware.Chain(
			cachestatic.New(cachestatic.Config{
				Skipper: cacheSkipper,
			}),
		)(h)
	}
	http.ListenAndServe(":8080", mux)
}
