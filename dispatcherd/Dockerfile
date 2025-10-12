FROM golang:1.25-alpine AS build

WORKDIR /build

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" -o /build/bin/dispatcherd ./cmd/


FROM scratch
COPY --from=build /build/bin/dispatcherd /dispatcherd
ENTRYPOINT ["/dispatcherd"]