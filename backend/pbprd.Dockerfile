
# light golang 1.19
FROM golang:1.22


WORKDIR /prod
# copio solo build e static file front end
COPY ./bedocker .
COPY ./pb_public ./pb_public
COPY ./views_template ./views_template

# perche i dati li prendo da un altro volume

# expose port
EXPOSE 8090

# run server
CMD ["./bedocker", "serve", "--http", "0.0.0.0:8090"]

