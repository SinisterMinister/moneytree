FROM golang:1.14 AS builder
ARG branch=master
LABEL stage=builder
WORKDIR /workspace
RUN git clone -b $branch https://github.com/SinisterMinister/moneytree.git && \
    cd moneytree/cmd/moneytree && CGO_ENABLED=0 GOOS=linux go build -a -o ../../moneytree

FROM alpine AS final
WORKDIR /
COPY --from=builder /workspace/moneytree .
RUN echo -n "{}" > config.yaml
CMD ["./moneytree", "server"]