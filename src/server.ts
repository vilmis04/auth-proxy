import express, { Request, Response, Application } from "express";
import dotenv from "dotenv";

dotenv.config();
enum Prefixes {
  GLOBAL_PREFIX = "/api",
  AUTH = "/auth",
  PROXY = "/proxy",
}

enum URIs {
  HEALTH = `${Prefixes.GLOBAL_PREFIX}/health`,
  AUTH = `${Prefixes.GLOBAL_PREFIX}${Prefixes.AUTH}`,
}

const app: Application = express();
const PORT = process.env.PORT || 8000;

app.get(URIs.HEALTH, (_req: Request, res: Response) => {
  res.status(200).send("OK");
});

app.listen(PORT, () => {
  console.log(`Server is live at http://localhost:${PORT}`);
});
