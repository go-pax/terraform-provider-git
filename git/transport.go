package git

import "net/http"

const (
	ctxEtag = ctxEtagType("etag")
)

// ctxEtagType is used to avoid collisions between packages using context
type ctxEtagType string

// etagTransport allows saving API quota by passing previously stored Etag
// available via context to request headers
type etagTransport struct {
	transport http.RoundTripper
}

func (e etagTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()

	etag := ctx.Value(ctxEtag)
	if v, ok := etag.(string); ok && v != "" {
		request.Header.Set("If-None-Match", v)
	}

	return e.transport.RoundTrip(request)
}

func NewEtagTransport(rt http.RoundTripper) *etagTransport {
	return &etagTransport{transport: rt}
}
