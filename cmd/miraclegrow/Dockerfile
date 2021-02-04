FROM hub.sinimini.com/docker/golang:1.14 AS builder
ARG branch=master
LABEL stage=builder
WORKDIR /workspace
RUN git clone -b $branch https://github.com/SinisterMinister/moneytree.git && \
    cd moneytree/cmd/miraclegrow && CGO_ENABLED=0 GOOS=linux go build -a -o /workspace/miraclegrow

FROM hub.sinimini.com/docker/alpine AS final
WORKDIR /
COPY --from=builder /workspace/miraclegrow .
RUN echo -n "{}" > config.yaml
CMD ["./miraclegrow", "grow"]