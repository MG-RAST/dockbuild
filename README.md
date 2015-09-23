# dockbuild

dockbuild is a wrapper around "git" and "docker" to conveniently build Docker images from Dockerfiles located in git repositories. A yaml file is used to specify git repository, branch and Dockerfile location of each image.

## Usage:

```bash
dockbuild <yaml> <imagename> <tag>
```

## Get or build Docker image

```bash
docker pull mgrast/dockbuild
```

or

```bash
git clone https://github.com/MG-RAST/dockbuild.git
cd dockbuild
docker rmi dockbuild
docker build -t dockbuild .
```


## Extract binary from Docker image:
```bash
rm -f ./dockbuild
docker rm dockbuild
docker create --name dockbuild dockbuild
docker cp dockbuild:/app/dockbuild .
./dockbuild
```

## Example yaml file
See the MG-RAST [yaml file](https://github.com/MG-RAST/MG-RAST-infrastructure/blob/master/mgrast.yaml) for an example.
