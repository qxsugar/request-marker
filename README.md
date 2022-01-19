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
          redisRulesKey: "abc"
          redisRuleMaxLen: 256
          redisLoadInterval: 15
          redisEnable: false
          markKey: X-MARK
          headerVersion: version
          headerIdentify: identify
          cookieIdentify: identify-cookie
          query_identify: identify-query
          rules:
            - serviceName: api
              name: name
              enable: true
              priority: 100
              type: identify
              markvalue: beta
              maxVersion: 3.3.3
              minVersion: 2.2.2
              userIds:
                - A001
                - A002
              weight: 10
              path: alpha
```

使用中间件

```yaml
  routers:
    api1:
      rule: host(`api.test.cn`)
      service: svc
      entryPoints:
        - web
      middlewares:
        - request-mark

  services:
    svc:
      loadBalancer:
        servers:
          - url: "http://localhost:8999"
```

## redis 中存储的格式

```yaml
type Rule struct {
  ServiceName string   `json:"service_name"`    // 规则对应的服务名字，例：api
  Name        string   `json:"name"`            // 规则名字，例：api灰度
  Enable      bool     `json:"enable"`          // 规则开关，例：1，取值：【0,1】
  Priority    int      `json:"priority"`        // 规则优先级，例：100
  Type        RuleType `json:"type"`            // 规则类型，例：path，取值：【version, identify, weight, path】
  MarkValue   string   `json:"mark_value"`      // 规则标记值*，例：*alpha
  Version     string   `json:"version"`         // 版本，例：1.1.1-2.2.2
  UserIds     string   `json:"user_ids"`        // 用户id列表, 例：uid01,uid02
  Weight      int      `json:"weight"`          // 权重，例：30
  Path        string   `json:"path"`
}
```

## 使用

请求符合规则，将会带上mark信息。