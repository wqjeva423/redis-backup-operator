FROM alpine:latest
MAINTAINER d3fk

RUN apk add --no-cache python3 py-pip py-setuptools git ca-certificates redis

RUN pip install python-magic redis redis-py-cluster\
  && git clone https://github.com/s3tools/s3cmd.git /tmp/s3cmd \
  && cd /tmp/s3cmd \
  && python3 /tmp/s3cmd/setup.py install \
  && cd / \
  && rm -rf /tmp/s3cmd \
  && apk del py-pip git

WORKDIR /s3

COPY redis_recover.py /s3/redis_recover.py
COPY redis_backup.py /s3/redis_backup.py
COPY redis_cluster_backup.py /s3/redis_cluster_backup.py
COPY redis_cluster_recover.py /s3/redis_cluster_recover.py

ENTRYPOINT ["s3cmd"]
CMD ["--help"]
