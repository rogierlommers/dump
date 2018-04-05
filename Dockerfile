FROM ubuntu
LABEL description="Greedy, a personal readinglist"
LABEL maintainer="Rogier Lommers <rogier@lommers.org>"

# install dependencies
RUN apt-get update  
RUN apt-get install -y ca-certificates

# add binary
COPY bin/dumper-linux-amd64 /dumper-linux-amd64

# change to data dir and run bianry
WORKDIR "/"
CMD ["/dumper-linux-amd64"]
