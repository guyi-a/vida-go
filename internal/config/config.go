package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	MinIO         MinIOConfig         `mapstructure:"minio"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Agent         AgentConfig         `mapstructure:"agent"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Log           LogConfig           `mapstructure:"log"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Mode    string `mapstructure:"mode"`
	Port    int    `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// DSN 返回PostgreSQL连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回Redis地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string   `mapstructure:"endpoint"`
	AccessKey string   `mapstructure:"access_key"`
	SecretKey string   `mapstructure:"secret_key"`
	UseSSL    bool     `mapstructure:"use_ssl"`
	Buckets   []string `mapstructure:"buckets"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string          `mapstructure:"brokers"`
	Topics  map[string]string `mapstructure:"topics"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Hosts []string          `mapstructure:"hosts"`
	Index map[string]string `mapstructure:"index"`
}

// AgentConfig Agent服务配置
type AgentConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"` // 秒
}

// TimeoutDuration 返回超时时间
func (a *AgentConfig) TimeoutDuration() time.Duration {
	return time.Duration(a.Timeout) * time.Second
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// ExpireDuration 返回过期时间
func (j *JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(j.ExpireHours) * time.Hour
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// 全局配置实例
var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 设置配置文件类型
	v.SetConfigType("yaml")

	// 读取环境变量
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 保存到全局变量
	globalConfig = &cfg

	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, please call Load() first")
	}
	return globalConfig
}

// GetApp 获取应用配置
func GetApp() *AppConfig {
	return &Get().App
}

// GetDatabase 获取数据库配置
func GetDatabase() *DatabaseConfig {
	return &Get().Database
}

// GetRedis 获取Redis配置
func GetRedis() *RedisConfig {
	return &Get().Redis
}

// GetMinIO 获取MinIO配置
func GetMinIO() *MinIOConfig {
	return &Get().MinIO
}

// GetKafka 获取Kafka配置
func GetKafka() *KafkaConfig {
	return &Get().Kafka
}

// GetElasticsearch 获取Elasticsearch配置
func GetElasticsearch() *ElasticsearchConfig {
	return &Get().Elasticsearch
}

// GetAgent 获取Agent配置
func GetAgent() *AgentConfig {
	return &Get().Agent
}

// GetJWT 获取JWT配置
func GetJWT() *JWTConfig {
	return &Get().JWT
}

// GetLog 获取日志配置
func GetLog() *LogConfig {
	return &Get().Log
}
