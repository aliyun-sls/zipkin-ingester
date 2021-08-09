package configure

import "github.com/spf13/viper"

type Configuration struct {
	BootstrapServers string
	GroupID          string
	AutoOffsetRest   string
	Topic            []string

	Project      string
	Instance     string
	AccessKey    string
	AccessSecret string
	Endpoint     string
}

func (c *Configuration) InitFromViper(v *viper.Viper) {
	c.BootstrapServers = v.GetString("kafka_bootstrap_services")
	c.GroupID = v.GetString("kafka_consumer_group")
	c.GroupID = v.GetString("kafka_topic")

	c.Project = v.GetString("project")
	c.Instance = v.GetString("instance")
	c.AccessKey = v.GetString("access_key")
	c.AccessSecret = v.GetString("access_secret")
	c.Endpoint = v.GetString("endpoint")
}
