## Environment config

config file for env vars:
ALLOWED_ORIGINS=<allowed origins, i.e., http://localhost:3000,http://localhost:3300>

### Database

```
PORT=<server port>
POSTGRES_PASSWORD=<password>
POSTGRES_PORT=<database port>
POSTGRES_USER=<database user>
POSTGRES_DB=<database name>
POSTGRES_HOST=<host name, i.e., "db">
SERVICE_URL=<service url, i.e., "http://service:8080">
```
## Deployment

- run command to deploy to image to dockerhub (replace x.x.x with version): `docker build -t vsud/utils:auth-proxy-x.x.x . && docker push vsud/utils:auth-proxy-x.x.x`
- ssh into the server, update docker compose with the new image version tag: `vi ~/PROJECTS/docker-compose.yaml`
- run `docker compose up -d` to start the service with the changes