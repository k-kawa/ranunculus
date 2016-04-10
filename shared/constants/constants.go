package constants

//go:generate stringer -type=ContextKey
type ContextKey int

const (
	CtxCliContext ContextKey = iota
	CtxRedis
)

type EnvKey string

const (
	EnvRedisAddr     = "RedisAddr"
	EnvRedisPassword = "RedisPassword"
	EnvRedisDb       = "RedisDb"
	EnvInQueueUrl    = "InQueueUrl"
	EnvOutQueueUrl   = "OutQueueUrl"
	EnvAwsAccessKey  = "AwsAccessKey"
	EnvAwsSecretKey  = "AwsSecretKey"
	EnvAwsRegion     = "AwsRegion"
	EnvWaitInterval  = "WaitInterval"
)
