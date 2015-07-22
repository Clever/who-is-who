# who-is-who worker

# In order to make an apple pie one must first create the universe etc
FROM ubuntu:12.04
RUN apt-get update
RUN apt-get install -y wget build-essential git golang bzr mercurial

# GO
# Go installed by apt-get is old, use godeb to install a specific (newer) version
RUN go get launchpad.net/godeb
RUN apt-get remove -y golang golang-go golang-doc golang-src
RUN /usr/lib/go/bin/godeb install 1.5.0

ADD . /who-is-who

RUN mkdir -p /usr/lib/go/src
RUN ln -s /who-is-who /usr/lib/go/src/pr-notifier
RUN GOPATH=/usr/lib/go go get -d who-is-who
RUN GOPATH=/usr/lib/go go build -o /who-is-who/who-is-who who-is-who
WORKDIR /who-is-who
CMD ["./who-is-who"]
