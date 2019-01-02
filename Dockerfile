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

RUN rm /etc/postgresql/11/main/pg_hba.conf \
    && echo "local all all trust" > /etc/postgresql/11/main/pg_hba.conf \
    && echo "host all all 0.0.0.0/0 trust" >> /etc/postgresql/11/main/pg_hba.conf \
    && service postgresql start \
    && psql -U postgres -c 'CREATE DATABASE forum;' \
    && service postgresql stop

WORKDIR /tmp
COPY --from=base /tmp/tp-project-db/service ./service

ENTRYPOINT service postgresql start && ./service
EXPOSE 5000
