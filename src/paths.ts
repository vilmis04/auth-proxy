import { URIs } from "./types/constants";

export enum Resources {
  HEALTH = "health",
  SIGNUP = "signup",
  LOGIN = "login",
  LOGOUT = "logout",
}

const buildPath = (prefix: URIs, resource?: string) => `${prefix}/${resource}`;

export const paths = {
  health: buildPath(URIs.GLOBAL, Resources.HEALTH),
  signup: buildPath(URIs.AUTH, Resources.SIGNUP),
  login: buildPath(URIs.AUTH, Resources.LOGIN),
  logout: buildPath(URIs.AUTH, Resources.LOGOUT),
};
