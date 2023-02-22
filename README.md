# Cookie Path Prefixer

Cookie Path Prefixer is a middleware plugin for [Traefik](https://traefik.io) which adds a prefix to the path of a cookie in the response. If no path is defined in the cookie, a new path constructed from the prefix will be set.

[![Build Status](https://github.com/SchmitzDan/traefik-plugin-cookie-path-prefix/workflows/Main/badge.svg?branch=main)](https://github.com/SchmitzDan/traefik-plugin-cookie-path-prefix/actions)

## Configuration

### Static

```yaml
experimental:
  plugins:
    cookiePathPrefix:
      modulename: "github.com/SchmitzDan/traefik-plugin-cookie-path-prefix"
      version: "v0.0.3" #replace with newest version
```

### Dynamic

To configure the  plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). 
The following example creates and uses the cookie path prefix middleware plugin to add the prefix "/foo" to the cookie paths:

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "cookiePathPrefix"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    cookiePathPrefix:
      plugin:
        cookiePathPrefix:
          prefix: "foo"
```

Configuration can also be set via toml or docker labels.
