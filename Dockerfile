FROM golang:1.14

RUN set -xe && mkdir /app
WORKDIR /app

COPY ./go.* ./
RUN set -xe && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o allure-portal .

FROM marvell/allure:2

COPY --from=0 /app/allure-portal /bin/allure-portal

VOLUME [ "/storage" ]
EXPOSE 80

ENTRYPOINT [ "allure-portal", "-storage-path=/storage" ]
