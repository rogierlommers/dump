FROM ubuntu
LABEL description="Dump, share your public files"
LABEL maintainer="Rogier Lommers <rogier@lommers.org>"

# install dependencies
RUN apt-get update  
RUN apt-get install -y ca-certificates curl

# add binary
COPY bin/dump-linux-amd64 /dump-linux-amd64
COPY /static /static

# change to data dir and run bianry
WORKDIR "/"
CMD ["/dump-linux-amd64", "-debug"]
