package utils

import (
	ghodssYaml "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

func MarshalYAMLToJSON(object interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return ghodssYaml.YAMLToJSON(bytes)

}
