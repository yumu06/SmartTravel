# SmartTravel（随心游）

SmartTravel 是一个基于 Go 和微信小程序实现的智能旅行规划与内容社区项目。项目围绕路线规划、旅行记录和游记互动构建业务闭环，并对认证安全、数据库访问、核心查询和工程配置进行了升级。

## 功能概览

### 用户与认证

- 微信小程序登录和用户自动注册；
- Access Token + Refresh Token 双令牌；
- Redis Refresh Session；
- Refresh Token 原子轮换和重放检测；
- 主动注销和 Access Token 黑名单；
- 用户资料查询与修改。

### 旅行规划

- 出发地、目的地和途经点输入；
- 驾车等出行方式的路线规划；
- 路线结果保存为旅行足迹；
- 足迹查询、收藏和删除；
- 腾讯地图 WebService 由后端调用，Key 不下发到小程序。

### 内容社区

- 文章发布、编辑、删除和详情展示；
- 文章分页、搜索和推荐；
- 评论发布、列表和删除；
- 点赞、取消点赞、收藏和取消收藏；
- 用户文章和收藏列表。

### 其他功能

- 用户聊天；
- 公告列表、详情和管理接口；
- 地点与心情推荐基础能力。

## 技术栈

| 类型 | 技术 |
| --- | --- |
| 后端 | Go 1.20、Gin、GORM |
| 数据库 | MySQL 8 |
| 缓存与会话 | Redis、Lua |
| 认证 | JWT v5、Access Token、Refresh Token |
| 小程序 | 微信小程序原生 JavaScript、WXML、WXSS |
| 地图服务 | 腾讯地图 WebService API |
| 配置 | Viper、环境变量 |
| 测试 | Go testing、miniredis、httptest |

## 项目结构

```text
SmartTravel/
├─ miniprogram-5/                 微信小程序
│  ├─ common/                     公共组件与请求工具
│  ├─ config/                     API 地址和配置说明
│  ├─ packages/                   小程序分包页面
│  ├─ pages/                      主包页面
│  ├─ app.js                      小程序入口
│  └─ project.config.json         微信开发者工具配置
├─ travel/                        Go 后端
│  ├─ auth/                       会话与认证服务
│  ├─ benchmark/                  压测工具、数据生成和报告
│  ├─ cache/                      Redis 客户端
│  ├─ config/                     后端配置与环境模板
│  ├─ controller/                 HTTP 控制器
│  ├─ logic/                      业务逻辑
│  ├─ middleware/                 鉴权、CORS 和恢复中间件
│  ├─ pkg/jwt/                    JWT 签发与解析
│  ├─ router/                     Gin 路由
│  ├─ TravelDate/                 数据访问层
│  └─ TravelModel/                数据模型
├─ 项目升级功能说明.md             升级内容和完成状态
└─ LICENSE                        GPL-3.0 许可证
```

## 认证流程

后端使用两种不同用途的 Token：

1. 小程序通过 `wx.login` 获取临时 Code；
2. 后端使用 Code 换取 OpenID 和 SessionKey；
3. 登录成功后签发短期 Access Token 和长期 Refresh Token；
4. 业务接口通过 `Authorization: Bearer <access-token>` 认证；
5. Access Token 过期后可调用 `/travel/auth/refresh`；
6. 注销时删除 Redis Session，并将 Access Token 加入黑名单。

当前后端双 Token 能力已经完成；小程序端尚未接入自动刷新、并发刷新合并和原请求重试。Access Token 失效后，当前小程序会重新登录。

## 数据库与性能优化

后端提供以下数据库配置：

- 最大打开连接数和最大空闲连接数；
- 连接生命周期和空闲生命周期；
- 慢 SQL 阈值；
- 用户、文章、评论、足迹和聊天索引；
- 点赞与收藏唯一关系约束。

推荐查询使用复合降序索引：

```text
idx_post_recommend(like_count DESC, view_count DESC, created_at DESC)
```

在本机、10,000 篇文章、200 并发的对比测试中：

| 指标 | 优化前 | 优化后 |
| --- | ---: | ---: |
| 成功 QPS | 471.27 | 7,020.93 |
| P95 | 1,144.69ms | 43.41ms |
| P99 | 1,482.10ms | 82.37ms |
| 错误率 | 0% | 0% |

这些数据用于本地版本对比，不代表生产服务器容量。完整结果见 [核心业务性能报告](travel/benchmark/2026-08-22-core-business-baseline.md)。

## 隐私与配置

仓库中的以下配置均为空，不包含个人信息：

- 微信 AppID 和 AppSecret；
- 腾讯地图 WebService Key；
- MySQL 和 Redis 密码；
- JWT Access/Refresh Secret；
- 管理员 Token；
- 个人域名和 OpenID。

### 后端配置

复制环境变量模板：

```powershell
cd travel
Copy-Item .\config\environment.example.ps1 .\config\environment.local.ps1
```

编辑 `environment.local.ps1`，填写本地 MySQL、Redis、JWT、微信和地图服务配置，然后加载：

```powershell
. .\config\environment.local.ps1
go run .
```

本地文件 `environment.local.ps1` 已加入 `.gitignore`，不要将真实凭据写入 `application.yml` 或提交到仓库。

### 小程序配置

1. 在微信开发者工具中导入 `miniprogram-5`；
2. 在项目设置中填写自己的小程序 AppID；
3. 修改 `miniprogram-5/config/index.js` 中的 `API_BASE_URL`；
4. 真机调试填写电脑局域网地址；
5. 正式发布使用微信公众平台已配置的 HTTPS 合法域名。

地图 Key 和微信 AppSecret 只能配置在后端，不应写入小程序代码。

## 启动后端

### 环境要求

- Go 1.20 或更高版本；
- MySQL 8；
- Redis；
- 已填写的本地环境变量文件。

### 启动命令

```powershell
cd travel
. .\config\environment.local.ps1
go mod download
go run .
```

默认监听地址为：

```text
http://127.0.0.1:1016
```

## 主要接口

### 无需登录

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/travel/login` | 微信登录 |
| POST | `/travel/auth/refresh` | 刷新 Token |
| GET | `/travel/notice/list` | 公告列表 |
| GET | `/travel/notice/show/:id` | 公告详情 |

### 需要 Access Token

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/travel/authorization` | 检查授权状态 |
| POST | `/travel/auth/logout` | 主动注销 |
| GET | `/travel/user/info` | 用户信息 |
| POST | `/travel/planRoute` | 路线规划 |
| GET | `/travel/post/show/:id` | 文章详情 |
| GET | `/travel/post/page/list` | 文章分页 |
| GET | `/travel/post/recommand` | 文章推荐 |
| POST | `/travel/post/like/:id` | 点赞文章 |
| POST | `/travel/post/:id/comment` | 发布评论 |
| GET | `/travel/foot/list` | 足迹列表 |
| GET | `/travel/user/chat` | 聊天记录 |

完整接口说明见 [接口文档](travel/documents/接口文档.md)。

## 测试与构建

在 `travel` 目录执行：

```powershell
go test -count=1 ./...
go vet ./...
go build ./...
```

测试覆盖 JWT、Redis Session、刷新令牌轮换、数据库连接池、模型索引、环境变量覆盖、基准数据和压测工具。

## 独立压测环境

项目提供 `travel_benchmark` 隔离数据库，基准数据包括：

| 数据类型 | 数量 |
| --- | ---: |
| 用户 | 1,000 |
| 文章 | 10,000 |
| 评论 | 50,000 |
| 点赞关系 | 100,000 |
| 收藏关系 | 30,000 |
| 旅行足迹 | 10,000 |

压测工具支持 Bearer Token、多 URL 轮询、错误分类、成功 QPS 和 P50/P95/P99 延迟统计。压测数据库只能使用固定名称 `travel_benchmark`，不会清空业务数据库 `travel_database`。

## 当前未完成

- 小程序自动刷新 Token 和请求重试；
- 鉴权用户缓存的生产接入及失效策略；
- 浏览量 Redis 聚合和批量回写；
- 推荐结果缓存；
- Docker、Kubernetes、RabbitMQ、Elasticsearch；
- 生产域名、HTTPS、监控和告警。

详细升级状态见 [项目升级功能说明](项目升级功能说明.md)。

## 许可证

本项目使用 [GNU General Public License v3.0](LICENSE)。
