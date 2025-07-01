package geoip

import (
	"embed"
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/uozi-tech/cosy/logger"
)

//go:embed GeoLite2-Country.mmdb
var fs embed.FS

var db *geoip2.Reader

func init() {
	dbBytes, err := fs.ReadFile("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	db, err = geoip2.FromBytes(dbBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func ParseIP(input string) string {
	ip := net.ParseIP(input)
	record, err := db.Country(ip)
	if err != nil {
		logger.Error(err)
		return "Unknown"
	}

	return record.Country.IsoCode
}
