defmodule TodoTaskManager.Tasks.Task do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  @derive {Jason.Encoder, only: [:id, :title, :description, :status, :user_id, :inserted_at, :updated_at]}
  schema "tasks" do
    field :title, :string
    field :description, :string
    field :status, :string

    belongs_to :user, TodoTaskManager.Accounts.User, type: :binary_id

    timestamps()
  end

  def changeset(task, attrs) do
    task
    |> cast(attrs, [:title, :description, :status, :user_id])
    |> validate_required([:title, :user_id])
  end
end
