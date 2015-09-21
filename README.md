# dockbuild


## Get or build image

```bash
docker pull mgrast/dockbuild
```

or

```bash
git clone <this repo> ;
cd dockbuild
docker build -t dockbuild .
```bash


##Extract binary:
```
docker create --name dockbuild dockbuild
docker cp dockbuild:/app/dockbuild .
docker rm dockbuild
./dockbuild
```
