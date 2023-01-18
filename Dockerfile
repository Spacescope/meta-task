FROM golang:1.19.5-bullseye as builder

COPY . /opt
RUN cd /opt && go build -o bin/observatory-task cmd/observatorytask/main.go

FROM debian:bullseye
RUN apt update && apt-get install ca-certificates -y
RUN adduser --gecos "Devops Starboard,Github,WorkPhone,HomePhone" --home /app/api-server --disabled-password spacescope
USER spacescope
COPY --from=builder /opt/bin/observatory-task /app/observatory-task/observatory-task

CMD ["--conf", "/app/observatory-task/service.conf"]
ENTRYPOINT ["/app/observatory-task/observatory-task"]
