import { Request } from "express";
import { verify, sign } from "jsonwebtoken";

const SECRET_KEY = process.env.SECRET_KEY || "";

export class JwtUtils {
  public static verifyToken(token: string) {
    return verify(token, SECRET_KEY);
  }

  public static generateToken(data: string | object, maxAge?: number): string {
    const token = sign(data, SECRET_KEY, {
      expiresIn: maxAge,
    });

    return token;
  }

  public static getAuthStatus(req: Request): boolean {
    const token = req.cookies?.jwt;
    if (!token) return false;
    const user = verify(token, SECRET_KEY);
    if (!user) return false;

    return true;
  }
}
