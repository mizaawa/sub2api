# Sub2API Deployment

This repository is maintained at [mizaawa/sub2api](https://github.com/mizaawa/sub2api). Deployment images, install scripts, and the in-app version update and rollback feature all use this repository's releases.

## Docker Compose

Prerequisites: Docker Engine 20.10+ and Docker Compose v2+.

### One-command preparation

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/mizaawa/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

The preparation script creates `.env`, generates `POSTGRES_PASSWORD`, `JWT_SECRET`, and `TOTP_ENCRYPTION_KEY`, and uses local `data`, `postgres_data`, and `redis_data` directories so the deployment can be backed up or moved as one directory.

Open `http://SERVER_IP:8080` after the containers become healthy. If no `ADMIN_PASSWORD` is set in `.env`, inspect the application log for the generated password.

### Manual deployment

```bash
git clone https://github.com/mizaawa/sub2api.git
cd sub2api/deploy
cp .env.example .env
chmod 600 .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
```

Before starting a production deployment, set strong values for `POSTGRES_PASSWORD`, `JWT_SECRET`, and `TOTP_ENCRYPTION_KEY` in `.env`. Optionally set `ADMIN_EMAIL`, `ADMIN_PASSWORD`, and `SERVER_PORT`.

### Update and rollback

Update Docker containers from the published image:

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

To pin a specific published image version, set `SUB2API_IMAGE=ghcr.io/mizaawa/sub2api:<version>` in `.env`, then run the same two commands. The administrator console checks releases from `mizaawa/sub2api`; its one-click update and rollback options require release assets published by this repository.

### Common commands

```bash
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs -f sub2api
docker compose -f docker-compose.local.yml restart
docker compose -f docker-compose.local.yml down
```
