
# light golang 1.19
FROM golang:1.20


WORKDIR /prod
# copio solo build e static file front end
COPY ./bedocker .
COPY ./pb_public ./pb_public
# perche i dati li prendo da un altro volume

# expose port
EXPOSE 8090

# run server
CMD ["./bedocker", "serve", "--http", "0.0.0.0:8090"]

