# Sub2API 部署

本仓库由 [mizaawa/sub2api](https://github.com/mizaawa/sub2api) 维护。部署镜像、安装脚本以及后台的一键更新和版本回退均使用此仓库的 Release。

## Docker Compose 部署

前置条件：Docker Engine 20.10+ 与 Docker Compose v2+。

### 一键准备

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/mizaawa/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

脚本会生成 `.env`，自动生成 `POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY`，并使用本地的 `data`、`postgres_data`、`redis_data` 目录，便于备份和整体迁移。

容器健康后访问 `http://服务器IP:8080`。如果 `.env` 未设置 `ADMIN_PASSWORD`，请从应用日志中查看首次生成的管理员密码。

### 手动部署

```bash
git clone https://github.com/mizaawa/sub2api.git
cd sub2api/deploy
cp .env.example .env
chmod 600 .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
```

生产环境启动前，请在 `.env` 中设置强随机值 `POSTGRES_PASSWORD`、`JWT_SECRET` 和 `TOTP_ENCRYPTION_KEY`；可按需设置 `ADMIN_EMAIL`、`ADMIN_PASSWORD`、`SERVER_PORT`。

### 更新与回退

Docker 部署更新到最新已发布镜像：

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

若要固定或回退 Docker 镜像版本，在 `.env` 中设置 `SUB2API_IMAGE=ghcr.io/mizaawa/sub2api:<版本号>` 后运行相同命令。管理后台的一键更新、回退均从 `mizaawa/sub2api` 的 GitHub Releases 获取版本，需先在本仓库发布带二进制资源的 Release。

### 常用命令

```bash
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs -f sub2api
docker compose -f docker-compose.local.yml restart
docker compose -f docker-compose.local.yml down
```
