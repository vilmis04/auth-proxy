import express, { Request, Response, Application } from "express";
import dotenv from "dotenv";
import morgan from "morgan";
import cors from "cors";
import { UserController } from "./User/User.controller";
import { Proxy } from "./Proxy/Proxy";
import { URIs } from "./types/constants";
import { paths } from "./paths";
import { renderAuthScreen } from "./UI";

dotenv.config();

const app: Application = express();
const PORT = process.env.PORT || 8000;

app.use(morgan("tiny"));
app.use(cors());
app.use(express.json());

app.get(paths.authUI, (_req: Request, res: Response) => {
  res.send(renderAuthScreen());
});

app.get(paths.health, (_req: Request, res: Response) => {
  res.status(200).send("OK");
});

// TODO:
app.post(paths.signup, (req: Request, res: Response) => {
  new UserController(req, res).signup();
});

// TODO:
app.post(paths.login, (req: Request, res: Response) => {
  new UserController(req, res).login();
});

// TODO:
app.get(paths.logout, (req: Request, res: Response) => {
  new UserController(req, res).logout();
});

// Keep these in the bottom of the file so auth paths matched first
app.use(URIs.GLOBAL, Proxy.authMiddleware);
app.use(URIs.GLOBAL, Proxy.proxyMiddleware());

app.listen(PORT, () => {
  console.log(`Server is live at http://localhost:${PORT}`);
});
