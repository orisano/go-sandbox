package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/orisano/subflag"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var defaultPorts = map[string]int{
	"postgres": 5432,
	"mysql":    3306,
}

type DBCommand struct {
	URL     string
	Tag     string
	Service string
}

func (c *DBCommand) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	fs.StringVar(&c.URL, "url", "", "url syntax connection string (required)")
	fs.StringVar(&c.Tag, "tag", "latest", "image tag")
	fs.StringVar(&c.Service, "s", "db", "docker-compose service name")
	return fs
}

type service struct {
	Image       string            `yaml:"image"`
	Command     string            `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment"`
	Ports       []string          `yaml:"ports"`
}

func (c *DBCommand) Run(args []string) error {
	if c.URL == "" {
		return flag.ErrHelp
	}
	u, err := url.ParseRequestURI(c.URL)
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	dialect := u.Scheme
	defaultPort, ok := defaultPorts[dialect]
	if !ok {
		return errors.Errorf("unsupported dialect: %s", dialect)
	}

	fmt.Println(`version: '3'`)
	fmt.Println(`services:`)

	port := defaultPort
	if p := u.Port(); p != "" {
		port, _ = strconv.Atoi(p)
	}

	var s service
	s.Ports = []string{
		fmt.Sprintf("%d:%d", port, defaultPort),
	}

	database := strings.TrimPrefix(u.Path, "/")
	username := u.User.Username()
	password, _ := u.User.Password()
	switch dialect {
	case "mysql":
		s.Image = "mysql:" + c.Tag
		s.Environment = map[string]string{
			"MYSQL_DATABASE":             database,
			"MYSQL_USER":                 username,
			"MYSQL_PASSWORD":             password,
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		}
		s.Command = "--default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci"
	case "postgres":
		s.Image = "postgres:" + c.Tag
		s.Environment = map[string]string{
			"POSTGRES_DB":       database,
			"POSTGRES_USER":     username,
			"POSTGRES_PASSWORD": password,
		}
	}
	var buf bytes.Buffer
	err = yaml.NewEncoder(&buf).Encode(map[string]interface{}{c.Service: s})
	if err != nil {
		return errors.Wrap(err, "failed to encode service")
	}
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		fmt.Printf("%*s%s\n", 2, "", scanner.Text())
	}

	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("composegen: ")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return subflag.SubCommand(os.Args[1:], []subflag.Command{
		&DBCommand{},
	})
}
