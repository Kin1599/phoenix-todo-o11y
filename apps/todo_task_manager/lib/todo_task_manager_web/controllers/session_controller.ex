defmodule TodoTaskManagerWeb.SessionController do
  use TodoTaskManagerWeb, :controller
  use PhoenixSwagger
  alias TodoTaskManager.Accounts

  swagger_path :create do
    post "/api/login"
    summary "User login"
    description "Authenticate user and return JWT token"
    parameters do
      body :body, Schema.ref(:LoginRequest), "User credentials", required: true
    end
    response 200, "Token", Schema.ref(:TokenResponse)
    response 401, "Invalid credentials"
  end

  def create(conn, %{"email" => email, "password" => password}) do
    case Accounts.authenticate_user(email, password) do
      {:ok, user} ->
        {:ok, token, _claims} = TodoTaskManager.Guardian.encode_and_sign(user)
        json(conn, %{token: token})
      _ ->
        conn
        |> put_status(:unauthorized)
        |> json(%{error: "invalid credentials"})
    end
  end


  def swagger_definitions do
    %{
      LoginRequest: swagger_schema do
        title "LoginRequest"
        properties do
          email :string, "Email", required: true
          password :string, "Password", required: true
        end
        example %{
          email: "user@example.com",
          password: "verysecret"
        }
      end,
      TokenResponse: swagger_schema do
        title "TokenResponse"
        properties do
          token :string, "JWT token", required: true
        end
        example %{
          token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        }
      end
    }
  end
end
