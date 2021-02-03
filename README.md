# BookingBackend

Backend for Facility Booking assignment.

[Link to Docker Image](https://hub.docker.com/repository/docker/qingzz/bookingbackend)

This docker image will requires BookingDB to be running.
Ideally setup in a environment where it could reach bookingDB container (E.g same docker network)

Example 


```sh
sudo docker network create bookingNetwork

sudo docker pull qingzz/bookingdb:latest
sudo docker run --name facilityBookingDB -e POSTGRES_PASSWORD=<PostgreSQL_PASSWORD> -p 5432:5432 -d qingzz/bookingdb:latest
sudo docker network connect bookingNetwork facilityBookingDB

sudo docker pull qingzz/bookingbackend:latest
sudo docker run --name facilityBookingBackend -e APP_DB_PASSWORD=<USER_PASSWORD> -e APP_DB_NAME=<DB_HOST> -e APP_DB_USERNAME=<USER_ID> -e APP_DB_PORT=5432 -e APP_DB_HOST=facilityBookingDB -p 80:8000 -d qingzz/bookingbackend:latest
sudo docker network connect bookingNetwork facilityBookingBackend
```

Default value for environment variables
USER_ID = facilityadmin
USER_PASSWORD = faci1ityAdmin
DB_HOST = facility_booking

USER_ID not postgres, it is a created user for the required database
DB_HOST is the IP for Database IP, if connect via docker network use docker name