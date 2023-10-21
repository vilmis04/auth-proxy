export abstract class Repo {
  constructor() {}
  // move this to class for specific collection
  createUser(email: string, password: string) {
    console.log(email, password);
  }
  loginUser(email: string, password: string) {
    console.log(email, password);
  }
}
