defmodule TodoTaskManagerWeb.Router do
  use TodoTaskManagerWeb, :router

  # --- браузерные маршруты ---
  pipeline :browser do
    plug :accepts, ["html"]
    plug :fetch_session
    plug :fetch_live_flash
    plug :put_root_layout, html: {TodoTaskManagerWeb.Layouts, :root}
    plug :protect_from_forgery
    plug :put_secure_browser_headers
  end

  # --- публичный API ---
  pipeline :api do
    plug :accepts, ["json"]
  end

  # --- API с авторизацией (JWT) ---
  pipeline :api_auth do
    plug :accepts, ["json"]
    plug TodoTaskManagerWeb.Plugs.Auth
  end

  # --- веб-маршруты (например, домашняя страница) ---
  scope "/", TodoTaskManagerWeb do
    pipe_through :browser

    get "/", PageController, :home
  end

  # --- публичный API ---
  scope "/api", TodoTaskManagerWeb do
    pipe_through :api

    get "/ping", PingController, :index
    post "/register", RegistrationController, :create
    post "/login", SessionController, :create
  end

  # --- защищённый API ---
  scope "/api", TodoTaskManagerWeb do
    pipe_through :api_auth

    get "/tasks", TaskApiController, :index
    post "/tasks", TaskApiController, :create
    put "/tasks/:id", TaskApiController, :update
    delete "/tasks/:id", TaskApiController, :delete_task
  end

  # --- LiveDashboard (в разработке) ---
  if Application.compile_env(:todo_task_manager, :dev_routes) do
    import Phoenix.LiveDashboard.Router

    scope "/dev" do
      pipe_through :browser

      live_dashboard "/dashboard", metrics: TodoTaskManagerWeb.Telemetry
    end
  end

  # --- Swagger UI ---
  forward "/swagger", PhoenixSwagger.Plug.SwaggerUI,
    otp_app: :todo_task_manager,
    swagger_file: "swagger.json"

  # --- Swagger Info ---
  def swagger_info do
    %{
      info: %{
        version: "1.0.0",
        title: "Todo Task Manager API"
      },
      schemes: ["http"],
      basePath: "/",
      consumes: ["application/json"],
      produces: ["application/json"],
      securityDefinitions: %{
        Bearer: %{
          type: "apiKey",
          name: "authorization",
          in: "header"
        }
      }
    }
  end

  # --- Prometheus metrics ---
  if Mix.env() == :dev or Mix.env() == :prod do
    forward "/metrics", PromEx.Plug, prom_ex_module: TodoTaskManager.PromEx
  end
end
