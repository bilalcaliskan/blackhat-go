# Abusing Databases And Filesystems

Now that we’ve covered the majority of common network protocols used for active service interrogation, command and control, 
and other malicious activity, let’s switch our focus to an equally important topic: data pillaging.

Although data pillaging may not be as exciting as initial exploitation, lateral network movement, or privilege escalation, 
it’s a critical aspect of the overall attack chain. After all, we often need data in order to perform those other activities. 
Commonly, the data is of tangible worth to an attacker. Although hacking an organization is thrilling, the data itself is 
often a lucrative prize for the attacker and a damning loss for the organization.

Depending on which study you read, a breach in 2020 can cost an organization approximately $4 to $7 million. An IBM study 
estimates it costs an organization $129 to $355 per record stolen. Hell, a black hat hacker can make some serious coin off 
the underground market by selling credit cards at a rate of $7 to $80 per card (http://online.wsj.com/public/resources/documents/secureworks_hacker_annualreport.pdf).

The Target breach alone resulted in a compromise of 40 million cards. In some cases, the Target cards were sold for as 
much as $135 per card (http://www.businessinsider.com/heres-what-happened-to-your-target-data-that-was-hacked-2014-10/). 
That’s pretty lucrative. We, in no way, advocate that type of activity, but folks with a questionable moral compass stand 
to make a lot of money from data pillaging.

Enough about the industry and fancy references to online articles—let’s pillage! In this chapter, you’ll learn to set up 
and seed a variety of SQL and NoSQL databases and learn to connect and interact with those databases via Go. We’ll also 
demonstrate how to create a database and filesystem data miner that searches for key indicators of juicy information.

### Setting Up Databases With Docker
In this section, you’ll install various database systems and then seed them with the data you’ll use in this chapter’s 
pillaging examples. We will use Docker containers for that.  The container is compartmentalized from the operating system 
in order to prevent the pollution of the host platform. This is nifty stuff.

And for this chapter, you will use a variety of prebuilt Docker images for the databases you’ll be working with.

#### _Installing and Seeding MongoDB_
`MongoDB` is the only NoSQL database that you’ll use in this chapter. Unlike traditional relational databases, MongoDB 
doesn’t communicate via SQL. Instead, MongoDB uses an easy-to-understand JSON syntax for retrieving and manipulating data. 
Entire books have been dedicated to explaining MongoDB, and a full explanation is certainly beyond the scope of this 
book. For now, you’ll install the Docker image and seed it with fake data.

Unlike traditional SQL databases, MongoDB is `schema-less`, which means that it doesn’t follow a predefined, rigid rule 
system for organizing table data. First, install the MongoDB Docker image with the following command:
```shell script
$ docker run -d --name some-mongo -p 27017:27017 mongo
```

This command downloads the image named mongo from the Docker repository, spins up a new instance named some-mongo—the 
name you give the instance is arbitrary—and maps local port 27017 to the container port 27017. The port mapping is key, 
as it allows us to access the database instance directly from our operating system. Without it, it would be inaccessible.

Check that the container started automatically by listing all the running containers:
```shell script
$ docker ps
```

In the event your container doesn’t start automatically, run the following command:
```
$ docker start some-mongo
```

The start command should get the container going.

Once your container starts, connect to the MongoDB instance by using the run command—passing it the MongoDB client; 
that way, you can interact with the database to seed data:
```shell script
$ docker run -it --link some-mongo:mongo --rm mongo sh \
  -c 'exec mongo "$MONGO_PORT_27017_TCP_ADDR:$MONGO_PORT_27017_TCP_PORT/store"'
>
```

This magical command runs a disposable, second Docker container that has the MongoDB client binary installed—so you 
don’t have to install the binary on your host operating system—and uses it to connect to the some-mongo Docker container’s 
MongoDB instance. In this example, you’re connecting to a database named test.

After you drop into mongo shell, put the below into mongo shell to insert something. you insert an array of documents 
into the transactions collection.
```shell script
> db.transactions.insert([
{
    "ccnum" : "4444333322221111",
    "date" : "2019-01-05",
    "amount" : 100.12,
    "cvv" : "1234",
    "exp" : "09/2020"
},
{
    "ccnum" : "4444123456789012",
    "date" : "2019-01-07",
    "amount" : 2400.18,
    "cvv" : "5544",
    "exp" : "02/2021"
},
{
    "ccnum" : "4465122334455667",
    "date" : "2019-01-29",
    "amount" : 1450.87,
    "cvv" : "9876",
    "exp" : "06/2020"
}
]);
```

That’s it! You’ve now created your MongoDB database instance and seeded it with a transactions collection that contains 
three fake documents for querying. You’ll get to the querying part in a bit, but first, you should know how to install 
and seed traditional SQL databases.

#### _Installing and Seeding PostgreSQL and MySQL Databases_
PostgreSQL (also called Postgres) and MySQL are probably the two most common, well-known, enterprise-quality, open source 
relational database management systems, and official Docker images exist for both. Because of their similarity and the 
general overlap in their installation steps, we batched together installation instructions for both here.

First, much in the same way as for the MongoDB example in the previous section, download and run the appropriate Docker 
image:
```shell script
$ docker run -d --name some-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -d mysql
$ docker run -d --name some-postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
```

After your containers are built, confirm they are running, and if they aren’t, you can start them via the `docker start name` 
command.

Next, you can connect to the containers from the appropriate client—again, using the Docker image to prevent installing 
any additional files on the host—and proceed to create and seed the database. As below, you can see the MySQL logic:
```shell script
$ docker run -it --link some-mysql:mysql --rm mysql sh -c \
'exec mysql -h "$MYSQL_PORT_3306_TCP_ADDR" -P"$MYSQL_PORT_3306_TCP_PORT" \
-uroot -p"$MYSQL_ENV_MYSQL_ROOT_PASSWORD"'
mysql> create database store;
mysql> use store;
mysql> create table transactions(ccnum varchar(32), date date, amount float(7,2),
    -> cvv char(4), exp date);
```

In above, like the one that follows, starts a disposable Docker shell that executes the appropriate database client binary. 
It creates and connects to the database named store and then creates a table named transactions. The two listings are 
identical, with the exception that they are tailored to different database systems.

In below, you can see the Postgres logic, which differs slightly in syntax from MySQL.
```shell script
$ docker run -it --rm --link some-postgres:postgres postgres psql -h postgres -U postgres
postgres=# create database store;
postgres=# \connect store
store=# create table transactions(ccnum varchar(32), date date, amount money, cvv
        char(4), exp date);
```

In both MySQL and Postgres, the syntax is identical for inserting your transactions. For example, in below, you can see 
how to insert three documents into a MySQL transactions collection.
```shell script
mysql> insert into transactions(ccnum, date, amount, cvv, exp) values
    -> ('4444333322221111', '2019-01-05', 100.12, '1234', '2020-09-01');
mysql> insert into transactions(ccnum, date, amount, cvv, exp) values
    -> ('4444123456789012', '2019-01-07', 2400.18, '5544', '2021-02-01');
mysql> insert into transactions(ccnum, date, amount, cvv, exp) values
    -> ('4465122334455667', '2019-01-29', 1450.87, '9876', '2019-06-01');
```

Try inserting the same three documents into your Postgres database. Same commands work.

#### _Installing and Seeding Microsoft SQL Server Databases_
In 2016, Microsoft began making major moves to open-source some of its core technologies. One of those technologies was 
Microsoft SQL (MSSQL) Server. It feels pertinent to highlight this information while demonstrating what, for so long, 
wasn’t possible—that is, installing MSSQL Server on a Linux operating system. Better yet, there’s a Docker image for it, 
which you can install with the following command:
```shell script
$ docker run -d --name some-mssql -p 1433:1433 -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=Password1!' -d microsoft/mssql-server-linux
```

That command is similar to the others you ran in the previous two sections, but per the documentation, the `SA_PASSWORD` value 
needs to be complex—a combination of uppercase letters, lowercase letters, numbers, and special characters—or you won’t be 
able to authenticate to it. Since this is just a test instance, the preceding value is trivial but minimally meets those 
requirements—just as we see on enterprise networks all the time!

With the image installed, start the container, create the schema, and seed the database, as below:
```shell script
$ docker exec -it some-mssql /opt/mssql-tools/bin/sqlcmd -S localhost \
-U sa -P 'Password1!'
> create database store;
> go
> use store;
> create table transactions(ccnum varchar(32), date date, amount decimal(7,2),
> cvv char(4), exp date);
> go
> insert into transactions(ccnum, date, amount, cvv, exp) values
> ('4444333322221111', '2019-01-05', 100.12, '1234', '2020-09-01');
> insert into transactions(ccnum, date, amount, cvv, exp) values
> ('4444123456789012', '2019-01-07', 2400.18, '5544', '2021-02-01');
> insert into transactions(ccnum, date, amount, cvv, exp) values
> ('4465122334455667', '2019-01-29', 1450.87, '9876', '2020-06-01');
> go
```

The previous listing replicates the logic we demonstrated for MySQL and Postgres earlier. It uses Docker to connect to 
the service, creates and connects to the store database, and creates and seeds a transactions table. We’re presenting 
it separately from the other SQL databases because it has some MSSQL-specific syntax.

### Connecting And Querying Databases In Go
Now that you have a variety of test databases to work with, you can build the logic to connect to and query those 
databases from a Go client. We’ve divided this discussion into two topics—one for MongoDB and one for traditional SQL 
databases.

#### _Querying MongoDB_
Despite having an excellent standard SQL package, Go doesn’t maintain a similar package for interacting with NoSQL 
databases. Instead you’ll need to rely on third-party packages to facilitate this interaction. Rather than inspect the 
implementation of each third-party package, we’ll focus purely on MongoDB. We’ll use the `mgo` (pronounce mango) DB 
driver for this.

We will keep our coding exercises in [db](db) subfolder. For this chapter, please refer to [db/mongo-connect/main.go](db/mongo-connect/main.go)

Start by installing the mgo driver with the following command:
```shell script
$ go get gopkg.in/mgo.v2
```

You can now establish connectivity and query your store collection (the equivalent of a table), which requires even less 
code than the SQL sample code we’ll create later. See [db/mongo-connect/main.go](db/mongo-connect/main.go):
```go
package main

import (
    "fmt"
    "log"

    mgo "gopkg.in/mgo.v2"
)

type Transaction struct { ❶
    CCNum      string  `bson:"ccnum"`
    Date       string  `bson:"date"`
    Amount     float32 `bson:"amount"`
    Cvv        string  `bson:"cvv"`
    Expiration string  `bson:"exp"`
}

func main() {
    session, err := mgo.Dial("127.0.0.1") ❷
    if err != nil {
        log.Panicln(err)
    }  
    defer session.Close()

    results := make([]Transaction, 0)
    if err := session.DB("store").C("transactions").Find(nil).All(&results)❸; err != nil {
        log.Panicln(err)
    }  
    for _, txn := range results { ❹
        fmt.Println(txn.CCNum, txn.Date, txn.Amount, txn.Cvv, txn.Expiration)
    }
}
```

First, you define a type, `Transaction`, which will represent a single document from your `store` collection ❶. The 
internal mechanism for data representation in MongoDB is `binary JSON`. For this reason, use tagging to define any 
marshaling directives. In this case, you’re using tagging to explicitly define the element names to be used in the 
binary JSON data.

In your `main()` function ❷, call `mgo.Dial()` to create a session by establishing a connection to your database, 
testing to make sure no errors occurred, and deferring a call to close the session. You then use the `session` variable 
to query the store database ❸, retrieving all the records from the `transactions` collection. You store the results in 
a `Transaction slice, named results`. Under the covers, your structure tags are used to unmarshal the binary JSON to 
your defined type. Finally, loop over your result set and print them to the screen ❹. In both this case and the SQL 
sample in the next section, your output should look similar to the following:
```shell script
$ go run main.go
4444333322221111 2019-01-05 100.12 1234 09/2020
4444123456789012 2019-01-07 2400.18 5544 02/2021
4465122334455667 2019-01-29 1450.87 9876 06/2020
```