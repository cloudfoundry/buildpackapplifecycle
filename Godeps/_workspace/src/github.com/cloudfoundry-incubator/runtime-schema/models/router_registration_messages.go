package models

type RouterRegistrationMessage struct {
	Host string            `json:"host"`
	Port int               `json:"port"`
	Tags map[string]string `json:"tags"`
}

func (msg RouterRegistrationMessage) Component() string {
	return msg.Tags["component"]
}
