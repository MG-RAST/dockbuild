# dockbuild

dockbuild is a wrapper around "git" and "docker" to conveniently build Docker images from Dockerfiles located in git repositories. A json file is used to specify git repository, branch and Dockerfile location of each image.


# Installation (with dockbuild using host-installed python)

```bash
mkir -p ~/git
cd ~/git
git clone https://github.com/MG-RAST/dockbuild.git
```

# Installation (with dockbuild in Docker container)
```bash
git pull mgrast/dockbuild
```

# Dockerimage build instructions
```bash
git clone https://github.com/MG-RAST/dockbuild.git
cd dockbuild
docker build -t mgrast/dockbuild .
```


## Usage (dockbuild on host):

```bash
dockbuild.py [--simulate] <imagename>:<tag>
```
## Usage (dockbuild in container):
This uses a wrapper script (CoreOS only at the moment)
```bash
dockbuild.sh [--simulate] <imagename>:<tag>
```
