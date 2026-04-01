package valueobject

import "errors"

type ConfigSku struct {
	value string
}

func NewConfigSku(v string) (ConfigSku, error) {
	if v == "" {
		return ConfigSku{}, errors.New("config sku cannot be empty")
	}
	return ConfigSku{value: v}, nil
}

func (c ConfigSku) String() string {
	return c.value
}
