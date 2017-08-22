############################################################
# TEMP
############################################################

FROM scratch

MAINTAINER baor

ADD ca-certificates.crt /etc/ssl/certs/
ADD main /

EXPOSE 8080

CMD ["/main"]