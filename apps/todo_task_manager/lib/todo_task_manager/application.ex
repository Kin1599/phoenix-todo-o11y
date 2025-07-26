defmodule TodoTaskManager.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    OpentelemetryPhoenix.setup()
    OpentelemetryEcto.setup([:todo_task_manager, TodoTaskManager.Repo])

    children = [
      TodoTaskManagerWeb.Telemetry,
      TodoTaskManager.Repo,
      {DNSCluster, query: Application.get_env(:todo_task_manager, :dns_cluster_query) || :ignore},
      {Phoenix.PubSub, name: TodoTaskManager.PubSub},
      # Start a worker by calling: TodoTaskManager.Worker.start_link(arg)
      # {TodoTaskManager.Worker, arg},
      # Start to serve requests, typically the last entry
      TodoTaskManagerWeb.Endpoint,
      TodoTaskManager.PromEx
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: TodoTaskManager.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  @impl true
  def config_change(changed, _new, removed) do
    TodoTaskManagerWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
