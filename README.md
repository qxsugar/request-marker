traefik 给请求打标签中间件
--------------------------

## 流程

## 配置

配置插件

```yaml
experimental:
  plugins:
    traefik-gray-tag:
      moduleName: github.com/qxsugar/traefik-gray-tag
      version: v1.0.1
```

配置中间件

```yaml
  middlewares:
    gray-tag:
      plugin:
        traefik-gray-tag:
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