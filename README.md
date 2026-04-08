# GH Proxy Auth

基于 [gh-proxy](https://github.com/hunshcn/gh-proxy) 的 Golang 实现版本，增加了用户认证和 Token 管理功能。

## 功能特性

- **GitHub 代理加速**：支持 GitHub releases、archive、blob/raw、raw.githubusercontent.com、gist、tags 以及 git clone 操作的代理加速
- **Token 认证**：通过 `X-XN-Token` Header 或 URL 参数传递 Token 进行身份验证
- **用户管理**：支持用户注册、登录，首位注册用户自动成为管理员
- **安全设置**：支持 TOTP 验证器和 Passkey（WebAuthn）双因素认证
- **Token 管理**：创建、管理访问 Token，支持设置有效期（小时/天/永不过期）
- **下载记录**：记录所有通过代理下载的日志
- **管理后台**：管理员可控制开放注册、查看所有用户和下载记录
- **单文件部署**：前端资源嵌入 Go 二进制文件，一个文件即可部署
- **多数据库支持**：支持 SQLite、MySQL、PostgreSQL

## 快速开始

### 下载

从 [Releases](../../releases) 页面下载最新的 `gh-proxy-auth-linux-amd64.tar.gz`。

```bash
tar -xzf gh-proxy-auth-linux-amd64.tar.gz
```

### 配置

复制配置文件并修改：

```bash
cp config.example.yaml config.yaml
```

配置文件说明：

```yaml
server:
  host: "0.0.0.0"        # 监听地址
  port: 8080              # 监听端口

database:
  driver: "sqlite"        # 数据库类型: sqlite, mysql, postgres
  dsn: "data.db"          # 数据库连接字符串

jwt:
  secret: "your-secret"   # JWT 签名密钥，请修改为随机字符串
  expire: "24h"           # JWT 过期时间

domain: "https://proxy.example.com"  # 服务域名（用于 WebAuthn/Passkey）

proxy:
  size_limit: 1073741824  # 代理文件大小限制（字节），默认 1GB
  jsdelivr: false          # 是否使用 jsDelivr 镜像
```

#### 数据库配置示例

**SQLite（默认）：**
```yaml
database:
  driver: "sqlite"
  dsn: "data.db"
```

**MySQL：**
```yaml
database:
  driver: "mysql"
  dsn: "user:password@tcp(127.0.0.1:3306)/gh_proxy?charset=utf8mb4&parseTime=True&loc=Local"
```

**PostgreSQL：**
```yaml
database:
  driver: "postgres"
  dsn: "host=127.0.0.1 user=user password=password dbname=gh_proxy port=5432 sslmode=disable"
```

### 启动

```bash
./gh-proxy-auth -config config.yaml
```

首次访问 Web 页面会自动跳转到注册页面，第一个注册的用户将成为管理员。

## 使用方式

### Web 页面

访问服务地址，在首页输入 GitHub 链接，系统会自动生成包含 Token 的代理下载命令。

### 代理下载

#### 通过 curl 下载

```bash
curl -H "X-XN-Token: YOUR_TOKEN" -L -O https://proxy.example.com/https://github.com/user/repo/releases/download/v1.0/file.zip
```

#### 通过 wget 下载

```bash
wget --header="X-XN-Token: YOUR_TOKEN" https://proxy.example.com/https://github.com/user/repo/releases/download/v1.0/file.zip
```

#### 通过 git clone

```bash
git -c http.extraHeader="X-XN-Token: YOUR_TOKEN" clone https://proxy.example.com/https://github.com/user/repo.git
```

#### 通过 URL 参数传递 Token

```
https://proxy.example.com/https://github.com/user/repo/releases/download/v1.0/file.zip?token=YOUR_TOKEN
```

### 支持的 GitHub URL 格式

| 类型 | 示例 |
|------|------|
| Releases | `github.com/user/repo/releases/download/v1.0/file.zip` |
| Archive | `github.com/user/repo/archive/main.zip` |
| Blob/Raw | `github.com/user/repo/blob/main/README.md` |
| Raw Content | `raw.githubusercontent.com/user/repo/main/file.txt` |
| Gist | `gist.githubusercontent.com/user/hash/raw/file.txt` |
| Tags | `github.com/user/repo/tags` |
| Git Clone | `github.com/user/repo/info/refs?service=git-upload-pack` |

## 反向代理配置

建议使用反向代理（Nginx 或 Caddy）来提供 HTTPS 支持。

### Nginx 配置

```nginx
server {
    listen 443 ssl http2;
    server_name proxy.example.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    client_max_body_size 0;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 支持大文件下载和流式传输
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;

        # WebSocket 支持（如需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name proxy.example.com;
    return 301 https://$host$request_uri;
}
```

### Caddy 配置

```
proxy.example.com {
    reverse_proxy 127.0.0.1:8080 {
        flush_interval -1
        transport http {
            read_timeout 600s
            write_timeout 600s
        }
    }
}
```

> **重要提示：**
> - 配置文件中的 `domain` 字段必须设置为实际访问的域名（包含协议），例如 `https://proxy.example.com`，这关系到 Passkey (WebAuthn) 功能是否能正常使用
> - Passkey 功能要求使用 HTTPS，请确保配置了 SSL 证书
> - 如果使用反向代理，建议将服务监听在 `127.0.0.1` 上以增加安全性

## Systemd 服务配置

创建 systemd 服务文件 `/etc/systemd/system/gh-proxy-auth.service`：

```ini
[Unit]
Description=GH Proxy Auth
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/gh-proxy-auth
ExecStart=/opt/gh-proxy-auth/gh-proxy-auth -config /opt/gh-proxy-auth/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable gh-proxy-auth
sudo systemctl start gh-proxy-auth
```

## API 说明

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/system/init-status` | 获取系统初始化状态 |
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| POST | `/api/auth/totp/verify` | TOTP 验证 |
| POST | `/api/auth/passkey/begin-login` | 开始 Passkey 认证 |
| POST | `/api/auth/passkey/finish-login` | 完成 Passkey 认证 |

### 需要登录（JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/profile` | 获取用户信息 |
| PUT | `/api/user/password` | 修改密码 |
| POST | `/api/user/totp/setup` | 设置 TOTP |
| POST | `/api/user/totp/enable` | 启用 TOTP |
| DELETE | `/api/user/totp` | 关闭 TOTP |
| PUT | `/api/user/mfa-priority` | 设置 MFA 优先级 |
| GET | `/api/user/passkeys` | 获取 Passkey 列表 |
| POST | `/api/user/passkey/begin-register` | 开始注册 Passkey |
| POST | `/api/user/passkey/finish-register` | 完成注册 Passkey |
| DELETE | `/api/user/passkey/:id` | 删除 Passkey |
| GET | `/api/tokens` | 获取 Token 列表 |
| POST | `/api/tokens` | 创建 Token |
| PUT | `/api/tokens/:id` | 更新 Token |
| DELETE | `/api/tokens/:id` | 删除 Token |
| GET | `/api/tokens/:id/logs` | 获取 Token 下载记录 |

### 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/settings` | 获取系统设置 |
| PUT | `/api/admin/settings` | 更新系统设置 |
| GET | `/api/admin/users` | 获取用户列表 |
| GET | `/api/admin/logs` | 获取所有下载记录 |

## 技术栈

- **后端**：Go + Gin + GORM
- **前端**：Vue 3 + TypeScript + Tailwind CSS + Lucide Icons
- **认证**：JWT + TOTP + WebAuthn/Passkey
- **数据库**：SQLite / MySQL / PostgreSQL

## 许可证

MIT License
