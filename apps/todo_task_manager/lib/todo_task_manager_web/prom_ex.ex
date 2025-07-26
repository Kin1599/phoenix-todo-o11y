defmodule TodoTaskManager.PromEx do
  use PromEx, otp_app: :todo_task_manager

  @impl true
  def plugins do
    [
      {PromEx.Plugins.Phoenix,
       endpoint: TodoTaskManagerWeb.Endpoint,
       router: TodoTaskManagerWeb.Router},
      {PromEx.Plugins.Ecto, repos: [TodoTaskManager.Repo]}
    ]
  end

  @impl true
  def dashboards, do: []

  @impl true
  def metrics_exporter, do: :plug
end
