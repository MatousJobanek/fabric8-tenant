package utils

import (
	"gopkg.in/yaml.v2"
	ghodssYaml "github.com/ghodss/yaml"
)

func MarshalYAMLToJSON(object interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return ghodssYaml.YAMLToJSON(bytes)

}
