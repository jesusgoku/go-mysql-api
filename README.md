# Contacts API (with MySQL)

## Build with Docker

```sh
cp .env.dist .env
docker-compose up -d
```

## Build local

```sh
cp .env.dist .env
go build -o bin/contacts-api src/main.go
```

## Usage API

### Get Contacts

```sh
curl http://localhost:5000/api/v1/contacts
```

### Get Contact

```sh
curl http://localhost:5000/api/v1/contacts/1
```

### Create Contact

```sh
curl -X POST -H 'Content-Type' -d '{ "name": "Some Name", "address": "Some Address", "email": "some@email.com" }' http://localhost:5000/api/v1/contacts
```

### Update Contact

```sh
curl -X PUT -H 'Content-Type' -d '{ "name": "Some Name", "address": "Some Address", "email": "some@email.com" }' http://localhost:5000/api/v1/contacts/1
```

### Delete Contact

```sh
curl -X DELETE http://localhost:5000/api/v1/contacts/1
```
