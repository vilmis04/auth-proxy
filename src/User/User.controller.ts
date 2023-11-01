import { Request, Response } from "express";
// import { UserService } from "./User.service";
// import { UserRepo } from "./User.repo";
// import { JwtUtils } from "../utilities/JwtUtils/JwtUtils";

export class UserController {
  constructor(
    private readonly request: Request,
    private readonly response: Response
  ) {}

  // private readonly userService = new UserService(
  //   new UserRepo(),
  //   new JwtUtils()
  // );

  async signup() {
    const body = this.request.body;
    // const { userId, password } = this.request.body;
    console.log(this.request);
    this.response.send(JSON.stringify(body));
    // await this.userService.signUp(userId, password);
  }
}
