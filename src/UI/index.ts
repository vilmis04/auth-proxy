import dotenv from "dotenv";

dotenv.config();

const onLoad = () => {
  // TODO: 2023-11-19 add listeners where needed
  document.querySelector("button")?.addEventListener("click", () => {});
};

const header = process.env.HEADER || "Welcome!";

export const renderAuthScreen = () => `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Auth</title>
    <style></style>
  </head>
  <body>
    <header>${header}</header>
    <main>Main</main>
    <footer>Footer</footer>
    <script>
      document.addEventListener('DOMContentLoaded', ${onLoad})
    </script>
  </body>
</html>
`;
