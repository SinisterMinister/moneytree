FROM hub.sinimini.com/docker/golang:1.14 AS builder
LABEL stage=builder
WORKDIR /workspace
COPY . .
RUN go get && CGO_ENABLED=0 GOOS=linux go build -a -o bin/miraclegrow

FROM hub.sinimini.com/docker/alpine AS final
WORKDIR /
COPY --from=builder /workspace/bin/miraclegrow .
RUN echo -n "{}" > config.yaml
CMD ["./miraclegrow", "grow"]