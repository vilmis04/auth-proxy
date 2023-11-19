import { URIs } from "./types/constants";

export enum Resources {
  AUTH_UI = "auth-ui",
  HEALTH = "health",
  SIGNUP = "signup",
  LOGIN = "login",
  LOGOUT = "logout",
}

const buildPath = (prefix: URIs, resource?: string) => `${prefix}/${resource}`;

export const paths = {
  authUI: buildPath(URIs.GLOBAL, Resources.AUTH_UI),
  health: buildPath(URIs.GLOBAL, Resources.HEALTH),
  signup: buildPath(URIs.AUTH, Resources.SIGNUP),
  login: buildPath(URIs.AUTH, Resources.LOGIN),
  logout: buildPath(URIs.AUTH, Resources.LOGOUT),
};
