# Gedis

## Purpose

Here I have created an in memory database inspired by Redis. This program also contains a tcp client to connect to the database. The database allows for O(1) lookup, adds and deletes on average. Gedis also suppports transactions. 

Please read Reasoning as to why the program was separated into a 2 parts (server, client).


## Installation

To install Gedis, you first most have (Go)[https://go.dev/] installed and in your PATH. In the root directory run the following command

> go build

Now you have a binary of gedis that you can run on your system. 

(for Unix)
For server Database
> ./gedis 

After you have the server running, you can use any tcp client you want to communicate with the server at address *localhost:8080*. I personally recommend **nc/netcat** on mac/linux. If you do not have a these programs, or want to use the client that comes bundled with gedis, use thee flag -cli.

For tcp client
> ./gedis -cli



## Reasoning

When deciding how to build this in memory database, I had a few objectives:

1. High performance
2. Low resource usage
3. Scalable

After establishing my goals, I concluded that I wouldn't use Python to build this database. While Python is a great language and my go to when it comes to scripting, it isn't great at performance. Next I looked at using Nodejs. Nodejs is great for I/O intensive task but it has one glaring weakness, it loves ram. It would seem counterintuitive to build an in memory database, with a tech stack that loves to eat memory.

Finally I landed on Golang. Go offers high performance with low memory usage. On top of that Go scales really well, allowing concurrency natively with channels and goroutine. On the subject of scalability, I decided to seperate the database from the client.

Since one of our focuses is scalability, I did not want our client to be the only one to connect to the database. I would like multiple on going streams to the same database, so I went with a tcp approach. I decided to go with tcp for my first version of Gedis.

## Usage

```bash
./gedis
nc localhost 8080
```

After starting the server and connecting to server with a TCP client. You can use the following commands to interact with the Database server.

### Commands

#### SET [key] [value]
Stores key value pair

#### GET [key]
Responds with value associated with key

#### DELETE [key]
Deletes key, value pair

#### COUNT [value]
Returns amount of times value is located in database

#### END
Exits Database

#### BEGIN 
Begins a transaction

#### ROLLBACK
Discards transactions without committing

#### COMMIT
Commits transactions

## Todos

* HTTP interface to allow for a web facing interface. 
* Due to the time constraint, I was not able to add testing.
* Persistant backup to a json file or a different database
