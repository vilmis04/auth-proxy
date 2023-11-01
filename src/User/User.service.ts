import { Response } from "express";
import { JwtUtils } from "../utilities/JwtUtils/JwtUtils";
import { UserEntity } from "./User.entity";
import { UserRepo } from "./User.repo";
import * as bcrypt from "bcryptjs";
import { Constants } from "../types/constants";

const JWT_MAX_AGE = Number(process.env.JWT_MAX_AGE) || undefined;

export class UserService extends UserEntity {
  constructor(private readonly userRepo: UserRepo) {
    super();
  }

  async login(
    userId: string,
    enteredPassword: string,
    response: Response
  ): Promise<void> {
    const user = await this.userRepo.getOne({ userId }).catch((err: Error) => {
      throw new Error(err.message);
    });

    if (!user) {
      response.status(400).send();
      throw new Error("Incorrect userId");
    }
    if (!this.validateUser(enteredPassword, user.password)) {
      response.status(400).send();
      throw new Error("Incorrect password");
    }

    const { password, ...userData } = user;
    const access_token = JwtUtils.generateToken(userData);

    response.cookie(Constants.JWT, access_token, {
      maxAge: JWT_MAX_AGE,
      httpOnly: true,
    });
  }

  async validateUser(enteredPassword: string, userPassword: string) {
    return await bcrypt.compare(enteredPassword, userPassword);
  }

  async logout(response: Response) {
    response.cookie(Constants.JWT, "", {
      maxAge: 1,
      httpOnly: true,
    });
  }

  async signUp(
    userId: string,
    enteredPassword: string,
    response: Response
  ): Promise<void> {
    const isuserIdTaken = Boolean(await this.userRepo.getOne({ userId }));

    if (isuserIdTaken) {
      response.status(400).send();
      throw new Error("userId already exists");
    }

    const salt = bcrypt.genSaltSync(10);
    const hash = bcrypt.hashSync(enteredPassword, salt);

    const { acknowledged: userAcknowledged } = await this.userRepo.create({
      userId: userId,
      password: hash,
    });

    if (!userAcknowledged) {
      response.status(500).send();
      throw new Error("Internal server error");
    }

    await this.login(userId, enteredPassword, response);
  }
}
