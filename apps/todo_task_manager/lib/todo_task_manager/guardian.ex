defmodule TodoTaskManager.Guardian do
  use Guardian, otp_app: :todo_task_manager

  def subject_for_token(user, _claims), do: {:ok, to_string(user.id)}

  def resource_from_claims(%{"sub" => id}) do
    case TodoTaskManager.Accounts.get_user!(id) do
      nil -> {:error, :not_found}
      user -> {:ok, user}
    end
  end
end
