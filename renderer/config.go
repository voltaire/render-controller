package renderer

type Config struct {
	AwsRegion          string `split_words:"true" default:"us-west-2"`
	AwsAccessKeyId     string `split_words:"true" required:"true"`
	AwsSecretAccessKey string `split_words:"true" required:"true"`

	SourceBucketName       string `default:"mc.sep.gg-backups"`
	SourceBucketAccountId  string `default:"006851364659"`
	SourceBucketPathPrefix string `default:"newworld"`

	DestinationBucketURI       string `split_words:"true" default:"s3://map-tonkat-su/"`
	DestinationBucketEndpoint  string `split_words:"true"`
	DestinationAccessKeyId     string `split_words:"true"`
	DestinationSecretAccessKey string `split_words:"true"`

	OverworldName     string `envconfig:"OVERWORLD_DIR" default:"pumpcraft"`
	NetherName        string `envconfig:"NETHER_DIR" default:"pumpcraft_nether"`
	TheEndName        string `envconfig:"THE_END_DIR" default:"pumpcraft_the_end"`
	RendererImage     string `default:"ghcr.io/voltaire/renderer:latest"`
	DiscordWebhookUrl string `split_words:"true"`

	RunnerName string `split_words:"true" default:"renderer"`

	GithubActionsPublicKey string `split_words:"true" required:"true"`
}
