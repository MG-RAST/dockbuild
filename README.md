# dockbuild

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
git clone <this repo>
cd dockbuild
docker rmi dockbuild
docker build -t dockbuild .
```


## Extract binary from Docker image:
```bash
rm -f ./dockbuild
docker create --name dockbuild dockbuild
docker cp dockbuild:/app/dockbuild .
docker rm dockbuild
./dockbuild
```
