package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
)

type Rule struct {
	From string
	To   string
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("dpproxy: ")

	config := flag.String("c", "", "config file (required)")
	listen := flag.String("l", ":8080", "listen addr")
	flag.Parse()

	if *config == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*listen, *config); err != nil {
		log.Fatal(err)
	}
}

func run(listen, config string) error {
	var c struct {
		rule []*Rule
	}
	if _, err := toml.DecodeFile(config, &c); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	replace := make(map[string]string, len(c.rule))
	for _, r := range c.rule {
		replace[r.From] = r.To
	}

	proxy := goproxy.NewProxyHttpServer()
	originalDial := proxy.ConnectDial
	proxy.ConnectDial = func(network string, addr string) (net.Conn, error) {
		if network == "tcp" {
			u := &url.URL{Host: addr}
			host := u.Hostname()
			if to, ok := replace[host]; ok {
				addr = to + addr[len(host):]
			}
		}
		if originalDial != nil {
			return originalDial(network, addr)
		}
		return net.Dial(network, addr)
	}

	return http.ListenAndServe(listen, proxy)
}
