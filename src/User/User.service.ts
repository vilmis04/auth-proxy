import { JwtUtils } from "../utilities/JwtUtils/JwtUtils";
import { UserEntity } from "./User.entity";
import { UserRepo } from "./User.repo";
import * as bcrypt from "bcryptjs";

export class UserService extends UserEntity {
  constructor(private readonly userRepo: UserRepo) {
    super();
  }

  async login(userId: string, enteredPassword: string): Promise<string> {
    const user = await this.userRepo.getOne({ userId }).catch((err: Error) => {
      throw new Error(err.message);
    });

    if (!user) {
      // TODO: add error code to error (?)
      throw Error("Incorrect userId");
    }
    if (!this.validateUser(enteredPassword, user.password)) {
      // TODO: add error code to error (?)
      throw new Error("Incorrect password");
    }

    const { password, ...userData } = user;
    const accessToken = JwtUtils.generateToken(userData);

    return accessToken;
  }

  async validateUser(enteredPassword: string, userPassword: string) {
    return await bcrypt.compare(enteredPassword, userPassword);
  }

  async signup(userId: string, enteredPassword: string): Promise<string> {
    const isUserIdTaken = Boolean(await this.userRepo.getOne({ userId }));

    if (isUserIdTaken) {
      // TODO: add error code to error (?)
      throw new Error("userId already exists");
    }

    const salt = bcrypt.genSaltSync(10);
    const hash = bcrypt.hashSync(enteredPassword, salt);

    const { acknowledged: userAcknowledged } = await this.userRepo.create({
      userId: userId,
      password: hash,
    });

    if (!userAcknowledged) {
      // TODO: add error code to error (?)
      throw new Error("Internal server error");
    }

    return await this.login(userId, enteredPassword);
  }
}
