
# this image is used for development only
# it has live reload
# data are erased and taken from the docker-compose specification

FROM golang:1.22

WORKDIR /home

COPY . .
RUN rm -rf /home/pb_data/*

# live reload
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

RUN go mod download
EXPOSE 8090

# -- debug
CMD CompileDaemon -build="go build be.go" -command="./be serve --debug" 
