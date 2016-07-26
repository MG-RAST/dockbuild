#!/usr/bin/env python3
import subprocess
import os
import json
import time
import argparse
import sys

tmp_dir='/tmp/dockerbuilds/'



build_config_json= """{
    "mgrast/api-server": {
        "latest": {
            "git_branch": "api",
            "git_path": "dockerfiles/api/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST"
        }
    },
    "mgrast/awe": {
        "latest": {
            "git_branch": "master",
            "git_path": "Dockerfile",
            "git_repository": "https://github.com/MG-RAST/AWE"
        },
        "deprecated": {
            "git_branch": "master",
            "git_path": "Dockerfile_deprecated",
            "git_repository": "https://github.com/MG-RAST/AWE"
        }
    },
    "mgrast/cassandra": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/cassandra/docker/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mgrast/log-courier": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/log-courier/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mgrast/logstash": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/logstash/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mgrast/memcached": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/memcached/docker/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mg-rast/nginxconfd": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/nginx/docker/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mgrast/v3-web": {
        "latest": {
            "git_branch": "master",
            "git_path": "dockerfiles/web/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST"
            }
    },
    "mgrast/v4-web": {
        "latest": {
            "git_branch": "master",
            "git_path": "Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RASTv4"
            }
    },
    "mgrast/pipeline": {
        "latest": {
            "git_branch": "master",
            "git_path": "dockerfiles/mgrast_base/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/pipeline"
            }
    },
    "mgrast/solr-m5nr": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/solr-m5nr/docker/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
    },
    "mgrast/solr-metagenome": {
        "latest": {
            "git_branch": "master",
            "git_path": "services/solr-metagenome/docker/Dockerfile",
            "git_repository": "https://github.com/MG-RAST/MG-RAST-infrastructure"
            }
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

def build_image(build_config, image, simulate=False, checkout=None):
    try:
        
        if not image:
           raise MyException("No image specified")
        
        image_split = image.split(':')
        
        image_name = image_split[0]
        image_tag=None
        if len(image_split) > 1:
            image_tag = image_split[1]
        
        if not image_tag:
            image_tag = 'latest'
        
        if not image_name in build_config:
            raise MyException("Image not found")
        
        image_obj= build_config[image_name]
        if not image_tag in image_obj:
            raise MyException("tag %s for image % not found" % (image_tag, image_name))
        
        tag_object = image_obj[image_tag]
        
        git_repository =  tag_object['git_repository']
        if not git_repository:
            raise MyException("field git_repository not found")
        
        directory = git_repository.split('/')[-1]
        git_path = tag_object['git_path']
        git_branch = tag_object['git_branch']
    
        
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
        fresh_clone = True
        if repository_dir_abs_exists:
            # TODO check that local git is mnot broken !
            chdir(dockerfile_dir_abs)
            run("git pull", shell=True) 
            run("git checkout %s" % (git_branch), shell=True)
        else:
            fresh_clone = False
            chdir(tmp_dir, simulate=simulate)
            cmd_clone ="git clone --recursive -b %s %s" % (git_branch, git_repository)
            run(cmd_clone, shell=True, simulate=simulate)
            run(cmd_cd, shell=True, simulate=simulate)
            chdir(dockerfile_dir_abs, simulate=simulate)
            
        # git reset --hard
        if checkout:
            checkout_cmd = "git checkout %s" % (checkout)
            run(checkout_cmd, shell=True, simulate=simulate)
        else:
            if not fresh_clone:
                reset_cmd = "git reset --hard"
                run(reset_cmd, shell=True, simulate=simulate)
            
    except Exception as e:
        raise MyException("Something went wrong in the git clone/pull phase: %s" % (str(e)))
    
    try:
        dockerfile_name = git_path.split('/')[-1]
    
        # convert <date> to actual date string
        #date_str = time.strftime('%Y%m%d.%H%M')
        #for i, value in enumerate(tags):
        #    if value == "<date>":
        #        tags[i] = date_str
    
    
        #first_tag = tags[0]
    
        # TODO check if container is running!?
    
        cmd_rmi = "docker rmi --force=true %s:%s" % ( image_name , image_tag)
        try:
            run(cmd_rmi, shell=True, simulate=simulate)
        except:
            pass
    
        cmd_build = "docker build -t %s:%s -f %s ." % ( image_name , image_tag , dockerfile_name)
        run(cmd_build, shell=True, simulate=simulate)
    except Exception as e:
        raise MyException("Something went wrong in the build phase: %s" % (str(e)))


###################################


parser = argparse.ArgumentParser()
parser.add_argument("-c", "--config", help='show config', action='store_true')
parser.add_argument("--checkout", help='checkout <branch> or <commit>', action='store')
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
for image_name in sorted(build_config_dict):
    print("  "+image_name)

if not os.path.exists(tmp_dir):
    os.makedirs(tmp_dir)

#build_image(build_config_dict, 'mgrast/awe', simulate=True)
if args.args:
    try:
        build_image(build_config_dict, args.args[0], simulate=args.simulate)
    except Exception as e:
        print("error building image: %s" % (str(e)))
    
    








