defmodule TodoTaskManager.Repo do
  use Ecto.Repo,
    otp_app: :todo_task_manager,
    adapter: Ecto.Adapters.Postgres
end
