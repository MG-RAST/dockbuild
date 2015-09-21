# dockbuild

```bash
git clone <this repo> ;
cd dockbuild
docker build -t dockbuild .
docker create --name dockbuild dockbuild
docker cp dockbuild:/app/dockbuild .
docker rm dockbuild
./dockbuild
```