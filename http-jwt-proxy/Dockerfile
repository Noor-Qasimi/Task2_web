FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/proxy ./cmd/proxy \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/token ./cmd/token

FROM  alpine:3.22
COPY --from=build /out/proxy /bin/proxy
COPY --from=build /out/token /bin/token
EXPOSE 8080 8443
USER 65532:65532
CMD ["/bin/proxy"]
