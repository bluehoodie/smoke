FROM python:slim

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get -y install gcc mono-mcs && \
    pip install --no-cache-dir httpbin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get purge -y --auto-remove gcc mono-mcs

EXPOSE 8000

ENTRYPOINT ["gunicorn", "-b", ":8000", "httpbin:app"]