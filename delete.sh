docker container rm -f web
docker image rm -f forum
docker system prune -a
docker container ls -a
docker image ls -a
PAUSE