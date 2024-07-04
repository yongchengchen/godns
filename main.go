package main

import (
	"flag"
	"fmt"

	"github.com/yongchengchen/godns/app/api"

	"github.com/sirupsen/logrus"

	"github.com/miekg/dns"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	_ "github.com/yongchengchen/godns/library/driver"
	_ "github.com/yongchengchen/godns/router"

	"os"
	"os/signal"
	"syscall"
)

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1) //since has 2 service, so exit twice
		os.Exit(1)
	}()

	s := g.Server()
	s.Start()

	cx := gctx.New()
	address, _ := g.Cfg().Get(cx, "dnsserver.address")
	port, _ := g.Cfg().Get(cx, "dnsserver.port")
	forwardto, _ := g.Cfg().Get(cx, "dnsserver.forwardto")

	bDebug := flag.Bool("debug", false, "Output debug message")
	flag.Parse()
	if *bDebug {
		logrus.Println("Debug ", *bDebug)
		logrus.SetLevel(logrus.DebugLevel)
	}

	// start server
	api.DnsRecordApi.Forwarder = forwardto.String()
	addr := fmt.Sprintf("%s:%s", address.String(), port.String())
	server := &dns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: api.DnsRecordApi,
	}
	logrus.Printf("Starting at %s\n", addr)
	if api.DnsRecordApi.Forwarder == "" {
		logrus.Println("  -> no fallback forward.")
	} else {
		logrus.Printf("  ->fallback will forward to %s\n", api.DnsRecordApi.Forwarder)
	}

	api.DnsRecordApi.ListRecords()

	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		logrus.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
