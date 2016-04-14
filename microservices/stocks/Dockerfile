FROM golang

RUN go get github.com/goji/httpauth && go get github.com/gorilla/mux

COPY src /go/src/github.com/awslabs/ecs-bootcamp/stocks

RUN go install github.com/awslabs/ecs-bootcamp/stocks

EXPOSE 80

ENTRYPOINT ["/go/bin/stocks"]
