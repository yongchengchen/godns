package model

type DnsRecord struct {
	Id     uint   `gorm:"primary_key" json:"id" form:"id"`
	Domain string `json:"domain" gorm:"type:varchar(250)"`
	Ip     string `json:"ip" gorm:"type:varchar(80)"`
}
