# Build: um binário só, com public/ embutido, via trilha build.
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN go install github.com/emersonjoe/trilha/cmd/trilha@v0.139.0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN trilha check --fix && trilha build -o /out/app

# Runtime: imagem mínima, processo sem root, dados em volume.
FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/app /usr/local/bin/app
USER app
WORKDIR /var/lib/app
ENV DATA_DIR=/var/lib/app/data TRILHA_ADDR=":3000"
VOLUME /var/lib/app/data
EXPOSE 3000
ENTRYPOINT ["app"]
