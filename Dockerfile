FROM golang

# Build app
RUN mkdir -p /static/app/bk
ENV GOPATH /static/app
WORKDIR /static/app/bk

RUN git clone --depth 1 git://github.com/shaalx/bookmark.git .
RUN go get github.com/shaalx/bookmark
RUN go build -o bookmark

EXPOSE 80

CMD ["/static/app/bk/bookmark"]