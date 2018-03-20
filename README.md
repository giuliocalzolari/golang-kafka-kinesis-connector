# kafka-kinesis

A simple command line tool to consume partitions of a topic and push to kinesis

### Installation

    go build
    ./kafka-kinesis -topic=test -region=eu-central-1 -stream=kafka-to-kinesis -offset=newest

### Usage

    # Minimum invocation
    ./kafka-kinesis -topic=test -brokers=kafka1:9092

    # It will pick up a KAFKA_PEERS environment variable
    export KAFKA_PEERS=kafka1:9092,kafka2:9092,kafka3:9092
    ./kafka-kinesis -topic=test

    # You can specify the offset you want to start at. It can be either
    # `oldest`, `newest`. The default is `newest`.
    ./kafka-kinesis -topic=test -offset=oldest
    ./kafka-kinesis -topic=test -offset=newest

    # You can specify the partition(s) you want to consume as a comma-separated
    # list. The default is `all`.
    ./kafka-kinesis -topic=test -partitions=1,2,3

    # Display all command line options
    ./kafka-kinesis -help

## Service

    cp kafka-kinesis-connector.service /etc/systemd/system/
    systemctl daemon-reload # Run if .service file has changed
    systemctl start kafka-kinesis (same with enable/disable/stop/restart/status)
