package forefs

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

//func ReadYaml(s stores.Storer, name string, out interface{}) error {
//	data, err := Read(s, name)
//	if err != nil {
//		return err
//	}
//	return yaml.Unmarshal(data, out)
//}
//
//func WriteYaml(s stores.Storer, name string, in interface{}) error {
//	d, err := yaml.Marshal(in)
//	if err != nil {
//		return err
//	}
//	return Write(s, name, d)
//}

func ReadYamlFile(name string, out interface{}) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

func WriteYamlFile(name string, in interface{}) error {
	d, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, d, 0533)
}
