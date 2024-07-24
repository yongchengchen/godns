package api

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/sirupsen/logrus"
	"github.com/yongchengchen/godns/app/model"
	"github.com/yongchengchen/godns/library/response"

	"github.com/miekg/dns"

	"strings"
)

// 用户API管理对象
var DnsRecordApi = new(dnsRecordsApi)

type dnsRecordsApi struct {
	Forwarder string
}

// redis Get command
func (a *dnsRecordsApi) GetRecords(r *ghttp.Request) {
	var records []model.DnsRecord
	err := g.DB("sqlite").Model("dns_records").Scan(&records)

	if err != nil {
		response.JsonExit(r, 500, err.Error())
	}
	response.JsonExit(r, 200, "success", records)
}

func (a *dnsRecordsApi) ListRecords() {
	var records []model.DnsRecord
	err := g.DB("sqlite").Model("dns_records").Scan(&records)

	if err != nil {
		logrus.Println("local dns record is empty.")
	}

	logrus.Println("local dns records:")
	for i := 1; i < len(records); i++ {
		logrus.Printf("  %s. %s\n", records[i].Domain, records[i].Ip)
	}
}

func (a *dnsRecordsApi) GetRecord(r *ghttp.Request) {
	var (
		item *model.DnsRecord
	)
	var id = r.Get("id")

	if err := g.DB("sqlite").Model("dns_records").Where("id", id).Scan(&item); err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	ret := item.Domain + "(" + item.Ip + ")"

	response.JsonExit(r, 200, "success", ret)
}

func (a *dnsRecordsApi) InsertRecord(r *ghttp.Request) {
	var (
		data *model.DnsRecord
	)
	if err := r.Parse(&data); err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	v, err := g.DB("sqlite").GetValue(gctx.New(), "select id from dns_records order by id desc limit 1")
	if err != nil {
		response.JsonExit(r, 400, err.Error())
	}
	data.Id = 1 + v.Uint()
	ret, err := g.DB("sqlite").Model("dns_records").Insert(data)
	if err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	response.JsonExit(r, 200, "success", ret)
}

func (a *dnsRecordsApi) UpdateRecord(r *ghttp.Request) {
	var (
		data *model.DnsRecord
	)
	if err := r.Parse(&data); err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	ret, err := g.DB("sqlite").Model("dns_records").Update(data, "id", data.Id)
	if err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	response.JsonExit(r, 200, "success", ret)
}

func (a *dnsRecordsApi) DeleteRecord(r *ghttp.Request) {
	var (
		data *model.DnsRecord
	)
	if err := r.Parse(&data); err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	ret, err := g.DB("sqlite").Model("dns_records").Delete("id", data.Id)
	if err != nil {
		response.JsonExit(r, 400, err.Error())
	}

	response.JsonExit(r, 200, "success", ret)
}

func getRootDomain(domain string) string {
	parts := strings.Split(domain, ".")
	length := len(parts)

	if length < 2 {
		return domain // Return the domain itself if it's less than two parts
	}

	// Ensure the root domain is at least two levels
	rootDomain := parts[length-2] + "." + parts[length-1]
	return rootDomain
}

func (h *dnsRecordsApi) ParseQuery(m *dns.Msg) {
	//type A only
	var (
		item *model.DnsRecord
	)
	var records []model.DnsRecord
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			domain := strings.TrimSuffix(q.Name, ".")
			rootDomain := getRootDomain(domain)
			testDomains := []string{
				domain,
				"*." + rootDomain,
			}
			if err := g.DB("sqlite").Model("dns_records").WhereIn("domain", testDomains).Scan(&records); err != nil {
				continue
			}

			if len(records) > 0 {
				item = &records[0]
				for i := 1; i < len(records); i++ {
					if records[i].Domain == domain {
						item = &records[i]
					}
				}
			}

			if item == nil {
				continue
			}

			ip := item.Ip

			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
				// rr := &dns.A{
				//     Hdr: dns.RR_Header{
				//         Name:   q.Name,
				//         Rrtype: dns.TypeA,
				//         Class:  dns.ClassINET,
				//         Ttl:    600,
				//     },
				//     A: net.ParseIP(ip),
				// }
				// msg.Answer = append(msg.Answer, rr)
			}
		}
	}
}

func (h *dnsRecordsApi) forwardRequest(w dns.ResponseWriter, r *dns.Msg) {
	c := new(dns.Client)
	resp, _, err := c.Exchange(r, h.Forwarder)
	if err != nil {
		return
	}
	w.WriteMsg(resp)
}

func (h *dnsRecordsApi) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		h.ParseQuery(m)
	}

	if len(m.Answer) > 0 {
		w.WriteMsg(m)
	} else {
		h.forwardRequest(w, r)
	}
}
