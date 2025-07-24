defmodule TodoTaskManagerWeb.PingController do
  use TodoTaskManagerWeb, :controller
  use PhoenixSwagger

  swagger_path :index do
    get "/api/ping"
    summary "Ping health check"
    description "Returns status and timestamp"
    response 200, "OK", Schema.ref(:PingResponse)
  end

  def index(conn, _params) do
    json(conn, %{
      status: "ok",
      timestamp: DateTime.utc_now() |> DateTime.to_iso8601()
    })
  end

  def swagger_definitions do
    %{
      PingResponse: swagger_schema do
        title "PingResponse"
        properties do
          status :string, "Status", required: true
          timestamp :string, "Timestamp", required: true
        end
        example %{
          status: "ok",
          timestamp: "2024-07-24T20:20:20Z"
        }
      end
    }
  end
end
