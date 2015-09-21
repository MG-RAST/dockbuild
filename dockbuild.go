package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
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
	Tags map[string]*Repo_tag `yaml:"tags"`
}

type Document struct {
	Repositories map[string]*Repository `yaml:"repositories"`
}

func getHeadCommitDate(git_user string, git_repo_name string) (tag_date string) {

	client := github.NewClient(nil)
	ref, _, err := client.Git.GetRef(git_user, git_repo_name, "refs/heads/master")
	if err != nil {
		fmt.Printf("Git.GetRef returned error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Git.GetRef returned: %+v \n", ref)
	sha := *ref.Object.SHA
	fmt.Printf("Commit SHA: %s\n", sha)

	commit, _, err := client.Git.GetCommit(git_user, git_repo_name, sha)
	if err != nil {
		fmt.Printf("Git.GetCommit returned error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Git.GetCommit returned: %+v\n", commit)
	author_date := commit.Author.Date
	fmt.Printf("author_date: %v\n", author_date)
	tag_date = fmt.Sprintf("%04d%02d%02d.%02d%02d", author_date.Year(), author_date.Month(), author_date.Day(), author_date.Hour(), author_date.Minute())
	return
}

func RunCommand(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	log.Info("executing: " + strings.Join(cmd.Args, " "))
	//log.Info("executing: " + cmd.Path + " " + strings.Join(cmd.Args, " "))
	//log.Info("executing: " + name + " " + strings.Join(arg, " "))
	//cmd := exec.Command(name, arg...)

	log.Debug(fmt.Sprintf("(RunCommand) cmd struct: %#v", cmd))

	cmdOutput := &bytes.Buffer{}
	cmdStderr := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdStderr

	if err = cmd.Start(); err != nil {
		log.Debug("(RunCommand) cmd.Start failed")
		return cmdOutput.Bytes(), cmdStderr.Bytes(), err
	}
	err = cmd.Wait()

	return cmdOutput.Bytes(), cmdStderr.Bytes(), err

}

func Parse_git_url(url string) (git_user string, git_repo_name string) {

	git_repo_trimmed := strings.TrimPrefix(url, "git@github.com:")
	git_repo_trimmed = strings.TrimPrefix(git_repo_trimmed, "https://github.com/")
	git_repo_trimmed = strings.TrimSuffix(git_repo_trimmed, ".git")
	fmt.Printf("git_repo_trimmed: %s\n", git_repo_trimmed)

	git_repo_trimmed_array := strings.Split(git_repo_trimmed, "/")

	if len(git_repo_trimmed_array) != 2 {
		fmt.Printf("parsing error: %s\n", url)
		os.Exit(1)
	}

	git_user = git_repo_trimmed_array[0]
	git_repo_name = git_repo_trimmed_array[1]
	return
}

func main() {
	log.SetLevel(log.DebugLevel)

	document := Document{}

	yaml_bytes, err := ioutil.ReadFile("test.yaml")

	err = yaml.Unmarshal([]byte(yaml_bytes), &document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	d, err := yaml.Marshal(&document)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- yaml file:\n%s\n\n", string(d))

	if len(os.Args) > 2 {
		image_repo := os.Args[1]
		image_tag := os.Args[2]

		fmt.Printf("image_repo: %s\n", image_repo)

		// find entry in yaml file
		repo, ok := document.Repositories[image_repo]
		if !ok {
			fmt.Printf("repo %s not found\n", image_repo)
			os.Exit(1)
		}

		tag, ok := repo.Tags[image_tag]
		if !ok {
			fmt.Printf("tag %s not found\n", image_tag)
			os.Exit(1)
		}

		git_user, git_repo_name := Parse_git_url(tag.Repository)

		log.Debug(git_user + " " + git_repo_name)

		// use date_str always, unless there is a real version
		//date_str = getHeadCommitDate(git_user, git_repo_name)

		//fmt.Printf("date_str: %s\n", date_str)

		// found entry, now build commands
		git_clone_args := []string{"clone", "--recursive"}
		if tag.Branch != "" {
			git_clone_args = append(git_clone_args, "-b", tag.Branch)
		}
		if tag.Tag != "" {
			git_clone_args = append(git_clone_args, "-b", tag.Tag)
		}
		git_clone_args = append(git_clone_args, tag.Repository)

		// git clone

		stdo, stde, err := RunCommand(exec.Command("git", git_clone_args...)) //"git", git_clone_args...
		log.Info("stdout: " + string(stdo))
		log.Info("stderr: " + string(stde))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)

		}

		// get commit date
		cmd := exec.Command("git", "log", "-1", "--pretty=format:\"%cd\"", "--date", "iso")
		cmd.Dir = git_repo_name
		stdo, stde, err = RunCommand(cmd)
		stdo_str := string(stdo) // example "2015-09-09 10:53:34 -0500"
		log.Info("stdout: " + stdo_str)
		log.Info("stderr: " + string(stde))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)

		}

		var validdate = regexp.MustCompile(`^\"[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)
		if !validdate.MatchString(stdo_str) {
			log.Fatal("Could not parse date: " + stdo_str)
			os.Exit(1)
		}

		date_str := stdo_str[1:5] + stdo_str[6:8] + stdo_str[9:11] + "." + stdo_str[12:14] + stdo_str[15:17] + stdo_str[18:20]
		log.Info("date_str: " + date_str)
		image_tag = date_str

		docker_build_args := []string{"build", "--force-rm", "--no-cache", "--rm", "-t", image_repo + ":" + image_tag}

		dockerfile_array := strings.Split(tag.Dockerfile, "/")

		dockerfile_filename := dockerfile_array[len(dockerfile_array)-1]
		if dockerfile_filename == "" {
			fmt.Printf("Dockerfile not defined !?")
			os.Exit(1)
		} else if dockerfile_filename != "Dockerfile" {
			docker_build_args = append(docker_build_args, "-f", dockerfile_filename)
		}

		dockerfile_path := ""
		for i := 0; i < len(dockerfile_array)-1; i++ {
			dockerfile_path += "/" + dockerfile_array[i]
		}

		last_slash := strings.LastIndexAny(tag.Repository, "/")
		suffix := tag.Repository[last_slash+1:]
		directory := strings.TrimSuffix(suffix, ".git")

		dockerfile_directory := " ./" + path.Join(directory, dockerfile_path) + "/"
		docker_build_args = append(docker_build_args, dockerfile_directory)

		stdo, stde, err = RunCommand(&exec.Cmd{Path: "docker", Args: docker_build_args})
		log.Info("stdout: " + string(stdo))
		log.Info("stderr: " + string(stde))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)

		}

	}

}

// gofmt -w . && go build . && ./dockbuild mgrast/v3-web develop
// curl  -X GET "https://api.github.com/repos/wgerlach/Skycore/git/refs/heads/master"
// show head commit : git rev-parse HEAD
// commit date: git log -1  --pretty=format:"%cd" --date=iso # returns 2015-09-18 23:22:36 -0500
