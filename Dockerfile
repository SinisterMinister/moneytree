FROM hub.sinimini.com/docker/golang:1.14 AS builder
LABEL stage=builder
WORKDIR /workspace
COPY . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a

FROM hub.sinimini.com/docker/alpine AS final
WORKDIR /
COPY --from=builder /workspace/moneytree .
RUN echo -n "{}" > config.yaml
CMD ["./moneytree"]