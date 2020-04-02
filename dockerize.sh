echo "### Building image"
docker build -t forum .
echo
echo "### Running image"
docker run -d -p 8080:8080 --name web forum
echo
echo "### Images list"
docker images
echo
echo "### Containers list"
docker container ls
PAUSE