# dpproxy
dpproxy is the dns poisoning http proxy.

## Installation
```
go get -u github.com/orisano/go-sandbox/dpproxy
```

## How to use
```
$ dpproxy -h
Usage of dpproxy:
  -c string
    	config file (required)
  -l string
    	listen addr (default ":8080")
  -v	verbose
```
```
$ cat config.toml
[[rule]]
from = "foo.com"
to = "bar.com"

[[rule]]
from = "example.com"
to = "to.example.com"
```

## Author
Nao Yonashiro (@orisano)

## License
MIT
