# kafka-kinesis

A simple command line tool to consume partitions of a topic and push to kinesis

### Installation

    go build


### Usage

    ./kafka-kinesis -topic=test -region=eu-central-1 -stream=kafka-to-kinesis -zookeeper=localhost:2181 -brokers=localhost:9092 -verbose


    Usage of ./kafka-kinesis:
      -brokers string
        	The comma separated list of brokers in the Kafka cluster (default "localhost:9092")
      -buffer-size int
        	The buffer size of the message channel. (default 256)
      -group string
        	The name of the consumer group, used for coordination and load balancing (default "default")
      -offset oldest
        	The offset to start with. Can be oldest, `newest` (default "newest")
      -partitions string
        	The partitions to consume, can be 'all' or comma-separated numbers (default "all")
      -region string
        	your AWS region (default "eu-west-1")
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
