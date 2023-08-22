
# quiesta immagine deve solo produrre una build del backend
# e metterla nella cartella del backend (/backend)
# in questa maniera test e deploy possono essere piu veloci
# NOTARE, è usata solo dai test
# perche il deploy deve prendere su solo cose che sono state controllate
#
# per eseguire il build:
# cd ./backend && docker build -t bedocker -f build.Dockerfile .
#
# per eseguire il container:
# docker run -v ./:/home -t dockergobuild
FROM golang:1.20

# creo una certella di lavoro
WORKDIR /builddir
# metto dentro tuttu i file da compilare
COPY . .

# remove database (anche se non findamentale)
RUN rm -rf /builddir/pb_data/*

# scarico le dipendenze (perche docker parte da 0)
RUN go mod download
# build
CMD go build -o bedocker be.go

