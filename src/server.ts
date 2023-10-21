import express, { Request, Response, Application } from "express";
import dotenv from "dotenv";

dotenv.config();

const app: Application = express();
const PORT = process.env.PORT || 8000;

app.get("/", (_req: Request, res: Response) => {
  res.send("Welcome to Express & TypeScript Server");
});

app.listen(PORT, () => {
  console.log(`Server is live at http://localhost:${PORT}`);
});
