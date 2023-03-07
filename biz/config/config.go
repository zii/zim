package config

import (
	"zim.cn/base"
	"zim.cn/base/db"
	"zim.cn/base/objstorage"
	"zim.cn/base/objstorage/cos"
	"zim.cn/base/objstorage/minio"
	"zim.cn/base/objstorage/obs"
	"zim.cn/base/objstorage/oss"
	"zim.cn/base/redis"
	"zim.cn/biz/def"
	"zim.cn/biz/kafka"

	"github.com/BurntSushi/toml"
)

type Zim struct {
	MultiDC bool
}

type RPCX struct {
	Port int
}

type Kafka struct {
	Enable      bool
	Brokers     []string
	Level1Topic string
	Level1Group string
	Level2Topic string
	Level2Group string
}

type Oss struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
	BucketHost      string
	RoleArn         string
}

type Obs struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	Bucket     string
	BucketHost string
}

type Cos struct {
	AppId      string
	SecretID   string
	SecretKey  string
	Region     string
	Bucket     string
	BucketHost string
}

type Minio struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        string
	Bucket          string
	BucketHost      string
}

type ObjStorage struct {
	Platform string
}

type Config struct {
	Zim        *Zim
	Rpcx       *RPCX
	Mysql      []*db.Config    `json:"mysql"`
	Redis      []*redis.Config `json:"redis"`
	Kafka      *Kafka
	Oss        *Oss
	Obs        *Obs
	Cos        *Cos
	Minio      *Minio
	ObjStorage *ObjStorage
}

func LoadConfigFile(tomlfile string) *Config {
	config := Config{}
	_, err := toml.DecodeFile(tomlfile, &config)
	base.Raise(err)
	return &config
}

func (c *Config) InitZim() {
	if c.Zim == nil {
		return
	}
	def.UseMultiDC = c.Zim.MultiDC
}

func (c *Config) InitMysql() {
	if err := db.Install(c.Mysql); err != nil {
		panic(err)
	}
}

func (c *Config) InitRedis() {
	redis.Install(c.Redis)
}

func (c *Config) InitKafka() {
	if c.Kafka == nil {
		return
	}
	if c.Kafka.Enable {
		kafka.P1 = kafka.NewProducer(c.Kafka.Brokers, c.Kafka.Level1Topic)
		kafka.P2 = kafka.NewProducer(c.Kafka.Brokers, c.Kafka.Level2Topic)
	}
}

func (c *Config) InitOss() {
	if c.Oss == nil {
		return
	}
	oss.AccessKeyID = c.Oss.AccessKeyID
	oss.AccessKeySecret = c.Oss.AccessKeySecret
	oss.EndPoint = c.Oss.Endpoint
	oss.Bucket = c.Oss.Bucket
	oss.BucketHost = c.Oss.BucketHost
	oss.RoleArn = c.Oss.RoleArn
	oss.InitClient()
}

func (c *Config) InitObs() {
	if c.Obs == nil {
		return
	}
	obs.AccessKey = c.Obs.AccessKey
	obs.SecretKey = c.Obs.SecretKey
	obs.Endpoint = c.Obs.Endpoint
	obs.Bucket = c.Obs.Bucket
	obs.BucketHost = c.Obs.BucketHost
	obs.InitClient()
}

func (c *Config) InitCos() {
	if c.Cos == nil {
		return
	}
	cos.AppId = c.Cos.AppId
	cos.SecretID = c.Cos.SecretID
	cos.SecretKey = c.Cos.SecretKey
	cos.Region = c.Cos.Region
	cos.Bucket = c.Cos.Bucket
	cos.BucketHost = c.Cos.BucketHost
}

func (c *Config) InitMinio() {
	if c.Minio == nil {
		return
	}
	minio.AccessKeyID = c.Minio.AccessKeyID
	minio.SecretAccessKey = c.Minio.SecretAccessKey
	minio.Region = c.Minio.Region
	minio.Endpoint = c.Minio.Endpoint
	minio.Bucket = c.Minio.Bucket
	minio.BucketHost = c.Minio.BucketHost
}

func (c *Config) InitObjStorage() {
	if c.ObjStorage != nil {
		objstorage.Platform = c.ObjStorage.Platform
	}
	c.InitOss()
	c.InitObs()
	c.InitCos()
	c.InitMinio()
}

func (c *Config) Init() {
	c.InitZim()
	c.InitMysql()
	c.InitRedis()
	c.InitKafka()
	c.InitObjStorage()
}
