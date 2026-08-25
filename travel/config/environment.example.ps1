# 复制为 environment.local.ps1，填写本机值后执行：
# . .\config\environment.local.ps1
# go run .

# 服务
$env:SERVER_PORT = '1016'
$env:SERVER_ACCESS_LOG = 'true'

# MySQL
$env:MYSQL_HOST = '127.0.0.1'
$env:MYSQL_PORT = '3306'
$env:MYSQL_ROOT = 'root'
# 必填：填写本机 MySQL 密码
$env:MYSQL_PASSWORD = ''
$env:MYSQL_DATABASE = 'travel_database'

# Redis
$env:REDIS_ADDR = '127.0.0.1:6379'
# Redis 未设置密码时可保持为空
$env:REDIS_PASSWORD = ''
$env:REDIS_DB = '0'

# JWT：生产环境请分别使用两个独立、随机且不少于 32 字节的值
$env:JWT_ACCESS_SECRET = '' # 必填：Access Token 签名密钥
$env:JWT_REFRESH_SECRET = '' # 必填：Refresh Token 签名密钥
$env:JWT_ISSUER = 'ongoing-trip'

# 微信小程序
$env:WX_APPID = '' # 必填：微信小程序 AppID
$env:WX_APPSECRET = '' # 必填：微信小程序 AppSecret

# 腾讯地图 WebService
$env:TENCENT_WEBSERVICE_API = '' # 使用路线规划时必填：腾讯地图 WebService Key

# 可选配置
$env:BAIDU_AK = ''
$env:BAIDU_REGION = ''
$env:ADMIN_TOKEN = '' # 使用管理员接口时填写
