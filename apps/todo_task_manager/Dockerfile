# Используем официальный образ Elixir с Node.js
FROM elixir:1.18.4-alpine as builder

# Устанавливаем зависимости
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем Hex и Rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем конфигурацию Mix
COPY mix.exs mix.lock ./
RUN mix deps.get

# Копируем остальной код
COPY . .

# Компилируем зависимости и приложение
RUN mix deps.compile
RUN mix compile

# Собираем фронтенд-ассеты
COPY assets ./assets
RUN cd assets && npm install && npm run deploy
RUN mix phx.digest

# Финальный образ
FROM elixir:1.15-slim

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/_build /app/_build
COPY --from=builder /app/deps /app/deps
COPY --from=builder /app/priv /app/priv
COPY --from=builder /app/config /app/config
COPY --from=builder /app/lib /app/lib
COPY --from=builder /app/mix.exs /app/mix.exs
COPY --from=builder /app/mix.lock /app/mix.lock

# Устанавливаем переменные окружения
ENV MIX_ENV=prod

# Запускаем сервер
CMD ["mix", "phx.server"]