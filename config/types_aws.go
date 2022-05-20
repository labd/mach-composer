package config

// AWSTFState AWS S3 bucket state backend configuration.
type AWSTFState struct {
	Bucket    string `yaml:"bucket"`
	KeyPrefix string `yaml:"key_prefix"`
	Region    string `yaml:"region"`
	RoleARN   string `yaml:"role_arn"`
	LockTable string `yaml:"lock_table"`
	Encrypt   bool   `yaml:"encrypt" default:"true"`
}

type SiteAWS struct {
	AccountID string `yaml:"account_id"`
	Region    string `yaml:"region"`

	DeployRoleName string            `yaml:"deploy_role_name"`
	ExtraProviders []AWSProvider     `yaml:"extra_providers"`
	DefaultTags    map[string]string `yaml:"default_tags"`
}

type AWSProvider struct {
	Name        string
	Region      string
	DefaultTags map[string]string `yaml:"default_tags"`
}

type AWSEndpoint struct {
	ThrottlingBurstLimit int  `yaml:"throttling_burst_limit"`
	ThrottlingRateLimit  int  `yaml:"throttling_rate_limit"`
	EnableCDN            bool `yaml:"enable_cdn"`
}
