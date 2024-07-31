FROM golang:1.22.5-bullseye AS builder

COPY . /opt
RUN cd /opt && go build -o bin/meta-task cmd/meta-task/main.go

FROM debian:bullseye
RUN apt update && apt-get install ca-certificates -y
RUN adduser --gecos "Devops Starboard,Github,WorkPhone,HomePhone" --home /app/meta-task --disabled-password spacescope
USER spacescope
COPY --from=builder /opt/bin/meta-task /app/meta-task/meta-task

CMD ["--conf", "/app/meta-task/service.conf"]
ENTRYPOINT ["/app/meta-task/meta-task"]
