defmodule TodoTaskManager.Repo.Migrations.CreateUsers do
  use Ecto.Migration

  def change do
    create table(:users, primary_key: true) do
      add :email, :string, null: false
      add :password_hash, :string, null: false
      timestamps()
    end

    create unique_index(:users, [:email])
  end
end
