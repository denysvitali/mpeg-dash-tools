package mpeg_dash_tools

import "net/http"

type CustomRoundTripper struct {
	embeddedRoundTripper http.RoundTripper
}

const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0"

func (c CustomRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request != nil {
		request.Header.Add("Accept", "*/*")
		request.Header.Add("User-Agent", UserAgent)
	}
	return c.embeddedRoundTripper.RoundTrip(request)
}

var _ http.RoundTripper = (*CustomRoundTripper)(nil)
