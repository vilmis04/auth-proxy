import { BadRequestException, Inject, Injectable } from '@nestjs/common';
import { LoginDto } from './dto/login.dto';
import { CreateAuthDto } from './dto/create-auth.dto';
import { Repo } from '../Repo/Repo';

@Injectable()
export class AuthService {
  constructor(@Inject() private readonly repo: Repo) {}
  signUp({ email, password, repeatPassword }: CreateAuthDto) {
    if (password !== repeatPassword)
      throw new BadRequestException('passwords do not match');

    return this.repo.createUser(email, password);
  }

  login({ email, password }: LoginDto) {
    return this.repo.loginUser(email, password);
  }

  refresh(id: number) {
    return `This action returns a #${id} auth`;
  }

  validate(id: string) {
    return `This action updates a #${id} auth`;
  }

  remove(id: number) {
    return `This action removes a #${id} auth`;
  }
}
