FROM register.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.23.0 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /backend
COPY . .

RUN go build --trimpath --buildmode=plugin -o ./backend.so
RUN echo "THIS IS WORKING"

FROM register.heroiclabs.com/heroiclabs/nakama:3.23.0

COPY --from=builder /backend/backend.so ./modules
COPY --from=builder /backend/local.yml .
COPY --from=builder /backend/*.json ./modules