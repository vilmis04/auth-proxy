import { NextFunction, Request, Response } from "express";
import { Constants } from "../types/constants";
import { JwtUtils } from "../utilities/JwtUtils/JwtUtils";
import { createProxyMiddleware } from "http-proxy-middleware";
import { Prefixes } from "../types/constants";

export class Proxy {
  public static authMiddleware = (
    request: Request,
    response: Response,
    next: NextFunction
  ) => {
    const token = request.cookies[Constants.JWT];
    try {
      const userId = JwtUtils.verifyToken(token);

      request.headers.authorization = `Basic ${btoa(
        process.env.TARGET_AUTH || ""
      )}`;
      request.headers.userId = `${userId}`;
    } catch (error) {
      response.status(401).send("Unauthorized");
    }

    next();
  };

  public static proxyMiddleware() {
    return createProxyMiddleware({
      target: process.env.PROXY_TARGET,
      changeOrigin: true,
      pathRewrite: (path) => path.replace(Prefixes.PROXY, ""),
    });
  }
}
