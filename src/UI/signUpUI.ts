import dotenv from "dotenv";

dotenv.config();

const baseUrl = process.env.BASE_URL || "http://localhost:8000/api";

const onSignUpLoad = () => {
  // TODO: 2023-11-19 add listeners where needed
  document.querySelector(".login-button")?.addEventListener("click", () => {
    window.open(`${baseUrl}/login-ui`, "_self");
  });
};

const signUpHeader = process.env.SIGN_UP_HEADER || "Welcome!";

export const renderSignUpUI = () => `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Auth</title>
    <style></style>
  </head>
  <body>
  <header>${signUpHeader}</header>
  <main>
    <form>
      <div>
        <label>
          User ID:
          <input type="text" name="userId" />
        </label>
      </div>
      <div>
        <label>
          Password:
          <input type="password" name="password" />
        </label>
        <label>
          Repeat password:
          <input type="password" name="repeatPassword" />
        </label>
      </div>
      <button type="submit">
        Sign up!
      </button>
    </form>
    <div>Passkeys!!</div>
    <div> 
      <p>
        Have an account?
      </p>
      <button class="login-button">
        Login!
      </button>
    </div>
  </main>
  <script>
    document.addEventListener('DOMContentLoaded', ${onSignUpLoad})
  </script>
  </body>
</html>
`;
