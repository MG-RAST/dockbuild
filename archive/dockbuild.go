package main

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/vaughan0/go-ini"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Repo_tag struct {
	Repository string `yaml:"repository"`
	Dockerfile string `yaml:"dockerfile"`
	Branch     string `yaml:"branch"` // takes tag as argument
	Commit     string `yaml:"commit"`
	Recursive  string `yaml:"recursive"`
}

type Repository struct {
	//Name string			`yaml:"name"`
	Tags map[string]*Repo_tag `yaml:"tags"`

	// To be used as defaults
	Repository string `yaml:"repository"`
	Dockerfile string `yaml:"dockerfile"`
	Branch     string `yaml:"branch"` // takes tag as argument
	Commit     string `yaml:"commit"`
	Recursive  string `yaml:"recursive"`
}

type Document struct {
	Repositories map[string]*Repository `yaml:"repositories"`
}

//func getHeadCommitDate(git_user string, git_repo_name string) (tag_date string) {
//    // "github.com/google/go-github/github"
//	client := github.NewClient(nil)
//	ref, _, err := client.Git.GetRef(git_user, git_repo_name, "refs/heads/master")
//	if err != nil {
//		fmt.Printf("Git.GetRef returned error: %v", err)
//		os.Exit(1)
//	}
//	fmt.Printf("Git.GetRef returned: %+v \n", ref)
//	sha := *ref.Object.SHA
//	fmt.Printf("Commit SHA: %s\n", sha)

//	commit, _, err := client.Git.GetCommit(git_user, git_repo_name, sha)
//	if err != nil {
//		fmt.Printf("Git.GetCommit returned error: %v", err)
//		os.Exit(1)
//	}
//	fmt.Printf("Git.GetCommit returned: %+v\n", commit)
//	author_date := commit.Author.Date
//	fmt.Printf("author_date: %v\n", author_date)
//	tag_date = fmt.Sprintf("%04d%02d%02d.%02d%02d", author_date.Year(), author_date.Month(), author_date.Day(), author_date.Hour(), author_date.Minute())
//	return
//}

func RunCommand(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	log.Info("executing: " + strings.Join(cmd.Args, " "))

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

func isTrue(bool_str string) bool {
	true_array := []string{"y", "true", "yes", "on"}
	bool_low := strings.ToLower(bool_str)
	for _, value := range true_array {
		if value == bool_low {
			return true
		}
	}
	return false
}

func Git_clone(base_dir string, repo_name string, tag *Repo_tag) (err error) {

	git_repo_dir := path.Join(base_dir, repo_name)

	// if it exists it should be empty
	_, err = os.Stat(git_repo_dir)
	if err != nil {
		os.Remove(git_repo_dir)
	}

	git_clone_args := []string{"clone"}
	if tag.Branch != "" {
		git_clone_args = append(git_clone_args, "-b", tag.Branch)
	}

	git_clone_args = append(git_clone_args, tag.Repository)

	// git clone
	cmd := exec.Command("git", git_clone_args...)
	cmd.Dir = base_dir
	log.Debug("work directory: " + base_dir)
	stdo, stde, err := RunCommand(cmd) //"git", git_clone_args...
	log.Debug("stdout: " + string(stdo))
	log.Debug("stderr: " + string(stde))
	if err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)

	}

	if tag.Commit != "" {
		cmd = exec.Command("git", "checkout", tag.Commit)
		cmd.Dir = git_repo_dir
		stdo, stde, err = RunCommand(cmd)
		log.Debug("stdout: " + string(stdo))
		log.Debug("stderr: " + string(stde))
		if err != nil {
			return err
		}
	}

	if !isTrue(tag.Recursive) {
		return
	}

	gitmodules_filename := path.Join(git_repo_dir, ".gitmodules")
	_, err = os.Stat(gitmodules_filename)
	if err != nil {
		return
	}

	log.Debug("found .gitmodules")

	ini_object, err := ini.LoadFile(gitmodules_filename)
	if err != nil {
		log.Errorf("could not read ini file %s", gitmodules_filename)
		os.Exit(1)
	}

	for name, section := range ini_object { // name, section
		log.Debugf("Section name: %s\n", name)

		submodule_path, ok := section["path"]
		if !ok {
			return errors.New("Key \"path\" not found")
		}
		submodule_url, ok := section["url"]
		if !ok {
			return errors.New("Key \"url\" not found")
		}
		submodule_branch, ok := section["branch"]
		if !ok {
			return errors.New("Key \"branch\" not found")
		}

		// extract commit
		//git ls-tree master <path>
		// example output: "160000 commit fdff68fdbd694d293a0bdf3c20ae3f6284a9478e	AWE"
		cmd = exec.Command("git", "ls-tree", "master", submodule_path)
		cmd.Dir = git_repo_dir
		stdo, stde, err = RunCommand(cmd)
		log.Debug("stdout: " + string(stdo))
		log.Debug("stderr: " + string(stde))
		if err != nil {
			return err
		}
		fields := strings.Fields(string(stdo))
		commit := fields[2]
		log.Debug("commit: " + commit)

		if len(commit) != 40 {
			return errors.New("Commit hash does not have length 40 :" + commit)
		}

		submodule_basepath, submodule_reponame := path.Split(submodule_path)

		err = Git_clone(path.Join(git_repo_dir, submodule_basepath), submodule_reponame, &Repo_tag{Repository: submodule_url, Branch: submodule_branch, Commit: commit, Recursive: tag.Recursive})
		if err != nil {
			return
		}

	}

	return
}

func read_yaml(yaml_file string) Document {

	document := Document{}

	yaml_bytes, err := ioutil.ReadFile(yaml_file)
	if err != nil {
		log.Fatalf("error: %v", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal([]byte(yaml_bytes), &document)
	if err != nil {
		log.Fatalf("error: %v", err)
		os.Exit(1)
	}

	d, err := yaml.Marshal(&document)
	if err != nil {
		log.Fatalf("error: %v", err)
		os.Exit(1)
	}
	log.Debugf("--- yaml file:\n%s\n\n", string(d))
	return document
}

func dockbuild(document Document, image_repo string, image_tag string) (err error) {

	log.Infof("image_repo: %s\n", image_repo)

	// find entry in yaml file
	repo, ok := document.Repositories[image_repo]
	if !ok {
		return errors.New(fmt.Sprintf("repo %s not found\n", image_repo))
	}

	tag, ok := repo.Tags[image_tag]
	if !ok {

		//return errors.New(fmt.Sprintf("tag %s not found\n", image_tag))
		log.Infof("dockbuild tag %s not found, will use github tag \n", image_tag)
		tag = &Repo_tag{Repository: repo.Repository, Branch: image_tag, Commit: repo.Commit, Recursive: repo.Recursive}
	}

	// inherit values from parent if not defined
	if tag.Repository == "" {
		tag.Repository = repo.Repository
	}
	if tag.Repository == "" {
		log.Fatal("Repository undefined !")
		os.Exit(1)
	}

	if tag.Branch == "" {
		tag.Branch = repo.Branch
	}
	if tag.Dockerfile == "" {
		tag.Dockerfile = repo.Dockerfile
	}
	if tag.Commit == "" {
		tag.Commit = repo.Commit
	}
	if tag.Recursive == "" {
		tag.Recursive = repo.Recursive
	}

	git_user, git_repo_name := Parse_git_url(tag.Repository)

	log.Infof("Found repository %s/%s", git_user, git_repo_name)

	// TODO clean dockbuild_*

	glob_old_dirs := path.Join(os.TempDir() + "dockbuild_*")
	log.Debug("glob_old_dirs: " + glob_old_dirs)
	old_tmp_dirs, _ := filepath.Glob(glob_old_dirs)
	for _, dir := range old_tmp_dirs {
		log.Debug("deleting " + dir)
		os.RemoveAll(dir)
	}

	tempdir, err := ioutil.TempDir("", "dockbuild_")
	if err != nil {
		log.Errorf("error: %v\n", err)
		return
	}
	log.Info("created temp dirctory: " + tempdir)

	git_repo_dir := path.Join(tempdir, git_repo_name)
	// use date_str always, unless there is a real version
	//date_str = getHeadCommitDate(git_user, git_repo_name)

	//fmt.Printf("date_str: %s\n", date_str)

	// found entry, now build commands

	err = Git_clone(tempdir, git_repo_name, tag)

	// get commit date
	cmd := exec.Command("git", "log", "-1", "--pretty=format:\"%cd\"", "--date", "iso")
	cmd.Dir = git_repo_dir
	stdo, stde, err := RunCommand(cmd)
	stdo_str := string(stdo) // example "2015-09-09 10:53:34 -0500"
	log.Debug("stdout: " + stdo_str)
	log.Debug("stderr: " + string(stde))
	if err != nil {
		log.Errorf("error: %v\n", err)
		return

	}

	var validdate = regexp.MustCompile(`^\"[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)
	if !validdate.MatchString(stdo_str) {
		log.Fatal("Could not parse date: " + stdo_str)
		return
	}

	date_str := stdo_str[1:5] + stdo_str[6:8] + stdo_str[9:11] + "." + stdo_str[12:14] + stdo_str[15:17] + stdo_str[18:20]
	log.Debug("date_str: " + date_str)
	image_tag = date_str

	docker_build_args := []string{"build", "--force-rm", "--no-cache", "--rm", "-t", image_repo + ":" + image_tag}

	dockerfile_array := strings.Split(tag.Dockerfile, "/")

	dockerfile_filename := dockerfile_array[len(dockerfile_array)-1]
	if dockerfile_filename == "" {
		log.Errorf("Dockerfile not defined !?")
		return
	} else if dockerfile_filename != "Dockerfile" {
		docker_build_args = append(docker_build_args, "-f", dockerfile_filename)
	}

	dockerfile_path := ""
	for i := 0; i < len(dockerfile_array)-1; i++ {
		dockerfile_path += "/" + dockerfile_array[i]
	}

	dockerfile_directory := path.Join(git_repo_dir, dockerfile_path) + "/"
	docker_build_args = append(docker_build_args, dockerfile_directory)

	log.Infof("Create image %s ...", image_repo+":"+image_tag)
	stdo, stde, err = RunCommand(exec.Command("docker", docker_build_args...))
	log.Debug("stdout: " + string(stdo))
	log.Debug("stderr: " + string(stde))
	if err != nil {
		log.Errorf("error: %v\n", err)
		return

	}
	log.Infof("Image created: %s ", image_repo+":"+image_tag)
	return
}

func main() {
	log.SetLevel(log.DebugLevel)
	//log.SetLevel(log.InfoLevel)

	if len(os.Args) <= 1 {
		fmt.Println("\nUsage: dockbuild <yaml-file> <reponame> <tag>\n\n")
		os.Exit(0)
	}

	var document Document
	if len(os.Args) > 1 {
		document = read_yaml(os.Args[1])
	}

	if len(os.Args) > 3 {

		err := dockbuild(document, os.Args[2], os.Args[3])
		if err != nil {
			log.Errorf("error: %v\n", err)
			os.Exit(1)

		}
	}

}

// gofmt -w . && go build . && ./dockbuild ~/git/MG-RAST-infrastructure/mgrast.yaml mgrast/v3-web develop
// curl  -X GET "https://api.github.com/repos/wgerlach/Skycore/git/refs/heads/master"
// show head commit : git rev-parse HEAD
// commit date: git log -1  --pretty=format:"%cd" --date=iso # returns 2015-09-18 23:22:36 -0500
