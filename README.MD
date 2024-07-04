# Golang DNS server

## build
run go build to build the binary

build for your env
```shell
go build
```
cross build for platform
```shell
env GOOS=linux GOARCH=amd64 go build
```

## Configs
edit these nodes below from the config/config.yaml file:
```yml
# Logger configurations.
localmode: true

dnsserver:
  address: ""
  port: 53
  forwardto: "1.1.1.1:53"

logger:
  path: "/tmp/"
  file: "dns-{Y-m-d}.log"
  level: "all"
  stdout: true

database:
  sqlite:
    type: "sqlite3"
    link: "config/dnsdb.db"
    debug: false
```

## run
```shell
./godns 
```