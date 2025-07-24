defmodule TodoTaskManagerWeb.Plugs.Auth do
  import Plug.Conn

  alias TodoTaskManager.Guardian

  def init(opts), do: opts

  def call(conn, _opts) do
    with ["Bearer " <> token] <- get_req_header(conn, "authorization"),
         {:ok, claims} <- Guardian.decode_and_verify(token),
         {:ok, user} <- Guardian.resource_from_claims(claims) do
      assign(conn, :current_user, user)
    else
      _ -> send_unauthorized(conn)
    end
  end

  defp send_unauthorized(conn) do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(401, Jason.encode!(%{error: "unauthorized"}))
    |> halt()
  end
end
