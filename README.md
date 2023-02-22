# Proxy Cookie Plugin

Proxy Cookie Plugin is a middleware plugin for [Traefik](https://traefik.io) which adds the functionylity of the proxy_cookie directives known from NGINX to Traefik.


## Configuration

### Static

```yaml
experimental:
  plugins:
    proxyCookie:
      modulename: "github.com/SchmitzDan/traefik-plugin-proxy-cookie"
      version: "v0.0.x" #replace with newest version
```

### Dynamic

To configure the  plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). 
The following example creates and uses the plugin to add the prefix "/foo" to the cookie paths and change the cookie domain from "subdomain.foo.*" to "foo.*":

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "proxyCookie"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    proxyCookie:
      plugin:
        proxyCookie:
          domain:
            rewrites:
              - regex: "^subdomain.foo.(.+)$"
                replacement: "foo.$1"
          path:
            prefix: "foo"
```

Configuration can also be set via toml or docker labels.

You can also define a set of rewrites for the path. If both, a prefix and one or more rewrites is defined for the path, the path will be prefixed first and afterwords the rewrites are applied to the path.
