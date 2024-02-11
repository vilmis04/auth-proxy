# Environment config

config file for env vars:

```
MONGODB_URI={connection uri to mongo db}
SECRET_KEY={encoding key}
JWT_MAX_AGE={defaults to unlimited}
PORT={defaults to 8080}
PROXY_TARGET={where to forward the validated request}
AUTH_TOKEN={<clientId>:<clientKey>}
```

Add these to configure UI:

```
LOGIN_HEADER={valid html}
SIGN_UP_HEADER={valid html}
```

# Roadmap

> Soon

- Standard UI

> Later

- Configurable standard UI
