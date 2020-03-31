FROM golang:latest
LABEL name="Forum"
LABEL description="alem school project"
LABEL authors="alseiitov; satsuls; bortico; isabekovgg; altynayk"
LABEL release-date="01.04.2020"
RUN mkdir /app
ADD . /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go build -o main .
CMD ["/app/main"]