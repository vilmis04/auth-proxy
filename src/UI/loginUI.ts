import dotenv from "dotenv";

dotenv.config();

const baseUrl = process.env.BASE_URL || "http://localhost:8000/api";

const onLoginLoad = () => {
  // TODO: 2023-11-19 add listeners where needed
  document.querySelector(".sign-up-button")?.addEventListener("click", () => {
    window.open(`${baseUrl}/sign-up-ui`, "_self");
  });
};

const loginHeader = process.env.LOGIN_HEADER || "Welcome!";

export const renderLoginUI = () => `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Auth</title>
    <style></style>
  </head>
  <body>
    <header>${loginHeader}</header>
    <main>
      <form>
        <div>
          <label>
            User ID:
            <input type="text" />
          </label>
        </div>
        <div>
          <label>
            Password:
            <input type="password" />
          </label>
        </div>
        <button type="submit">
          Login
        </button>
      </form>
      <div>Passkeys!!</div>
      <div>
        <button class="sign-up-button">
          Sign up!
        </button>
      </div>
    </main>
    <script>
      document.addEventListener('DOMContentLoaded', ${onLoginLoad})
    </script>
  </body>
</html>
`;
