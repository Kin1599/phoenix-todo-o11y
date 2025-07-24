defmodule TodoTaskManager.Repo.Migrations.CreateTasks do
  use Ecto.Migration

  def change do
    create table(:tasks) do
      add :title, :string, null: false
      add :description, :string
      add :status, :string
      add :user_id, references(:users, type: :binary_id), null: false
      timestamps()
    end

    create index(:tasks, [:user_id])
  end
end
