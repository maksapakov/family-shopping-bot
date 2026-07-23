FROM golang:1.26-bookworm AS build
WORKDIR /src
LABEL authors="maksapakov"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /out/bot /bot

ENV DATABASE_PATH=/data/shopping.db

EXPOSE 8181

USER nonroot:nonroot

ENTRYPOINT ["/bot"]