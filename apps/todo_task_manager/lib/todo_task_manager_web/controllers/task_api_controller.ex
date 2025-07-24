defmodule TodoTaskManagerWeb.TaskApiController do
  use TodoTaskManagerWeb, :controller
  use PhoenixSwagger

  alias TodoTaskManager.Tasks
  alias TodoTaskManager.Tasks.Task

  swagger_path :index do
    get("/api/tasks")
    summary("List all tasks for user")
    description("Returns tasks of authenticated user")
    produces("application/json")
    security([%{Bearer: []}])
    response(200, "List of tasks", Schema.array(:Task))
    response(401, "Unauthorized")
  end

  def index(conn, _params) do
    user = conn.assigns.current_user
    tasks = Tasks.list_tasks(user.id)
    json(conn, tasks)
  end

  swagger_path :create do
    post("/api/tasks")
    summary("Create task")
    description("Create a new task for user")
    produces("application/json")
    security([%{Bearer: []}])
    parameter(:body, :body, :json, "Task data", required: true, schema: Schema.ref(:TaskCreate))
    response(200, "Created task", Schema.ref(:Task))
    response(401, "Unauthorized")
    response(422, "Unprocessable Entity")
  end

  def create(conn, params) do
    user = conn.assigns.current_user
    case Tasks.create_task(user.id, params) do
      {:ok, task} -> json(conn, task)
      {:error, changeset} ->
        conn
        |> put_status(:unprocessable_entity)
        |> json(%{errors: changeset.errors})
    end
  end

  swagger_path :update do
    put("/api/tasks/{id}")
    summary("Update task")
    description("Update task fields")
    produces("application/json")
    security([%{Bearer: []}])
    parameter(:id, :path, :string, "Task ID", required: true)
    parameter(:body, :body, :json, "Task update params", required: true, schema: Schema.ref(:TaskUpdate))
    response(200, "Updated task", Schema.ref(:Task))
    response(401, "Unauthorized")
    response(422, "Unprocessable Entity")
  end

  def update(conn, %{"id" => id} = params) do
    user = conn.assigns.current_user
    case Tasks.update_task(user.id, id, params) do
      {:ok, task} -> json(conn, task)
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "Task not found"}))
      {:error, changeset} -> put_status(conn, :unprocessable_entity) |> json(%{errors: changeset.errors})
    end
  end

  swagger_path :delete_task do
    PhoenixSwagger.Path.delete("/api/tasks/{id}")
    summary("Delete task")
    description("Delete a user's task")
    produces("application/json")
    security([%{Bearer: []}])
    parameter(:id, :path, :string, "Task ID", required: true)
    response(200, "OK", Schema.ref(:TaskDelete))
    response(401, "Unauthorized")
    response(404, "Not found")
  end

  def delete_task(conn, %{"id" => id}) do
    user = conn.assigns.current_user
    case Tasks.delete_task(user.id, id) do
      {:ok, _task} -> json(conn, %{message: "Task deleted"})
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "Task not found"}))
    end
  end

  def swagger_definitions do
    %{
      Task: swagger_schema do
        title("Task")
        description("Task schema")
        properties do
          id(:string, "Task ID")
          title(:string, "Title")
          description(:string, "Description")
          status(:string, "Status")
          user_id(:string, "User ID")
          inserted_at(:string, "Created at", format: :datetime)
          updated_at(:string, "Updated at", format: :datetime)
        end
        example(%{
          id: "a1b2c3d4-5678-90ab-cdef-1234567890ab",
          title: "Read docs",
          description: "Learn Phoenix Swagger",
          status: "pending",
          user_id: "b2a1c3d4-5678-90ab-cdef-1234567890ab",
          inserted_at: "2025-07-24T20:20:20Z",
          updated_at: "2025-07-24T20:20:20Z"
        })
      end,
      TaskCreate: swagger_schema do
        title("TaskCreate")
        description("Task creation schema")
        properties do
          title(:string, "Title", required: true)
          description(:string, "Description")
          status(:string, "Status")
        end
        example(%{
          title: "Read docs",
          description: "Learn Phoenix Swagger",
          status: "pending"
        })
      end,
      TaskUpdate: swagger_schema do
        title("TaskUpdate")
        description("Task update schema")
        properties do
          title(:string, "Title")
          description(:string, "Description")
          status(:string, "Status")
        end
        example(%{
          title: "Read docs UPDATED",
          status: "done"
        })
      end,
      TaskDelete: swagger_schema do
        title "TaskDelete"
        description "Task deletion confirmation"
        properties do
          message :string, "Confirmation message"
        end
        example %{
          message: "Task deleted"
        }
      end
    }
  end
end
