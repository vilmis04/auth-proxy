import { Response } from "express";
import { JwtUtils } from "../utilities/JwtUtils/JwtUtils";
import { UserEntity } from "./User.entity";
import { UserRepo } from "./User.repo";
import * as bcrypt from "bcryptjs";

const JWT_MAX_AGE = Number(process.env.JWT_MAX_AGE) || undefined;
const JWT = "jwt";

export class UserService extends UserEntity {
  constructor(
    private readonly userRepo: UserRepo,
    private readonly jwtUtils: JwtUtils
  ) {
    super();
  }

  async login(
    username: string,
    enteredPassword: string,
    response: Response
  ): Promise<void> {
    const user = await this.userRepo
      .getOne({ username })
      .catch((err: Error) => {
        throw new Error(err.message);
      });

    if (!user) {
      response.status(400).send();
      throw new Error("Incorrect username");
    }
    if (!this.validateUser(enteredPassword, user.password)) {
      response.status(400).send();
      throw new Error("Incorrect password");
    }

    const { password, ...userData } = user;
    const access_token = await this.jwtUtils.generateToken(userData);

    response.cookie(JWT, access_token, {
      maxAge: JWT_MAX_AGE,
      httpOnly: true,
    });
  }

  async validateUser(enteredPassword: string, userPassword: string) {
    return await bcrypt.compare(enteredPassword, userPassword);
  }

  async logout(response: Response) {
    response.cookie(JWT, "", {
      maxAge: 1,
      httpOnly: true,
    });
  }

  async signUp(
    username: string,
    enteredPassword: string,
    response: Response
  ): Promise<void> {
    const isUsernameTaken = Boolean(await this.userRepo.getOne({ username }));

    if (isUsernameTaken) {
      response.status(400).send();
      throw new Error("Username already exists");
    }

    const salt = bcrypt.genSaltSync(10);
    const hash = bcrypt.hashSync(enteredPassword, salt);

    const { acknowledged: userAcknowledged } = await this.userRepo.create({
      username,
      password: hash,
    });

    if (!userAcknowledged) {
      response.status(500).send();
      throw new Error("Internal server error");
    }

    await this.login(username, enteredPassword, response);
  }
}
