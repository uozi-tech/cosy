package geoip

import (
	"bytes"
	"embed"
	"io"
	"log"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
	"github.com/ulikunitz/xz"
	"github.com/uozi-tech/cosy/logger"
)

//go:embed GeoLite2-Country.mmdb.xz
var fs embed.FS

var db *geoip2.Reader

func init() {
	compressedBytes, err := fs.ReadFile("GeoLite2-Country.mmdb.xz")
	if err != nil {
		log.Fatal(err)
	}

	reader, err := xz.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		log.Fatal(err)
	}

	dbBytes, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	db, err = geoip2.OpenBytes(dbBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func ParseIP(input string) string {
	ip, err := netip.ParseAddr(input)
	if err != nil {
		logger.Error(err)
		return "Unknown"
	}
	record, err := db.Country(ip)
	if err != nil {
		logger.Error(err)
		return "Unknown"
	}

	return record.Country.ISOCode
}
