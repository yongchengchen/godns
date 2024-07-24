package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"github.com/miekg/dns"
)

// 用户API管理对象
var DnsDoh = new(dnsDoh)

type dnsDoh struct {
	cloudflareIPRanges []string
}

func init() {
	cx := gctx.New()
	IPs, _ := g.Cfg().Get(cx, "cloudflareIps")
	DnsDoh.cloudflareIPRanges = IPs.Strings()
}

func (a *dnsDoh) DohHandler(r *ghttp.Request) {
	if r.Method != "POST" && r.Method != "GET" {
		r.Response.WriteStatusExit(400, "Method not allowed")
		return
	}

	var dnsQuery []byte
	var err error

	switch r.Method {
	case http.MethodGet:
		query := r.Get("dns")

		dnsQuery, err = base64.RawURLEncoding.DecodeString(query.String())
		if err != nil {
			r.Response.WriteStatusExit(http.StatusBadRequest, "Failed to decode query")
			return
		}
	case http.MethodPost:
		dnsQuery = r.GetBody()
		// if err != nil {
		// 	r.Response.WriteStatusExit(http.StatusBadRequest, "Failed to read body")
		// 	return
		// }
	}

	m := new(dns.Msg)
	err = m.Unpack(dnsQuery)
	if err != nil {
		r.Response.WriteStatusExit(http.StatusBadRequest, "Failed to unpack DNS query. "+err.Error())
		return
	}

	wMsg := new(dns.Msg)
	wMsg.SetReply(m)
	wMsg.Compress = false

	isPublicIp := true
	clientIP := getClientIP(r)
	fmt.Printf("client IP: %s", clientIP)
	if a.isCloudflareIP(clientIP) {
		isPublicIp = false
		DnsRecordApi.ParseQuery(wMsg)
	}

	if !isPublicIp && len(wMsg.Answer) > 0 {
		dnsResponse, err := wMsg.Pack()
		if err != nil {
			r.Response.WriteStatusExit(http.StatusInternalServerError, "Failed to pack DNS response")
			return
		}

		r.Response.Header().Set("Content-Type", "application/dns-message")
		r.Response.Write(dnsResponse)
	} else {
		c := new(dns.Client)
		resp, _, err := c.Exchange(m, DnsRecordApi.Forwarder)
		if err != nil {
			r.Response.WriteStatusExit(http.StatusInternalServerError, "Failed to forward DNS")
			return
		}
		dnsResponse, err := resp.Pack()
		if err != nil {
			r.Response.WriteStatusExit(http.StatusInternalServerError, "Failed to pack DNS response")
			return
		}
		r.Response.Write(dnsResponse)
	}
}

// Check if the given IP is within the Cloudflare IP ranges
func (a *dnsDoh) isCloudflareIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range a.cloudflareIPRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Error parsing CIDR %s: %v", cidr, err)
			continue
		}

		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// Get the client IP from the request, considering X-Forwarded-For header
func getClientIP(r *ghttp.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For may contain multiple IPs, use the first one
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}
