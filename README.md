# go-wiki

A simple wiki application developed to create additional programming experience on the topics covered in the go.dev tutorials for building web applications and for accessing databases in Golang.

In the spirit of cross learning the application utilises a microservices approach and can be deployed using a container orchestration service such as Docker Compose / Tilt / Kubernetes etc.

### DB Container Instructions

Build the image:
```docker build ./database -f Dockerfile --tag=wiki-db```

Run the container:
```docker run -d -p 3306:3306 --name=wiki-db wiki-db```

Note, the code currently assumes the DB is running on the Docker bridge network and is assigned the ```172.17.0.2``` IP.  If other containers are already running on the host the IP Address is the ```DBCXN()``` function will need to be updated with the correct IP address for the database container.

## To Do

* Add last modification details, if updated > created then dispay who and when
* Publish new page content
* Error handling / 404 responses
* Implement http response codes
* Add media, 3x images?
* Add Search funtional
* Add users and user administration

## Useful Resources

* Go Play Ground: https://go.dev/play/
* Lorem Ipsum generator: https://www.lipsum.com/


