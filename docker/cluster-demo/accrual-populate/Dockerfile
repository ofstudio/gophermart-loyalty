FROM alpine
RUN apk add --no-cache curl
COPY ./populate.sh /
RUN chmod +x /populate.sh
CMD ["./populate.sh"]
