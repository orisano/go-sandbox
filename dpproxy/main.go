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

type Options struct {
	config  string
	addr    string
	verbose bool
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("dpproxy: ")

	var opt Options
	flag.StringVar(&opt.config, "c", "", "config file (required)")
	flag.StringVar(&opt.addr, "l", ":8080", "listen addr")
	flag.BoolVar(&opt.verbose, "v", false, "verbose")
	flag.Parse()

	if opt.config == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(&opt); err != nil {
		log.Fatal(err)
	}
}

func run(opt *Options) error {
	var c struct {
		Rule []*Rule
	}
	if _, err := toml.DecodeFile(opt.config, &c); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	if opt.verbose {
		log.Print("rewrite rules:")
		for _, r := range c.Rule {
			log.Print("  - ", r.From, " -> ", r.To)
		}
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
				if opt.verbose {
					log.Print("rewrite ", host, " -> ", to)
				}
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

	log.Print("listen ", opt.addr, " ...")
	return http.ListenAndServe(opt.addr, proxy)
}
