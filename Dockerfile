FROM golang:1.24.7-trixie AS builder

COPY . /opt
RUN cd /opt && go build -o bin/meta-task cmd/meta-task/main.go

FROM debian:trixie
RUN apt update && apt-get install ca-certificates -y
RUN useradd -m -d /app/meta-task -c "Devops Starboard,Github,WorkPhone,HomePhone" -s /usr/sbin/nologin spacescope
USER spacescope
COPY --from=builder /opt/bin/meta-task /app/meta-task/meta-task

CMD ["--conf", "/app/meta-task/service.conf"]
ENTRYPOINT ["/app/meta-task/meta-task"]
