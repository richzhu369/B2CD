/*******
*统一配置文件中心
*******/

package config

import (
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/shima-park/agollo"
	"gopkg.in/ini.v1"
	"reflect"
	"strings"
	"time"
	"unicode"
)

var g *Configurations

// 公共配置加`global:"true"`
type Configurations struct {
	ago            agollo.Agollo
	app            ApolloConfig // apollo配置文件
	Application    Application  // 应用配置
	LogConf        LogConf
	Mysql          Mysql
	CommonConfig   CommonConfig
	Redis          Redis
	Mongodb        MongoDBConfig `apollo:"mongodb"`
	RabbitMQ       RabbitMQConfig
	Zookeep        ZKConfig
	Consul         ConsulConfig
	Activimq       ActiveMQConfig
	DzWebApi       DzWebApi
	DzRfom         DzRfom
	Room           Room
	Omaha          Omaha
	ShortCard      ShortCard
	AofShortCard   AofShortCard
	AofomahaConfig AofomahaConfig
	SRD            SRD
	Admin          Admin
	//MinIO          MinIOConfig    `global:"true"`
	//AuthTidb       Mysql          `global:"true"`
	//MessageRedis   Redis          `global:"true"`
	//MessageEs      Elastic        `global:"true"`
	//MessageEsAll   Elastic        `global:"true"`
	//Kafka          Kafka          `global:"true"`
	AwsS3Config AwsS3Config `global:"true"`
	//Encryption     Encryption     `global:"true"`
	//ServiceConfig  ServiceConfig  `global:"true"`
	//EsIndex        EsIndex        `global:"true"`
	//KafkaTopicName KafkaTopicName `global:"true"`
}

type SRD struct {
	ValidTimes    string
	ValidDuration string
}

type AofomahaConfig struct {
	Port        string //对外提供对端口
	ServerPort  string //对内提供的端口
	ServiceType string
}

type DzRfom struct {
	//SocketPort  string `ini:"socket.port" json:"socket.port"` //对外提供对端口
	//Room        string `ini:"room" json:"room"`
	//CardControl int64  `ini:"card.control" json:",string"`
	ServiceType string
}

type ShortCard struct {
	ServiceType string
}

type AofShortCard struct {
	ServiceType string
}

type Omaha struct {
	ServiceType string
}

type Room struct {
	ServiceType string
}

type MinIOConfig struct {
	Bucket          string
	AccessKeyId     string
	SecretAccessKey string
	EndPoint        string
	MinioDomain     string
	UseSsl          bool `json:",string"`
}

type Mysql struct {
	Address      string //  // 主数据库地址
	AddressSlave string // 从数据库地址
	LogEnable    bool   `json:",string"`
	IdleConnect  int    `json:",string"`
	MaxConnect   int    `json:",string"`
	MaxLifeTime  int    `json:",string"`
}

type Redis struct {
	Host     string
	Auth     string
	PoolSize int `json:",string"`
}

type Elastic struct {
	Address           string
	AuthPass          string
	AuthUser          string
	EnableRequestBody bool `json:",string"`
}

type CommonConfig struct {
	VerifyProxyUrl    string
	ProxyURL          string
	ServerIP          string
	DBSecretKey       string `json:"DBSecretKey"`
	AuthFlag          bool   `json:",string"`
	NumberSplitTables string
}

type ServiceConfig struct {
	CryptoService string // 内网 host 消息加密解密服务
	SnowService   string // 内网 host 雪花id服务
	SecretData    string // 内网 host 应用加密服务
}

type Kafka struct {
	KafkaAddr                string
	UserKafkaAddr            string
	TopicLoginLog            string
	TopicMsgPush             string // 聊天消息发送
	SingleTopicMsgPushPrefix string // 单线
	GroupTopicMsgPushPrefix  string // 群组
	BotTopicMsgPushPrefix    string // 机器人
}

type EsIndex struct {
	UserIndexName  string
	BotIndexName   string
	GroupIndexName string
}

type KafkaTopicName struct {
	UserTopicName  string
	BotTopicName   string
	GroupTopicName string
}

type Common struct {
	EsAddr               string // es 地址
	EsAddrV2             string // es 地址 新es用于查询热点信息
	VenueGRPCAddr        string // 场馆微服务地址
	FileUploadBase       string
	SmsGateway           string
	FidUrl               string
	KafkaAddr            string
	SecretKey            string
	MasterOpenSecretKey  bool   `json:",string"` // 是否校验X-API-XXX请求头
	BytesMasterSecretKey string // X-API-XXX请求头密钥
	BytesMasterIV        string // X-API-XXX请求头加密向量
}

// aws s3上传服务参数 用于图片等文件上传功能
type AwsS3Config struct {
	Bucket                 string          // 创建的桶
	AccessKeyId            string          // AccessKeyId
	SecretAccessKey        string          // SecretAccessKey
	Region                 string          // 区域 如 us-east-1
	UploadS3BackendUrl     string          // s3后台上传地址 同 EndPoint
	UploadS3FrontendDomain string          // s3前端显示地址
	MaxFileSize            int64           `json:",string"` // 最大文件上传大小
	MaxVideoSize           int64           `json:",string"` // 视频最大
	StaticDomainFilters    string          // 过滤白名单添加域名
	SourceFilterMap        map[string]bool // 过滤白名单数据
	BackUploadMaxSize      int64           `json:",string"` // 后台上传大小设置 包含视频大小
	EndPoint               string
}

type Admin struct {
	Bucket                 string          `apollo:"as3.bucketName" json:"as3.bucketName"` // 创建的桶
	AccessKeyId            string          `apollo:"as3.accessKey" json:"as3.accessKey"`   // AccessKeyId
	SecretAccessKey        string          `apollo:"as3.secretKey" json:"as3.secretKey"`   // SecretAccessKey
	Region                 string          `apollo:"as3.region" json:"as3.region"`         // 区域 如 us-east-1
	UploadS3BackendUrl     string          // s3后台上传地址 同 EndPoint
	UploadS3FrontendDomain string          `apollo:"s3.host" json:"s3.host"`                        // s3前端显示地址
	MaxFileSize            int64           `apollo:"as3.maxFileSize" json:"as3.maxFileSize,string"` // 最大文件上传大小
	MaxVideoSize           int64           `json:",string"`                                         // 视频最大
	StaticDomainFilters    string          // 过滤白名单添加域名
	SourceFilterMap        map[string]bool // 过滤白名单数据
	BackUploadMaxSize      int64           `json:",string"` // 后台上传大小设置 包含视频大小
	EndPoint               string
}

// Apollo 从本地ini文件读取Apollo配置
type Apollo struct {
	Prefix  string `ini:"Prefix"`  // 站点前缀标识
	Dynamic bool   `ini:"Dynamic"` // 是否开启动态配置
	AppID   string `ini:"AppID"`   // 应用id
	Cluster string `ini:"Cluster"` // 不同的环境，cluster读取不同的配置
	Address string `ini:"Address"` // apollo ip地址
	Secret  string `ini:"Secret"`  // apollo配置密钥
	Backup  string `ini:"Backup"`  // 备份地址
}

type ApolloConfig struct {
	Apollo `ini:"Apollo"`
}

type Application struct {
	Port            string
	MessageHttpPort string
	AuthFlag        bool   `json:",string"` // 鉴权开关
	TokenKey        string // 解析 token 的 aes key
	TokenExpire     int64  `json:",string"` //token 过期时间
	AesKey          string
	Env             string
	ProxyConfig     string
	BizTypes        int    `json:",string"` // 业务类型数量
	NetName         string // 网卡命
	ZkIntLen        string // zk保存int的字节数
}

// LogConf 本地日志的配置
type LogConf struct {
	LogPath  string `ini:"LogPath"`
	LogLevel string `ini:"LogLevel"`
	LogType  string `ini:"LogType"`
}

type MongoDBConfig struct {
	Url      string `ini:"url"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

type RabbitMQConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

type ZKConfig struct {
	Host string `ini:"host"`
	Port string `ini:"port"`
}

type ConsulConfig struct {
	Host string `ini:"host"`
}

type ActiveMQConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	User     string
	Password string
}

type DzWebApi struct {
	HttpPort                    string `json:"http_port"`
	IP                          string `json:"ip"`
	MttServerIp                 string `json:"MttServerIp"`
	MttServerPort               string `json:"MttServerPort"`
	TokenEnable                 string `json:"tokenEnable"`           //是否启用
	TokenHeaderName             string `json:"tokenHeaderName"`       //签名对应的HTTP头部名称:token
	TokenDeviceHeaderName       string `json:"tokenDeviceHeaderName"` //签名对应的HTTP头部名称:设备ID
	TokenCharsetName            string `json:"tokenCharsetName"`      //字符集编码
	TokenWhitelistUrls          string `json:"tokenWhitelistUrls"`    //白名单URL列表
	TokenWatchPath              string `json:"tokenWatchPath"`
	TokenFrequencyEnable        string `json:"tokenFrequencyEnable"` //防频配置
	TokenFrequencyPerSec        string `json:"tokenFrequencyPerSec"`
	TokenFrequencyWhitelistUrls string `json:"tokenFrequencyWhitelistUrls"`
	TokenFrequencyMaxNum        string `json:"tokenFrequencyMaxNum"`
	TokenFrequencyDenySec       string `json:"tokenFrequencyDenySec"`
	ServiceType                 string `json:"ServiceType"`
}

// InitConfig 初始化配置
func InitConfig(confPath string) error {
	apo := new(ApolloConfig)
	if err := ini.MapTo(apo, confPath); err != nil {
		return err
	}

	g = new(Configurations)
	g.app = *apo

	err := InitApollo()
	if err != nil {
		return err
	}

	data, err := g.loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("------------------最后读取到的配置文件信息--------------------", string(data))

	err = json.Unmarshal(data, g)
	if err != nil {
		return err
	}

	result, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}

	fmt.Println("------------------------最后读取到的配置文件信息 序列化之后的配置------------------", string(result))

	return nil
}
func (c *Configurations) loadConfig() ([]byte, error) {

	//小写直接过滤
	t := reflect.TypeOf(c).Elem()

	var build strings.Builder
	build.WriteString("{")

	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		//小写开头过滤掉
		if len(name) > 0 && unicode.IsLower([]rune(name)[0]) {
			continue
		}

		var tmpName = name

		//判断标签是否包含前缀，获取到apollo内的配置
		if t.Field(i).Tag.Get("global") != "" {
			tmpName = c.app.Prefix + "." + name
		}
		if tmpName == "DzRfom" {
			tmpName = "dz-rfom"
		}
		Cjson := jsoniter.ConfigCompatibleWithStandardLibrary
		b, err := Cjson.Marshal(c.ago.GetNameSpace(tmpName))
		if err != nil {
			fmt.Println("配置反序列化失败")
			return nil, err
		}

		build.WriteString(`"`)
		build.WriteString(name)
		build.WriteString(`"`)
		build.WriteString(`:`)
		build.WriteString(string(b))
		build.WriteString(`,`)
	}

	build.WriteString("}")

	data := build.String()
	//去掉末尾逗号
	i := strings.LastIndexAny(data, `,`)

	data = data[0:i] + data[i+1:]

	return []byte(data), nil
}
func InitApollo() error {
	var opts []agollo.Option

	if g.app.Secret != "" {
		opts = append(opts, agollo.AccessKey(g.app.Secret))
	}

	if g.app.Backup != "" {
		opts = append(opts, agollo.BackupFile(g.app.Backup))
	}

	// 自动获取所有namespace的值
	opts = append(
		opts,
		agollo.AutoFetchOnCacheMiss(),
		agollo.FailTolerantOnBackupExists(),
		agollo.LongPollerInterval(time.Second*5),
	)

	var err error
	g.ago, err = agollo.New(g.app.Address, g.app.AppID, opts...)
	if err != nil {
		return err
	}

	return nil
}

// GetLogConf 获取日志的配置
func GetLogConf() LogConf {
	return g.LogConf
}

func GetSecretKey() string {
	return g.CommonConfig.DBSecretKey
}

func GetApplication() Application {
	return g.Application
}

type Encryption struct {
	AesKey            string `ini:"AesKey"`
	AesIv             string `ini:"AesIv"`
	AesInitStr        string
	AesInitStrEn      string
	SRAPublicKey      string
	SRAPrimeKey       string
	AesMessageDataIv  string
	AesMessageDataKey string
}

//	func GetEncryptionConfig() Encryption {
//		return g.Encryption
//	}
func GetConfig() *Configurations {
	return g
}
func GetAuthFlag() bool {
	return g.Application.AuthFlag
}
func GetTokenExpire() int64 {
	if g.Application.TokenExpire <= 0 {
		return 12
	}
	return g.Application.TokenExpire
}
func GetAesKey() string {
	return g.Application.AesKey
}

// GetEnv 获取环境信息
func GetEnv() string {
	return g.Application.Env
}

// // 替换ali 地址为s3地址
func GetAwsDomain() string {
	return g.AwsS3Config.UploadS3FrontendDomain
}

// // 添加过滤静态域名白名单 这里加的是存在的域名 替换成现有的域名
func GetSourceFilterMap() map[string]bool {
	return g.AwsS3Config.SourceFilterMap
}

//func GetAwsBackendDomain() string {
//	return g.AwsS3Config.UploadS3FrontendDomain
//}
