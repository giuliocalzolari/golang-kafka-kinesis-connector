# kafka-kinesis
[![Release](https://img.shields.io/github/release/giuliocalzolari/golang-kafka-kinesis-connector.svg?style=flat-square)](https://github.com/giuliocalzolari/golang-kafka-kinesis-connector/releases/latest)
[![Travis](https://img.shields.io/travis/giuliocalzolari/golang-kafka-kinesis-connector.svg?style=flat-square)](https://travis-ci.org/giuliocalzolari/golang-kafka-kinesis-connector)
[![Go Report Card](https://goreportcard.com/badge/github.com/giuliocalzolari/golang-kafka-kinesis-connector?style=flat-square)](https://goreportcard.com/report/github.com/giuliocalzolari/golang-kafka-kinesis-connector)


A simple command line tool to consume partitions of a topic and push to kinesis

### Installation

    git clone git@github.com:giuliocalzolari/golang-kafka-kinesis-connector.git
    cd ./golang-kafka-kinesis-connector/
    go build


### Usage

    $ ./kafka-kinesis -topic=test -region=eu-central-1 -stream=kafka-to-kinesis -zookeeper=localhost:2181 -brokers=localhost:9092 -verbose

    $ ./kafka-kinesis -help

    Usage of ./kafka-kinesis:
      -brokers string
        	The comma separated list of brokers in the Kafka cluster (default "localhost:9092")
      -bulk int
        	number of recod to send to kinesis max value:500 (default 1)
      -group string
        	The name of the consumer group, used for coordination and load balancing (default "default-group")
      -offset oldest
        	The offset to start with. Can be `oldest`, `newest` (default "newest")
      -proctime int
        	processing time for kafka event (default 4)
      -region string
        	AWS region (default "eu-central-1")
      -stream string
        	your stream name (default "your-stream")
      -topic string
        	REQUIRED: the topic to consume
      -verbose
        	Whether to turn on sarama logging
      -zookeeper zookeeper1.local:2181,zookeeper2.local:2181
        	A comma-separated Zookeeper connection string (e.g. zookeeper1.local:2181,zookeeper2.local:2181)

## Service

    cp kafka-kinesis-connector.service /etc/systemd/system/
    systemctl daemon-reload # Run if .service file has changed
    systemctl start kafka-kinesis (same with enable/disable/stop/restart/status)


## Local deployment
if you want to test the code I strongly suggest to use Docker to build a local kafka instance using the following command

    $ docker pull spotify/kafka
    $ docker run -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST="kafka" --env ADVERTISED_PORT=9092 --name kafka spotify/kafka
  
