defmodule TodoTaskManager.Tasks do
  import Ecto.Query

  alias TodoTaskManager.Tasks.Task
  alias TodoTaskManager.Repo

  def list_tasks(user_id) do
    Repo.all(from t in Task, where: t.user_id == ^user_id)
  end

  def get_task(user_id, id), do: Repo.get_by(Task, id: id, user_id: user_id)

  def create_task(user_id, attrs) do
    attrs = Map.put(attrs, "user_id", user_id)
    %Task{} |> Task.changeset(attrs) |> Repo.insert()
  end

  def update_task(user_id, id, attrs) do
    case get_task(user_id, id) do
      nil -> {:error, :not_found}
      task -> task |> Task.changeset(attrs) |> Repo.update()
    end
  end

  def delete_task(user_id, id) do
    case get_task(user_id, id) do
      nil -> {:error, :not_found}
      task -> Repo.delete(task)
    end
  end


  # Преобразование ID к integer
  defp to_integer(id) when is_binary(id), do: String.to_integer(id)
  defp to_integer(id), do: id
end
