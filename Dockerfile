FROM debian:latest
RUN apt update -y
RUN apt upgrade -y
RUN apt install golang-go -y
COPY . /root/
WORKDIR /root/cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ScytheLinux
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o Scythe.exe
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o ScytheDarwin

# Tests
# RUN ./ScytheLinux -h
RUN ./ScytheLinux Netcat -h
RUN ./ScytheLinux FileTransfer -h
