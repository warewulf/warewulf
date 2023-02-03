# Debian, Buster

## Docker container

### build and push

> PSI specific, change for upstream

change into your docker build directory
```bash
docker build -t docker.psi.ch:5000/debian:buster-slim .
docker push docker.psi.ch:5000/debian:buster-slim
```

### import docker container into warewulf

```bash
wwctl container import docker://docker.psi.ch:5000/debian:buster-slim debian-10:slim
```


