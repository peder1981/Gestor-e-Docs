FROM fluent/fluentd:v1.16-1

# Use root account to use apt
USER root

# Install dependencies
RUN apk add --no-cache --update --virtual .build-deps \
    sudo build-base ruby-dev \
 && sudo gem install fluent-plugin-elasticsearch \
 && sudo gem sources --clear-all \
 && apk del .build-deps \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

COPY ./conf/fluent.conf /fluentd/etc/

USER fluent
