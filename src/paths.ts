import { URIs } from "./types/constants";

export enum Resources {
  LOGIN_UI = "login-ui",
  SIGN_UP_UI = "sign-up-ui",
  HEALTH = "health",
  SIGNUP = "signup",
  LOGIN = "login",
  LOGOUT = "logout",
}

const buildPath = (prefix: URIs, resource?: string) => `${prefix}/${resource}`;

export const paths = {
  loginUI: buildPath(URIs.GLOBAL, Resources.LOGIN_UI),
  signUpUI: buildPath(URIs.GLOBAL, Resources.SIGN_UP_UI),
  health: buildPath(URIs.GLOBAL, Resources.HEALTH),
  signup: buildPath(URIs.AUTH, Resources.SIGNUP),
  login: buildPath(URIs.AUTH, Resources.LOGIN),
  logout: buildPath(URIs.AUTH, Resources.LOGOUT),
};
