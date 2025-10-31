FROM golang:1.25-alpine AS build

LABEL "org.opencontainers.image.description"="Dispatcherd is a message dispatching service that routes messages based on configurable rules to various dispatchers such as log files, counters, or email."

WORKDIR /build

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" -o /build/bin/dispatcherd ./cmd/


FROM scratch
COPY --from=build /build/bin/dispatcherd /dispatcherd
ENTRYPOINT ["/dispatcherd"]