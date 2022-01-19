mark request for traefik plugin
-------------------------------

## 配置

配置插件

```yaml
experimental:
  plugins:
    request-mark:
      moduleName: github.com/qxsugar/request-mark
      version: v0.0.1
```

配置中间件

```yaml
  middlewares:
    request-mark:
      plugin:
        request-mark:
          serviceName: api
          logLevel: DEBUG
          redisAddr: redis.com
          redisPassword: "***"
          redisEnable: false
          redisRulesKey: "abc"
          redisRuleMaxLen: 256
          redisLoadInterval: 15
          tagKey: TAG
          headerVersion: X-version
          headerIdentify: identify
          cookieIdentify: identify-cookie
          query_identify: identify-query
          rules:
            - serviceName: api
              name: name
              enable: true
              priority: 100
              type: path
              tagValue: alpha
              maxVersion: 3.3.3
              minVersion: 2.2.2
              userIds:
                - A001
                - A002
              weight: 80
              path: alpha
```