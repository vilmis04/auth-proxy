import { Request } from "express";
import { verify, JsonWebTokenError, sign } from "jsonwebtoken";

const SECRET_KEY = process.env.SECRET_KEY || "";

export class JwtUtils {
  public getUser(req: Request) {
    const token = req.cookies.jwt;
    const user = verify(token, SECRET_KEY);

    if (user == null) {
      throw new JsonWebTokenError("Unauthorized");
    }

    return user;
  }

  public verifyToken(token: string) {
    const user = verify(token, SECRET_KEY);

    if (user == null) {
      throw new JsonWebTokenError("Unauthorized");
    }

    return user;
  }

  async generateToken(data: string | object, maxAge?: number): Promise<string> {
    const token = sign(data, SECRET_KEY, {
      expiresIn: maxAge,
    });

    return token;
  }

  public getAuthStatus(req: Request): boolean {
    const token = req.cookies?.jwt;
    if (!token) return false;
    const user = verify(token, SECRET_KEY);
    if (!user) return false;

    return true;
  }
}
