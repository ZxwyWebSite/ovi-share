package share

const (
	ctKey = `Content-Type`
	uaKey = `User-Agent`
)

var (
	ctJsonOdata = []string{`application/json;odata=verbose`}
	uaFireFox   = []string{`Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0`}
	prefer      = []string{`autoredeem`}
)
