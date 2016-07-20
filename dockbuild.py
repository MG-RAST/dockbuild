#!/usr/bin/env python3
import subprocess
import os
import json
import time
import argparse


tmp_dir='/tmp/dockerbuilds/'



build_config_json= """{
    "mgrast/api-server": {
        "git_branch": "api",
        "git_path": "dockerfiles/api/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST",
        "tags" : ["<date>"]
    },
    "mgrast/awe": {
        "git_branch": "master",
        "git_path": "Dockerfile",
        "git_repository": "https://github.com/MG-RAST/AWE",
        "tags" : ["<date>"]
    },
    "mgrast/cassandra": {
        "git_branch": "master",
        "git_path": "services/cassandra/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "log-courier": {
        "git_branch": "master",
        "git_path": "services/log-courier/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "logstash": {
        "git_branch": "master",
        "git_path": "services/logstash/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "memcached": {
        "git_branch": "master",
        "git_path": "services/memcached/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "mg-rast-confd / mg-rast-nginx": {
        "git_branch": "master",
        "git_path": "services/nginx/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "mg-rast-v3-web": {
        "git_branch": "master",
        "git_path": "dockerfiles/web/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST"
    },
    "mg-rast-v4-web": {
        "git_branch": "master",
        "git_path": "Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RASTv4"
    },
    "pipeline": {
        "git_branch": "master",
        "git_path": "dockerfiles/mgrast_base/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/pipeline"
    },
    "solr-m5nr": {
        "git_branch": "master",
        "git_path": "services/solr-m5nr/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    },
    "solr-metagenome": {
        "git_branch": "master",
        "git_path": "services/solr-metagenome/docker/Dockerfile",
        "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
    }
}
"""


class MyException(Exception):
    pass


def run(cmd, shell=False, simulate=False):
    print(cmd)
    if simulate:
        return
    result = subprocess.call(cmd, shell=shell)
    if result != 0:
        raise MyException("Error code: %d" % (result))

def chdir(dir, simulate=False):
    print("cd", dir)
    if simulate:
        return
    os.chdir(dir)

def build_image(build_config, service, simulate=False):
    try:
        if not service in build_config:
            raise MyException("Image not found")
        
        service_obj= build_config[service]
        git_repository =  service_obj['git_repository']
        if not git_repository:
            raise MyException("field git_repository not found")
        
        directory = git_repository.split('/')[-1]
        git_path = service_obj['git_path']
        git_branch = service_obj['git_branch']
    
        if not 'tags' in service_obj:
            raise MyException("field tags not found")
    
        tags = service_obj['tags']
    
        if len(tags) == 0:
            raise MyException("field tags empty")
    
        repository_dir_abs = "%s%s" % (tmp_dir, directory)
        print("repository_dir_abs: ", repository_dir_abs)
        dockerfile_dir = '/'.join(git_path.split('/')[:-1])
        if dockerfile_dir:
            dockerfile_dir_abs = "%s/%s/" % (repository_dir_abs, dockerfile_dir)
        else:
            dockerfile_dir_abs = repository_dir_abs
        
        cmd_cd = "cd %s" % (dockerfile_dir_abs)
    except Exception as e:
        raise MyException("Something went wrong in the preparation phase: %s" % (str(e)))
    
    try:    
        repository_dir_abs_exists = os.path.exists(repository_dir_abs)
        if simulate:
            repository_dir_abs_exists = False
        print("simulate:", simulate)
        if repository_dir_abs_exists:
            # TODO check that local git is mnot broken !
            chdir(dockerfile_dir_abs)
            run("git pull", shell=True) 
            run("git checkout %s" % (git_branch), shell=True)
        else:
            chdir(tmp_dir, simulate=simulate)
            cmd_clone ="git clone --recursive -b %s %s" % (git_branch, git_repository)
            run(cmd_clone, shell=True, simulate=simulate)
            run(cmd_cd, shell=True, simulate=simulate)
            chdir(dockerfile_dir_abs, simulate=simulate)
    except Exception as e:
        raise MyException("Something went wrong in the git clone/pull phase: %s" % (str(e)))
    
    try:
        dockerfile_name = git_path.split('/')[-1]
    
        # convert <date> to actual date string
        date_str = time.strftime('%Y%m%d.%H%M')
        for i, value in enumerate(tags):
            if value == "<date>":
                tags[i] = date_str
    
    
        first_tag = tags[0]
    
        # TODO check if container is running!?
    
        cmd_rmi = "docker rmi --force=true %s:%s" % ( service , first_tag)
        try:
            run(cmd_rmi, shell=True, simulate=simulate)
        except:
            pass
    
        cmd_build = "docker build -t %s:%s -f %s ." % ( service , first_tag , dockerfile_name)
        run(cmd_build, shell=True, simulate=simulate)
    except Exception as e:
        raise MyException("Something went wrong in the build phase: %s" % (str(e)))


###################################


parser = argparse.ArgumentParser()
parser.add_argument("-c", "--config", help='show config', action='store_true')
parser.add_argument("-s", "--simulate", help='simulate the build process', action='store_true')
parser.add_argument('args', help='images to build', nargs=argparse.REMAINDER)
args = parser.parse_args()


# load config
build_config_dict = json.loads(build_config_json)
    

# show config
if args.config:
    print(json.dumps(build_config_dict, sort_keys=True, indent=4))
    sys.exit(0)


print("Available images:")
for service in sorted(build_config_dict):
    print("  "+service)

if not os.path.exists(tmp_dir):
    os.makedirs(tmp_dir)

#build_image(build_config_dict, 'mgrast/awe', simulate=True)
if args.args:
    try:
        build_image(build_config_dict, args.args[0], simulate=args.simulate)
    except Exception as e:
        print("error building image: %s" % (str(e)))
    
    








