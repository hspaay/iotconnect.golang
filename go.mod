module github.com/hspaay/iotconnect.golang

go 1.13

// replace github.com/hspaay/iotconnect.golang => ../iotconnect.golang

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	gopkg.in/yaml.v2 v2.2.2
)

// replace github.com/hspaay/iotconnect.golang/standard => ../standard
