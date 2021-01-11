FROM hub.sinimini.com/docker/golang:1.14 AS builder
LABEL stage=builder
WORKDIR /workspace
COPY . .
RUN cd cmd/moneytree && go get && CGO_ENABLED=0 GOOS=linux go build -a -o ../../bin/moneytree

FROM hub.sinimini.com/docker/alpine AS final
WORKDIR /
COPY --from=builder /workspace/bin/moneytree .
RUN echo -n "{}" > config.yaml
CMD ["./moneytree server"]