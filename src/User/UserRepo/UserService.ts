// const LOGIN_METHOD = process.env.LOGIN_METHOD || "email";
// const PASSWORDLESS = process.env.PASSWORDLESS === "true" ? true : false;

import { UserRepo } from "./UserRepo";

export class UserService {
  constructor(private readonly userRepo: UserRepo) {}
}
