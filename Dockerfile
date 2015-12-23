FROM golang

# Build app
RUN mkdir -p /gopath/app/bk
ENV GOPATH /gopath/app
WORKDIR /gopath/app/bk

RUN git clone --depth 1 git://github.com/shaalx/bookmark.git .
RUN go get github.com/shaalx/bookmark
RUN go build -o bookmark

EXPOSE 80

CMD ["/gopath/app/bk/bookmark"]