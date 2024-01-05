module github.com/andrdru/joplin-auto

go 1.21.5

replace github.com/andrdru/joplin-auto/joplin_provider v0.0.0 => ./joplin_provider

require (
	github.com/andrdru/go-template/graceful v0.1.0
	github.com/andrdru/joplin-auto/joplin_provider v0.0.0
	github.com/google/uuid v1.5.0
	github.com/robfig/cron/v3 v3.0.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go v1.49.15 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
