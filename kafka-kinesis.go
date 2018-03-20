package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

var (
	brokerList = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topic      = flag.String("topic", "", "REQUIRED: the topic to consume")
	partitions = flag.String("partitions", "all", "The partitions to consume, can be 'all' or comma-separated numbers")
	offset     = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose    = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	bufferSize = flag.Int("buffer-size", 256, "The buffer size of the message channel.")

	consumerGroup  = flag.String("group", "default", "The name of the consumer group, used for coordination and load balancing")
	zookeeper      = flag.String("zookeeper", os.Getenv("ZOOKEEPER_PEERS"), "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181`)")
	zookeeperNodes []string

	stream = flag.String("stream", "your-stream", "your stream name")
	region = flag.String("region", "eu-west-1", "your AWS region")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	flag.Parse()

	if *brokerList == "" {
		printUsageErrorAndExit("You have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	}

	if *topic == "" {
		printUsageErrorAndExit("-topic is required")
	}

	if *zookeeper == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *verbose {
		sarama.Logger = logger
	}

	config := consumergroup.NewConfig()

	switch *offset {
	case "oldest":
		config.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Offsets.Initial = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}

	config.Offsets.ProcessingTimeout = 10 * time.Second

	zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(*zookeeper)

	kafkaTopics := strings.Split(*topic, ",")

	consumer, consumerErr := consumergroup.JoinConsumerGroup(*consumerGroup, kafkaTopics, zookeeperNodes, config)
	if consumerErr != nil {
		log.Fatalln(consumerErr)
	}

	s := session.New(&aws.Config{Region: aws.String(*region)})
	kc := kinesis.New(s)
	streamName := aws.String(*stream)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if err := consumer.Close(); err != nil {
			sarama.Logger.Println("Error closing the consumer", err)
		}
	}()

	go func() {
		for err := range consumer.Errors() {
			log.Println(err)
		}
	}()

	eventCount := 0
	offsets := make(map[string]map[int32]int64)

	for message := range consumer.Messages() {
		if offsets[message.Topic] == nil {
			offsets[message.Topic] = make(map[int32]int64)
		}

		eventCount += 1
		if offsets[message.Topic][message.Partition] != 0 && offsets[message.Topic][message.Partition] != message.Offset-1 {
			log.Printf("Unexpected offset on %s:%d. Expected %d, found %d, diff %d.\n", message.Topic, message.Partition, offsets[message.Topic][message.Partition]+1, message.Offset, message.Offset-offsets[message.Topic][message.Partition]+1)
		}

		fmt.Printf("Partition:%d\t", message.Partition)
		fmt.Printf("Offset:%d\t", message.Offset)
		// fmt.Printf("Key:\t%s\n", string(msg.Key))
		fmt.Printf("Value:%s ", string(message.Value))

		_, err := kc.PutRecord(&kinesis.PutRecordInput{
			Data:         []byte(string(message.Value)),
			StreamName:   streamName,
			PartitionKey: aws.String(string(message.Partition)),
		})
		if err != nil {
			panic(err)
		}
		// Simulate processing time
		// time.Sleep(10 * time.Millisecond)

		offsets[message.Topic][message.Partition] = message.Offset
		consumer.CommitUpto(message)
	}

	log.Printf("Processed %d events.", eventCount)
	log.Printf("%+v", offsets)
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
