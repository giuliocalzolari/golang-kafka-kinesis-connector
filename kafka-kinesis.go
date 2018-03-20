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
	offset     = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest` default: newest")
	bulk       = flag.Int("bulk", 1, "number of recod to send to kinesis default:1")
	proctime   = flag.Int("proctime", 4, "processing time for kafka event default:4")
	verbose    = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	consumerGroup  = flag.String("group", "default-group", "The name of the consumer group, used for coordination and load balancing default:default-group")
	zookeeper      = flag.String("zookeeper", os.Getenv("ZOOKEEPER_PEERS"), "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181`)")
	zookeeperNodes []string

	stream = flag.String("stream", "your-stream", "your stream name")
	region = flag.String("region", os.Getenv("AWS_DEFAULT_REGION"), "AWS region")

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
		sarama.Logger = log.New(os.Stderr, "", log.LstdFlags)
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

	config.Offsets.ProcessingTimeout = 5 * time.Second

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
	bulkCount := 0

	bulk_buffer := make(map[int][]string)
	offsets := make(map[string]map[int32]int64)

	for message := range consumer.Messages() {
		if offsets[message.Topic] == nil {
			offsets[message.Topic] = make(map[int32]int64)
		}

		eventCount += 1
		if offsets[message.Topic][message.Partition] != 0 && offsets[message.Topic][message.Partition] != message.Offset-1 {
			log.Printf("Unexpected offset on %s:%d. Expected %d, found %d, diff %d.\n", message.Topic, message.Partition, offsets[message.Topic][message.Partition]+1, message.Offset, message.Offset-offsets[message.Topic][message.Partition]+1)
		}

		sarama.Logger.Printf("P:%d O:%d V:%s", message.Partition, message.Offset, string(message.Value))

		if *bulk <= 1 {

			_, err := kc.PutRecord(&kinesis.PutRecordInput{
				Data:         []byte(string(message.Value)),
				StreamName:   streamName,
				PartitionKey: aws.String(string(message.Partition)),
			})
			if err != nil {
				panic(err)
			}

		} else {
			bulk_buffer[bulkCount] = []string{string(message.Partition), string(message.Value)}
			bulkCount += 1

			if len(bulk_buffer) >= *bulk {

					sarama.Logger.Printf("send event %d to kinesis %s", len(bulk_buffer), *stream)
					// put 10 records using PutRecords API
					entries := make([]*kinesis.PutRecordsRequestEntry, len(bulk_buffer))
					for key, value := range bulk_buffer {
						entries[key] = &kinesis.PutRecordsRequestEntry{
							Data:         []byte(string(value[1])),
							PartitionKey: aws.String(value[0]),
						}
					}
					_, err := kc.PutRecords(&kinesis.PutRecordsInput{
						Records:    entries,
						StreamName: streamName,
					})
					if err != nil {
						fmt.Println(err)
						panic(err)
					}
					bulkCount = 0
					bulk_buffer = make(map[int][]string)

			}

		}

		offsets[message.Topic][message.Partition] = message.Offset
		consumer.CommitUpto(message)
	}

}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
