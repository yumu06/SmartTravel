# 小程序本地配置说明

仓库中的个人标识和第三方密钥均已清空，使用前需要自行填写。

## 微信 AppID

微信项目配置文件 `project.config.json` 的 `appid` 当前为空。导入微信开发者工具后，可以在项目设置中填写自己的小程序 AppID；请勿提交包含个人 AppID 的私有配置。

## 后端 API 地址

在 `config/index.js` 中修改 `API_BASE_URL`：

- 开发者工具可使用 `http://127.0.0.1:1016`；
- 真机调试填写电脑的局域网地址；
- 正式发布填写微信公众平台已配置的 HTTPS 合法域名。

## 地图与微信密钥

腾讯地图 WebService Key、微信 AppSecret 等服务端凭据不得写入小程序代码。请复制后端的 `travel/config/environment.example.ps1`，在本地私有文件中填写，并通过环境变量传给后端。
