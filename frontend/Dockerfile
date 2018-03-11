FROM ubuntu:xenial

RUN apt-get update && apt-get install -y \
    nginx \
    && rm -rf /var/lib/apt/lists/*

COPY site.conf /etc/nginx/sites-available
RUN ln -s /etc/nginx/sites-available/site.conf /etc/nginx/sites-enabled
COPY .htpasswd /etc/nginx

COPY startup.sh /home/
RUN chmod 777 /home/startup.sh
CMD ["bash","/home/startup.sh"]

EXPOSE 9000

COPY vue_project/dist /home/html/