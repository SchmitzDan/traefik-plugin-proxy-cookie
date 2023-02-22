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
		{
			desc: "Replace foo by bar in path",
			pathRewrites: []Rewrite{
				{
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/bar"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Replace foo by bar in not present path",
			pathRewrites: []Rewrite{
				{
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=www.foo.com"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=www.foo.com"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Remove leading subpath",
			pathRewrites: []Rewrite{
				{
					Regex:       "^/bar/foo(.+)$",
					Replacement: "/foo$1",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/bar/foo/something"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo/something"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Add leading subpath",
			pathRewrites: []Rewrite{
				{
					Regex:       "^/foo(.+)$",
					Replacement: "/bar/foo$1",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/foo/something"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Path=/bar/foo/something"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Replace foo by bar in domain",
			domainRewrites: []Rewrite{
				{
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=www.foo.com"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=www.bar.com"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Replace foo by bar in not present domain",
			domainRewrites: []Rewrite{
				{
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue"},
				"anotherHeader": {"Path=/"},
			},
		},
		{
			desc: "Remove subdomain",
			domainRewrites: []Rewrite{
				{
					Regex:       "^subdomain.foo.(.+)$",
					Replacement: "foo.$1",
				},
			},
			reqHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=subdomain.foo.bar"},
				"anotherHeader": {"Path=/"},
			},
			expRespHeader: map[string][]string{
				"set-cookie":    {"someName=someValue; Domain=foo.bar"},
				"anotherHeader": {"Path=/"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{
				PathConfig: pathConfig{
					Prefix:   test.pathPrefix,
					Rewrites: test.pathRewrites,
				},
				DomainConfig: domainConfig{
					Rewrites: test.domainRewrites,
				},
			}

			next := func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range test.reqHeader {
					for _, h := range v {
						rw.Header().Add(k, h)
					}
				}
				rw.WriteHeader(http.StatusOK)
			}

			proxyCookiePlugin, err := New(context.Background(), http.HandlerFunc(next), config, "proxyCookie")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			proxyCookiePlugin.ServeHTTP(recorder, req)
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
