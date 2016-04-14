FROM golang

RUN go get github.com/goji/httpauth && go get github.com/gorilla/mux

COPY src /go/src/github.com/awslabs/ecs-bootcamp/portal

COPY public /var/www/html/

RUN go install github.com/awslabs/ecs-bootcamp/portal

ENV HTML_FILE_DIR /var/www/html

EXPOSE 80

ENTRYPOINT ["/go/bin/portal"]
