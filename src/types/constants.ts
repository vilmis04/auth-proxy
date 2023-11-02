export enum Constants {
  JWT = "jwt",
}

export enum Prefixes {
  GLOBAL_PREFIX = "/api",
  AUTH = "/auth",
  PROXY = "/proxy",
}

export enum URIs {
  GLOBAL = `${Prefixes.GLOBAL_PREFIX}`,
  AUTH = `${Prefixes.GLOBAL_PREFIX}${Prefixes.AUTH}`,
  PROXY = `${Prefixes.GLOBAL_PREFIX}${Prefixes.PROXY}`,
}
