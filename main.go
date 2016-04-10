package main

import (
	"github.com/k-kawa/ranunculus/commands"
	"github.com/k-kawa/ranunculus/shared/constants"
	"github.com/codegangsta/cli"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"gopkg.in/redis.v3"
	"log"
	"os"
)

func NewApp() (app *cli.App) {
	app = cli.NewApp()

	app.Name = "ch"
	app.Usage = "Crawler helper"

	c := context.Background()

	flags := []cli.Flag{
		cli.StringFlag{
			Name:   constants.EnvAwsAccessKey,
			EnvVar: constants.EnvAwsAccessKey,
		},
		cli.StringFlag{
			Name:   constants.EnvAwsSecretKey,
			EnvVar: constants.EnvAwsSecretKey,
		},
		cli.StringFlag{
			Name:   constants.EnvAwsRegion,
			EnvVar: constants.EnvAwsRegion,
		},
		cli.StringFlag{
			Name:   constants.EnvInQueueUrl,
			EnvVar: constants.EnvInQueueUrl,
		},
		cli.StringFlag{
			Name:   constants.EnvOutQueueUrl,
			EnvVar: constants.EnvOutQueueUrl,
		},
		cli.StringFlag{
			Name:   constants.EnvRedisAddr,
			EnvVar: constants.EnvRedisAddr,
		},
		cli.StringFlag{
			Name:   constants.EnvRedisDb,
			EnvVar: constants.EnvRedisDb,
		},
		cli.StringFlag{
			Name:   constants.EnvRedisPassword,
			EnvVar: constants.EnvRedisPassword,
		},
		cli.StringFlag{
			Name:   constants.EnvWaitInterval,
			EnvVar: constants.EnvWaitInterval,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "Start process",
			Action: wrapCliContext(
				c,
				wrapRedis(commands.Start),
			),
			Flags: flags,
		},
	}

	return
}

func main() {
	godotenv.Load()
	app := NewApp()
	app.Run(os.Args)
}

func wrapCliContext(c context.Context, next func(c context.Context)) func(*cli.Context) {
	return func(cliContext *cli.Context) {
		next(context.WithValue(c, constants.CtxCliContext, cliContext))
	}
}

func wrapRedis(next func(c context.Context)) func(context.Context) {
	return func(ctx context.Context) {
		c := ctx.Value(constants.CtxCliContext).(*cli.Context)

		redisClient := redis.NewClient(&redis.Options{
			Addr:     c.String(constants.EnvRedisAddr),
			Password: c.String(constants.EnvRedisPassword), // no password set
			DB:       int64(c.Int(constants.EnvRedisDb)),   // use default DB
		})

		_, err := redisClient.Ping().Result()
		if err != nil {
			log.Fatalf("Failed to connect Redis %s", err.Error())
		}
		next(context.WithValue(ctx, constants.CtxRedis, redisClient))
	}
}
