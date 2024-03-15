FROM debian:latest
RUN apt update -y
RUN apt upgrade -y
RUN apt install golang-go -y
COPY . /root/
WORKDIR /root/cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o backdoorBoiLinux
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o backdoorBoiWindows.exe
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o backdoorBoiDarwin

# Tests
# RUN ./backdoorBoiLinux -h
RUN ./backdoorBoiLinux Netcat -h
RUN ./backdoorBoiLinux FileTransfer -h
