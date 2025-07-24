defmodule TodoTaskManager.Accounts do
  alias TodoTaskManager.Accounts.User
  alias TodoTaskManager.Repo

  def get_user_by_email(email), do: Repo.get_by(User, email: email)

  def get_user!(id), do: Repo.get!(User, id)

  def create_user(attrs) do
    %User{} |> User.registration_changeset(attrs) |> Repo.insert()
  end

  def authenticate_user(email, password) do
    case get_user_by_email(email) do
      nil -> {:error, :not_found}
      user ->
        if Bcrypt.verify_pass(password, user.password_hash) do
          {:ok, user}
        else
          {:error, :invalid_password}
        end
    end
  end
end
