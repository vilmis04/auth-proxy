import { Request, Response } from "express";

export class UserController {
  constructor(
    private readonly request: Request,
    private readonly response: Response
  ) {}
}
