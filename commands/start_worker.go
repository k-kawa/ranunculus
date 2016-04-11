package commands

import (
	"github.com/k-kawa/ranunculus/shared/constants"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/codegangsta/cli"
	"golang.org/x/net/context"
	"gopkg.in/redis.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type InQueueMessage struct {
	Url     string
	Headers map[string]string
	Depth   int
}

type OutQueueMessage struct {
	HasBody    bool
	Url        string
	Depth      int
	StatusCode int
	Headers    map[string][]string
}

func newOutQueueMessage(res *http.Response, inMsg *InQueueMessage) *OutQueueMessage {
	if res == nil {
		return &OutQueueMessage{
			HasBody: false,
			Url:       inMsg.Url,
			Depth:     inMsg.Depth,
		}
	}
	return &OutQueueMessage{
		HasBody:  true,
		Url:        inMsg.Url,
		Depth:      inMsg.Depth,
		StatusCode: res.StatusCode,
		Headers:    res.Header,
	}
}

type RedisObject struct {
	Requested  bool
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

func newRedisObject(res *http.Response) []byte {
	if res == nil {
		return make([]byte, 0)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf(err.Error())
		return make([]byte, 0)
	}
	return data
}

func StartWorker(ctx context.Context) {

	c := ctx.Value(constants.CtxCliContext).(*cli.Context)

	awsAccessKeyId := c.String(constants.EnvAwsAccessKey)
	awsSecretAccessKey := c.String(constants.EnvAwsSecretKey)
	awsRegion := c.String(constants.EnvAwsRegion)

	awsSession := session.New(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			awsAccessKeyId,
			awsSecretAccessKey,
			"",
		),
	})

	svc := sqs.New(awsSession)
	redisClient := ctx.Value(constants.CtxRedis).(*redis.Client)

	receiveParam := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.String(constants.EnvInQueueUrl)), // Required
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(20),
	}

	done := make(chan struct{})
	taskChan := make(chan int)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()

		for {
		INTERVAL_LOOP:
			for {
				select {
				case <-done:
					return
				case taskChan <- 1:
					break INTERVAL_LOOP
				}
			}

			// Fetch a URL from InQueue
			respReceiveInQueue, err := svc.ReceiveMessage(receiveParam)

			if err != nil {
				log.Println(err.Error())
			}

			if len(respReceiveInQueue.Messages) == 0 {
				continue
			}
			firstMessage := respReceiveInQueue.Messages[0]
			var inMsg InQueueMessage

			err = json.Unmarshal([]byte(*firstMessage.Body), &inMsg)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			// Send a request to the URL
			req, err := http.NewRequest("GET", inMsg.Url, nil)
			if err != nil {
				log.Println(err.Error())
			}
			for headerName, headerValue := range inMsg.Headers {
				req.Header.Add(headerName, headerValue)
			}

			log.Printf(inMsg.Url)

			httpRes, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf(err.Error())
			}

			// Delete the URL from InQueue
			deleteParams := &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.String(constants.EnvInQueueUrl)),
				ReceiptHandle: firstMessage.ReceiptHandle,
			}
			_, err = svc.DeleteMessage(deleteParams)
			if err != nil {
				log.Printf(err.Error())
			}

			// Write the header of the response to OutQueue
			msgByte, err := json.Marshal(newOutQueueMessage(httpRes, &inMsg))
			if err != nil {
				log.Printf(err.Error())
			}
			params := &sqs.SendMessageInput{
				MessageBody:  aws.String(string(msgByte)),                    // Required
				QueueUrl:     aws.String(c.String(constants.EnvOutQueueUrl)), // Required
				DelaySeconds: aws.Int64(1),
			}
			_, err = svc.SendMessage(params)

			if err != nil {
				log.Printf(err.Error())
			}

			// Write the body of the response to Redis
			_, err = redisClient.Set(inMsg.Url, newRedisObject(httpRes), 2*time.Hour).Result()
			if err != nil {
				log.Printf(err.Error())
			}
		}
	}()

	go controlInterval(wg, done, taskChan, time.Duration(c.Int(constants.EnvWaitInterval)))

	waitSignal(done)

	wg.Wait()
}

func controlInterval(wg sync.WaitGroup, done chan struct{}, in chan int, interval time.Duration) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-time.After(interval * time.Second):
			select {
			case <-in:
			default:
			}
		case <-done:
			return
		}
	}
}

func waitSignal(done chan struct{}) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
	)
	for {
		switch <-signalChan {
		default:
			log.Printf("Receive Signal")
			close(done)
			return
		}
	}
}
