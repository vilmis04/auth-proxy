interface IUserConstructorParams {
  email?: string;
  username?: string;
  password?: string;
}

export class User {
  // @ts-expect-error
  private email: string;
  // @ts-expect-error
  private username: string;
  // @ts-expect-error
  private password: string;

  constructor(userInputs: IUserConstructorParams) {
    this.email = userInputs.email || "";
    this.password = userInputs.password || "";
    this.username = userInputs.username || "";
  }
}
