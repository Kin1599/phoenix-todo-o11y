FROM elixir:1.18.4-alpine

RUN apk add --no-cache git build-base nodejs npm openssl ncurses-libs

WORKDIR /app

RUN mix local.hex --force && \
    mix local.rebar --force

COPY mix.exs mix.lock ./

RUN mix deps.get

COPY . .

RUN mix deps.compile
RUN mix compile

WORKDIR /app/assets
RUN npm install
RUN npm run deploy

WORKDIR /app

RUN mix phx.digest

ENV MIX_ENV=prod

CMD ["mix", "phx.server"]