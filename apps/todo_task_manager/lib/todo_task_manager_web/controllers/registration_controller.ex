defmodule TodoTaskManagerWeb.RegistrationController do
  use TodoTaskManagerWeb, :controller
  use PhoenixSwagger
  alias TodoTaskManager.Accounts

  swagger_path :create do
    post "/api/register"
    summary "User registration"
    description "Creates a new user with email and password"
    parameters do
      body :body, Schema.ref(:UserRegister), "User registration params", required: true
    end
    response 200, "User created", Schema.ref(:UserResponse)
    response 400, "Validation error"
  end

  def create(conn, %{"email" => email, "password" => password}) do
    case Accounts.create_user(%{"email" => email, "password" => password}) do
      {:ok, user} -> json(conn, %{id: user.id, email: user.email})
      {:error, changeset} -> conn |> put_status(:bad_request) |> json(%{errors: translate_errors(changeset)})
    end
  end

  defp translate_errors(changeset) do
    Ecto.Changeset.traverse_errors(changeset, fn {msg, _opts} -> msg end)
  end

  def swagger_definitions do
    %{
      UserRegister: swagger_schema do
        title "UserRegister"
        description "Params for user registration"
        properties do
          email :string, "Email", required: true
          password :string, "Password", required: true
        end
        example %{
          email: "user@example.com",
          password: "verysecret"
        }
      end,
      UserResponse: swagger_schema do
        title "UserResponse"
        description "User creation result"
        properties do
          id :string, "User ID"
          email :string, "Email"
        end
        example %{
          id: "6f1b3a66-1234-4567-89ab-0cffe07d0001",
          email: "user@example.com"
        }
      end
    }
  end
end
