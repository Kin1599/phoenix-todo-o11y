# This file is responsible for configuring your application
# and its dependencies with the aid of the Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

# General application configuration
import Config

config :todo_task_manager,
  ecto_repos: [TodoTaskManager.Repo],
  generators: [timestamp_type: :utc_datetime],
  migration_primary_key: [type: :binary_id],
  migration_foreign_key: [type: :binary_id]

config :todo_task_manager, TodoTaskManager.Repo,
  migration_primary_key: [type: :binary_id],
  migration_foreign_key: [type: :binary_id]

config :ecto,
  primary_key_type: :binary_id,
  foreign_key_type: :binary_id

config :todo_task_manager, :phoenix_swagger,
  swagger_files: %{
    "priv/static/swagger.json" => [
      router: TodoTaskManagerWeb.Router,
      endpoint: TodoTaskManagerWeb.Endpoint
    ]
  }

config :todo_task_manager, TodoTaskManager.Guardian,
  issuer: "todo_task_manager",
  secret_key: "kq3dA2jVtIlQ0A3eLoZp9RyspnScyUz2rC1yq1VgrUR0EzpZ6Nn62w=="

# Configures the endpoint
config :todo_task_manager, TodoTaskManagerWeb.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  render_errors: [
    formats: [html: TodoTaskManagerWeb.ErrorHTML, json: TodoTaskManagerWeb.ErrorJSON],
    layout: false
  ],
  pubsub_server: TodoTaskManager.PubSub,
  live_view: [signing_salt: "O3JBL/Bf"]

# Configure esbuild (the version is required)
config :esbuild,
  version: "0.17.11",
  todo_task_manager: [
    args:
      ~w(js/app.js --bundle --target=es2017 --outdir=../priv/static/assets --external:/fonts/* --external:/images/*),
    cd: Path.expand("../assets", __DIR__),
    env: %{"NODE_PATH" => Path.expand("../deps", __DIR__)}
  ]

# Configure tailwind (the version is required)
config :tailwind,
  version: "3.4.3",
  todo_task_manager: [
    args: ~w(
      --config=tailwind.config.js
      --input=css/app.css
      --output=../priv/static/assets/app.css
    ),
    cd: Path.expand("../assets", __DIR__)
  ]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{config_env()}.exs"
