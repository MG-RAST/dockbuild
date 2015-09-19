package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var data = `
a: Easy!
b:
  c: 2
  d: [3, 4]
`

type T struct {
	A string
	B struct {
		C int
		D []int ",flow"
	}
}

type Repo_tag struct {
	Repository string `yaml:"repository"`
	Dockerfile string `yaml:"dockerfile"`
	Branch     string `yaml:"branch"`
	Tag        string `yaml:"tag"`
}

type Repository struct {
	//Name string			`yaml:"name"`
	Tags map[string]*Repo_tag	`yaml:"tags"`
}

type Document struct {
	Repositories map[string]*Repository `yaml:"repositories"`
}

func main() {
	t := T{}
	document := Document{}

	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))

	yaml_bytes, err := ioutil.ReadFile("test.yaml")

	err = yaml.Unmarshal([]byte(yaml_bytes), &document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- document:\n%v\n\n", document)

	d, err = yaml.Marshal(&document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}
