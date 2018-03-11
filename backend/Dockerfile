FROM golang

VOLUME /var/log/backend

COPY src /go/src

RUN go install go_project

CMD /go/bin/go_project

EXPOSE 8000
