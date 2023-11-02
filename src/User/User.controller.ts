import { Request, Response } from "express";
import { UserService } from "./User.service";
import { UserRepo } from "./User.repo";
import { Constants } from "../types/constants";

const JWT_MAX_AGE = Number(process.env.JWT_MAX_AGE) || undefined;

export class UserController {
  private readonly userService = new UserService(new UserRepo());

  constructor(
    private readonly request: Request,
    private readonly response: Response
  ) {}

  async signup() {
    const { userId, password } = this.request.body;
    const accessToken = await this.userService
      .signup(userId, password)
      .catch((err) => {
        console.log(err);
        this.response.send();
      });

    this.response.status(200).cookie(Constants.JWT, accessToken, {
      maxAge: JWT_MAX_AGE,
      httpOnly: true,
    });
  }

  // TODO:
  public async login() {}

  public async logout() {
    this.response.cookie(Constants.JWT, "", {
      maxAge: 1,
      httpOnly: true,
    });
  }
}
