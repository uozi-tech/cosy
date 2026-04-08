package sls

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	headerContentMD5    = "Content-MD5"
	headerContentType   = "Content-Type"
	headerDate          = "Date"
	headerAuthorization = "Authorization"
	headerUserAgent     = "User-Agent"
	headerAPIVersion    = "x-log-apiversion"
	headerSignMethod    = "x-log-signaturemethod"
	headerBodyRawSize   = "x-log-bodyrawsize"
	headerCompressType  = "x-log-compresstype"

	apiVersion      = "0.6.0"
	signatureMethod = "hmac-sha1"
	userAgent       = "cosy-sls-sdk"
)

var gmtLoc = time.FixedZone("GMT", 0)

// sign computes the SLS V1 signature and populates Authorization and related headers.
func sign(method, uri string, headers map[string]string, body []byte, creds Credentials) {
	// Content-MD5
	if len(body) > 0 {
		headers[headerContentMD5] = fmt.Sprintf("%X", md5.Sum(body))
	}

	// Date
	if _, ok := headers[headerDate]; !ok {
		headers[headerDate] = time.Now().In(gmtLoc).Format(time.RFC1123)
	}

	headers[headerSignMethod] = signatureMethod
	headers[headerAPIVersion] = apiVersion
	headers[headerUserAgent] = userAgent

	// CanonicalizedSLSHeaders
	var keys []string
	slsHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		lower := strings.ToLower(strings.TrimSpace(k))
		if strings.HasPrefix(lower, "x-log-") || strings.HasPrefix(lower, "x-acs-") {
			slsHeaders[lower] = strings.TrimSpace(v)
			keys = append(keys, lower)
		}
	}
	sort.Strings(keys)
	var canoHeaders strings.Builder
	for i, k := range keys {
		if i > 0 {
			canoHeaders.WriteByte('\n')
		}
		canoHeaders.WriteString(k)
		canoHeaders.WriteByte(':')
		canoHeaders.WriteString(slsHeaders[k])
	}

	// CanonicalizedResource
	u, _ := url.Parse(uri)
	canoResource := u.EscapedPath()
	if u.RawQuery != "" {
		vals := u.Query()
		var qkeys []string
		for k := range vals {
			qkeys = append(qkeys, k)
		}
		sort.Strings(qkeys)
		canoResource += "?"
		for i, k := range qkeys {
			if i > 0 {
				canoResource += "&"
			}
			for _, v := range vals[k] {
				canoResource += k + "=" + v
			}
		}
	}

	contentType, _ := headers[headerContentType]
	contentMD5, _ := headers[headerContentMD5]

	signStr := method + "\n" +
		contentMD5 + "\n" +
		contentType + "\n" +
		headers[headerDate] + "\n" +
		canoHeaders.String() + "\n" +
		canoResource

	mac := hmac.New(sha1.New, []byte(creds.AccessKeySecret))
	mac.Write([]byte(signStr))
	digest := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	headers[headerAuthorization] = fmt.Sprintf("LOG %s:%s", creds.AccessKeyID, digest)
}
