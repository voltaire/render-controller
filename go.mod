module github.com/voltaire/render-controller

go 1.15

require (
	github.com/aws/aws-lambda-go v1.20.0
	github.com/aws/aws-sdk-go v1.35.23
	github.com/bsdlp/update-docker-image v0.0.0-20201105200224-b35146985022
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/machine v0.16.2
	github.com/golang/protobuf v1.4.3
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/twitchtv/twirp v7.1.0+incompatible
)

replace github.com/docker/machine v0.16.2 => github.com/bsdlp/machine v0.7.1-0.20201114195251-29ab5be05b0c
