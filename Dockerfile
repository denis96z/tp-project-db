FROM golang:1.11.4 as base
WORKDIR /tmp/tp-project-db
COPY . .
RUN go build -mod=vendor -o service .

FROM ubuntu:18.04

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y curl ca-certificates gnupg

RUN curl https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
    && echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main" > /etc/apt/sources.list.d/pgdg.list

RUN apt-get update \
    && apt-get install -y postgresql-11

COPY ./dbconfig /tmp/config
RUN mv -f /tmp/config/pg_hba.conf /etc/postgresql/11/main/pg_hba.conf

RUN service postgresql start \
    && psql -U postgres -c 'CREATE DATABASE forum;' \
    && service postgresql stop

WORKDIR /tmp
COPY --from=base /tmp/tp-project-db/service ./service

ENTRYPOINT service postgresql start && ./service
EXPOSE 5000
