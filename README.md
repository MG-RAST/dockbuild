# dockbuild

```bash
git clone <this repo> ; mkdir bin
docker build -t dockbuild dockbuild
docker create --name dockbuild dockbuild
docker cp dockbuild:/app/dockbuild bin/
docker rm dockbuild
bin/dockbuild
```