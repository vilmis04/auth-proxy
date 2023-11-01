export enum Constants {
  JWT = "jwt",
}

export enum Prefixes {
  GLOBAL_PREFIX = "/api",
  AUTH = "/auth",
  PROXY = "/proxy",
}

export enum Resources {
  HEALTH = "health",
}

export enum URIs {
  HEALTH = `${Prefixes.GLOBAL_PREFIX}/${Resources.HEALTH}`,
  AUTH = `${Prefixes.GLOBAL_PREFIX}${Prefixes.AUTH}`,
  PROXY = `${Prefixes.GLOBAL_PREFIX}${Prefixes.PROXY}`,
}
