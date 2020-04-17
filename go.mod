module github.com/turbonomic/prometurbo

go 1.13

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/prometheus/common v0.9.1
	github.com/turbonomic/turbo-go-sdk v6.4.3+incompatible
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/turbonomic/turbo-go-sdk => ../turbo-go-sdk