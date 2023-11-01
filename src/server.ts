import express, { Request, Response, Application } from "express";
import dotenv from "dotenv";
import morgan from "morgan";
import cors from "cors";
import { UserController } from "./User/User.controller";
import { Proxy } from "./Proxy/Proxy";
import { URIs } from "./types/constants";

dotenv.config();

const app: Application = express();
const PORT = process.env.PORT || 8000;

app.use(morgan("tiny"));
app.use(cors());
app.use(URIs.PROXY, Proxy.authMiddleware);
app.use(URIs.PROXY, Proxy.proxyMiddleware());

app.get(URIs.HEALTH, (_req: Request, res: Response) => {
  res.status(200).send("OK");
});

// TODO:
app.post(`${URIs.AUTH}/signup`, (req: Request, res: Response) => {
  new UserController(req, res).signup();
});

// TODO:
app.post(`${URIs.AUTH}/login`, (_req: Request, res: Response) => {
  res.status(200).send("OK");
});

// TODO:
app.get(`${URIs.AUTH}/logout`, (_req: Request, res: Response) => {
  res.status(200).send("OK");
});

app.listen(PORT, () => {
  console.log(`Server is live at http://localhost:${PORT}`);
});
