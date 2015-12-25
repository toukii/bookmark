FROM busybox:latest

# Build app
RUN mkdir -p /usr/static/app
ADD . /usr/static/app/

EXPOSE 80

CMD ["/usr/static/app/bookmark"]