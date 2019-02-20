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
		Rule []*Rule
	}
	if _, err := toml.DecodeFile(config, &c); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	replace := make(map[string]string, len(c.Rule))
	for _, r := range c.Rule {
		replace[r.From] = r.To
	}

	rewriteAddr := func(network string, addr string) string {
		if network == "tcp" {
			u := &url.URL{Host: addr}
			host := u.Hostname()
			if to, ok := replace[host]; ok {
				addr = to + addr[len(host):]
			}
		}
		return addr
	}

	proxy := goproxy.NewProxyHttpServer()

	// goproxy does not use DialContext
	proxy.Tr.Dial = func(network, addr string) (conn net.Conn, e error) {
		addr = rewriteAddr(network, addr)
		return net.Dial(network, addr)
	}

	originalDial := proxy.ConnectDial
	proxy.ConnectDial = func(network string, addr string) (net.Conn, error) {
		addr = rewriteAddr(network, addr)
		if originalDial != nil {
			return originalDial(network, addr)
		}
		return net.Dial(network, addr)
	}

	return http.ListenAndServe(listen, proxy)
}
