package main

import (
	"fmt"
	"log"
	"os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)



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
	
	document := Document{}

	

	yaml_bytes, err := ioutil.ReadFile("test.yaml")

	err = yaml.Unmarshal([]byte(yaml_bytes), &document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- document:\n%v\n\n", document)

	d, err := yaml.Marshal(&document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))


	if len(os.Args) > 2 {
		fmt.Printf("%s\n", os.Args[1])
		
		repo, ok := document.Repositories[os.Args[1]]
		if !ok {
			fmt.Printf("repo %s not found\n", os.Args[1])
			os.Exit(1)
		}
		
		tag, ok := repo.Tags[os.Args[2]]
		if !ok {
			fmt.Printf("tag %s not found\n", os.Args[2])
			os.Exit(1)
		}
		
		
		git_clone_cmd := "git clone --recursive "
		if tag.Branch != "" {
			git_clone_cmd += "-b "+tag.Branch+" "
		}
		if tag.Tag != "" {
			git_clone_cmd += "-b "+tag.Tag+" "
		}
		git_clone_cmd += tag.Repository
		
		fmt.Printf("%s\n", git_clone_cmd)
	
	
	    dockerfile_array := strings.Split(tag.Dockerfile, "/")
	
		opt_f := ""
	
		dockerfile_filename := dockerfile_array[len(dockerfile_array)-1]
		if dockerfile_filename == "" {
		    fmt.Printf("Dockerfile not defined !?")
			os.Exit(1)
		} else if dockerfile_filename != "Dockerfile" {
			opt_f = " -f "+dockerfile_filename+" "
		} 
	
		dockerfile_path := ""
		for i := 0 ; i<len(dockerfile_array)-1 ; i++ {
			dockerfile_path += "/"+dockerfile_array[i]
		}
	
		last_slash := strings.LastIndexAny( tag.Repository, "/")
		suffix := tag.Repository[last_slash+1:]
		directory := strings.TrimSuffix(suffix , ".git")
	
		docker_build_cmd := "docker build --force-rm --no-cache --rm -t " + os.Args[1] + ":" + os.Args[2] + opt_f +" ./"+directory+"/"+dockerfile_path+"/"
		
		
	    fmt.Printf("%s\n", docker_build_cmd)
	
		
	}

}
