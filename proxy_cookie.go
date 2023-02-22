// Package traefik_plugin_proxy_cookie a traefik plugin providing the functionality of the nginx proxy_cookie directives tp traefik.
package traefik_plugin_proxy_cookie //nolint

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

const setCookieHeader string = "Set-Cookie"

// Rewrite definition of a replacement.
type Rewrite struct {
	Regex       string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
}

type rewrite struct {
	regex       *regexp.Regexp
	replacement string
}

type domainConfig struct {
	Rewrites []Rewrite `json:"rewrites,omitempty" toml:"rewrites,omitempty" yaml:"rewrites,omitempty"`
}

type pathConfig struct {
	Prefix   string    `json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty"`
	Rewrites []Rewrite `json:"rewrites,omitempty" toml:"rewrites,omitempty" yaml:"rewrites,omitempty"`
}

// Config holding the prefix to add.
type Config struct {
	PathConfig   pathConfig   `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty"`
	DomainConfig domainConfig `json:"domain,omitempty" toml:"domain,omitempty" yaml:"domain,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// ProxieCookiePlugin a traefik plugin providing the functionality of the nginx proxy_cookie directives tp traefik.
type ProxieCookiePlugin struct {
	next           http.Handler
	name           string
	domainRewrites []rewrite
	pathPrefix     string
	pathRewrites   []rewrite
}

// New creates a Path Prefixer.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	domainRewrites, err := convertRewrites(config.DomainConfig.Rewrites)
	if err != nil {
		return nil, err
	}

	pathRewrites, err := convertRewrites(config.PathConfig.Rewrites)
	if err != nil {
		return nil, err
	}

	return &ProxieCookiePlugin{
		name:           name,
		next:           next,
		domainRewrites: domainRewrites,
		pathPrefix:     config.PathConfig.Prefix,
		pathRewrites:   pathRewrites,
	}, nil
}

func convertRewrites(rewriteConfigs []Rewrite) ([]rewrite, error) {
	rewrites := make([]rewrite, len(rewriteConfigs))

	for i, rewriteConfig := range rewriteConfigs {
		regexp, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}
		rewrites[i] = rewrite{
			regex:       regexp,
			replacement: rewriteConfig.Replacement,
		}
	}
	return rewrites, nil
}

func (p *ProxieCookiePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	myWriter := &responseWriter{
		writer:         rw,
		domainRewrites: p.domainRewrites,
		pathPrefix:     p.pathPrefix,
		pathRewrites:   p.pathRewrites,
	}

	p.next.ServeHTTP(myWriter, req)
}

type responseWriter struct {
	writer         http.ResponseWriter
	domainRewrites []rewrite
	pathPrefix     string
	pathRewrites   []rewrite
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	// workaround to get the cookies
	headers := r.writer.Header()
	req := http.Response{Header: headers}
	cookies := req.Cookies()

	// Delete set-cookie headers
	r.writer.Header().Del(setCookieHeader)

	// Add new cookie with modified path and domain
	for _, cookie := range cookies {
		// add the prefix if any defined
		if len(r.pathPrefix) > 0 {
			cookie.Path = prefixPath(cookie.Path, r.pathPrefix)
		}
		// rewrite the path
		if len(r.pathRewrites) > 0 {
			cookie.Path = handleRewrites(cookie.Path, r.pathRewrites)
		}
		// rewrite the domain
		if len(r.domainRewrites) > 0 {
			cookie.Domain = handleRewrites(cookie.Domain, r.domainRewrites)
		}

		http.SetCookie(r, cookie)
	}

	r.writer.WriteHeader(statusCode)
}

func prefixPath(path, prefix string) string {
	if path == "/" {
		// prevent trailing /
		return "/" + prefix
	}
	return "/" + prefix + path
}

func handleRewrites(value string, rewrites []rewrite) string {
	for _, rewrite := range rewrites {
		value = rewrite.regex.ReplaceAllString(value, rewrite.replacement)
	}
	return value
}
