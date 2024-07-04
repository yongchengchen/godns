package router

import (
	"io/ioutil"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/yongchengchen/godns/app/api"
)

func init() {
	s := g.Server()

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.GET("/dns-records", api.DnsRecordApi.GetRecords)
		group.GET("/dns-records/:id", api.DnsRecordApi.GetRecord)
		group.POST("/dns-records", api.DnsRecordApi.InsertRecord)
		group.POST("/dns-records/:id", api.DnsRecordApi.UpdateRecord)
		group.DELETE("/dns-records/:id", api.DnsRecordApi.DeleteRecord)
	})

	// path := gfile.MainPkgPath() + "/dist"
	path := "./dist"

	s.BindStatusHandler(404, func(r *ghttp.Request) {
		// r.Response.w
		file := path + "/index.html"
		c, err := ioutil.ReadFile(file)
		if err != nil {
			r.Response.WriteStatus(404, file+"Not Found")
		}
		r.Response.WriteStatus(200, c)
	})

	// logrus.Println(path)
	s.SetServerRoot(path)
	s.SetPort(8299)
}
