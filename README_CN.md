# Request Marker - Traefik 请求标记插件

一个为 Traefik 设计的请求标记插件，根据可配置的规则为 HTTP 请求添加标记头。支持版本范围、用户识别、金丝雀灰度和路径匹配，用于流量管理、灰度发布和 A/B 测试。

## 功能特性

- **多种规则类型**：版本范围、用户识别、金丝雀（概率性）、路径匹配
- **动态规则加载**：从 Redis 定期加载和更新规则，无需重启
- **优先级控制**：按优先级顺序评估规则，首个匹配的规则生效
- **灵活的用户识别**：支持从 HTTP 头、Cookie 或查询参数提取用户标识
- **简单日志系统**：兼容 Traefik 的受限环境，支持 DEBUG/INFO/ERROR 三个日志级别

## 安装

### 1. 配置 Traefik 插件

在 Traefik 配置文件中添加插件声明：

```yaml
experimental:
  plugins:
    request-marker:
      moduleName: github.com/qxsugar/request-marker
      version: v0.0.2
```

### 2. 配置中间件

```yaml
middlewares:
  request-marker:
    plugin:
      request-marker:
        tag: api                        # 标签，用于匹配规则
        logLevel: DEBUG                 # 日志级别：DEBUG, INFO, ERROR
        markerKey: X-MARK               # 标记头的键名
        versionHeader: X-Version        # 版本信息的头名
        identifyHeader: X-User-ID       # 用户标识的头名
        identifyCookie: user_id         # 用户标识的 Cookie 名
        identifyQuery: uid              # 用户标识的查询参数名
        
        # Redis 动态规则配置（可选）
        redisConfig:
          enable: true
          addr: redis.example.com:6379
          password: "your-password"
          db: 0
          ruleListKeys: marker:api:rules
          refreshInterval: 15
        
        # 静态规则配置（Redis 禁用时使用）
        staticRules:
          - tag: api
            name: beta-users
            enable: true
            priority: 100
            type: identify
            markValue: beta
            userIds:
              - user001
              - user002
          
          - tag: api
            name: version-range
            enable: true
            priority: 90
            type: version
            markValue: v2
            minVersion: 2.0.0
            maxVersion: 2.9.9
          
          - tag: api
            name: canary-30
            enable: true
            priority: 80
            type: canary
            markValue: canary
            canary: 30
          
          - tag: api
            name: admin-path
            enable: true
            priority: 70
            type: path
            markValue: admin
            path: /admin
```

### 3. 在路由中使用中间件

```yaml
routers:
  api-router:
    rule: host(`api.example.com`)
    service: api-service
    entryPoints:
      - web
    middlewares:
      - request-marker

services:
  api-service:
    loadBalancer:
      servers:
        - url: "http://backend-1:8080"
        - url: "http://backend-2:8080"
```

## 规则类型详解

### 1. 版本规则 (version)

根据请求的版本号范围进行匹配。版本号采用语义化版本格式（如 `1.2.3`）。

```yaml
type: version
minVersion: 2.0.0
maxVersion: 2.9.9
markValue: v2-stable
```

### 2. 用户识别规则 (identify)

根据用户标识列表进行精确匹配。用户标识从请求头、Cookie 或查询参数中提取。

```yaml
type: identify
userIds:
  - user001
  - user002
markValue: beta-tester
```

### 3. 金丝雀规则 (canary)

基于用户标识的哈希值进行概率性匹配。金丝雀值范围 0-100，表示匹配的百分比。

```yaml
type: canary
canary: 30
markValue: canary
```

### 4. 路径规则 (path)

根据请求 URL 中的路径子串进行匹配。

```yaml
type: path
path: /admin
markValue: admin-panel
```

## Redis 规则存储格式

规则存储在 Redis 中作为哈希表。使用分层 key 结构，按标签组织。

### Key 结构

```
{ruleListKeys}                    # 规则索引（List），存储所有规则 key
marker:api:rule:{ruleName}        # 单条规则（Hash）
```

示例（ruleListKeys = `marker:api:rules`）：
```
marker:api:rules                  → List ["marker:api:rule:beta-users", ...]
marker:api:rule:beta-users        → Hash {name, enable, priority, ...}
```

### 规则哈希表格式

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 规则名称 |
| `enable` | 0/1 | 是否启用 |
| `priority` | int | 优先级（数值越大优先级越高） |
| `type` | string | 规则类型：version/identify/canary/path |
| `mark_value` | string | 匹配时设置的标记值 |
| `min_version` | string | 最小版本（version 类型） |
| `max_version` | string | 最大版本（version 类型） |
| `user_ids` | string | 用户 ID 列表（逗号分隔） |
| `canary` | int | 金丝雀百分比 0-100 |
| `path` | string | 路径子串 |

## 开发

### 构建和测试

```bash
make lint      # 运行代码检查
make test      # 运行测试
make vendor    # 生成 vendor 目录
make clean     # 清理 vendor 目录
```

### 运行单个测试

```bash
go test -v -run TestName ./...
```

## 工作原理

1. **初始化**：插件启动时加载静态规则并按优先级排序
2. **Redis 刷新**：如果启用 Redis，后台定期从 Redis 加载最新规则
3. **请求处理**：对每个请求，按优先级顺序评估规则
4. **规则匹配**：
   - 检查规则是否启用
   - 检查标签是否匹配
   - 根据规则类型执行相应的匹配逻辑
5. **标记设置**：首个匹配的规则将其 `markValue` 设置到请求头中
6. **转发**：带有标记的请求转发到后端服务

## 用户识别优先级

用户标识按以下优先级从请求中提取：

1. HTTP 头（`identifyHeader` 配置）
2. Cookie（`identifyCookie` 配置）
3. 查询参数（`identifyQuery` 配置）

如果三个位置都未找到用户标识，金丝雀规则将无法匹配。

## 许可证

MIT
