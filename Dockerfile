FROM golang

# Build app
RUN mkdir -p /usr/static/app/bk
ENV GOPATH /usr/static/app
WORKDIR /usr/static/app/bk

RUN git clone --depth 1 git://github.com/shaalx/bookmark.git . && go get github.com/shaalx/bookmark && go build -o bookmark

EXPOSE 80

CMD ["/usr/static/app/bk/bookmark"]