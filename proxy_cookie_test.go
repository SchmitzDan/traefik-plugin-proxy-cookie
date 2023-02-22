// Package traefik_plugin_proxy_cookie a traefik plugin providing the functionality of the nginx proxy_cookie directives tp traefik.
package traefik_plugin_proxy_cookie //nolint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHttp(t *testing.T) {
	tests := []struct {
		desc           string
		domainRewrites []Rewrite
		pathRewrites   []Rewrite
		pathPrefix     string
		reqHeader      http.Header
		expRespHeader  http.Header
	}{
		{
			desc:       "Add foo as prefix to one cookie Path",
			pathPrefix: "foo",
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc:       "Add foo as prefix to two cookies Paths",
			pathPrefix: "foo",
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/", "someOtherName=someValue; Path=/bar"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo", "someOtherName=someValue; Path=/foo/bar"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc:       "Add foo as prefix to no cookie",
			pathPrefix: "foo",
			reqHeader: map[string][]string{
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc:       "Add foo as prefix to cookie without Path",
			pathPrefix: "foo",
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo"},
				"anotherHeader": {"Path=/"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			pathConfig := pathConfig{
				Prefix:   test.pathPrefix,
				Rewrites: test.pathRewrites,
			}

			domainConfig := domainConfig{
				Rewrites: test.domainRewrites,
			}

			config := &Config{
				PathConfig:   pathConfig,
				DomainConfig: domainConfig,
			}

			next := func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range test.reqHeader {
					for _, h := range v {
						rw.Header().Add(k, h)
					}
				}
				rw.WriteHeader(http.StatusOK)
			}

			pathPrefix, err := New(context.Background(), http.HandlerFunc(next), config, "prefixCookiePath")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			pathPrefix.ServeHTTP(recorder, req)
			for k, expected := range test.expRespHeader {
				values := recorder.Header().Values(k)

				if !testEq(values, expected) {
					t.Errorf("Unexpected Header: expected %+v, result: %+v", expected, values)
				}
			}
		})
	}
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
