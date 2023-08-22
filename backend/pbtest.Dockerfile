
# light golang 1.19
FROM golang:1.20


WORKDIR /pb

COPY ./bedocker .
COPY ./pb_public ./pb_public

# expose port
EXPOSE 8090

# run server
CMD ["./bedocker", "serve", "--http", "0.0.0.0:8090"]