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

#### _Querying SQL Databases_
Go contains a standard package, called `database/sql`, that defines an interface for interacting with SQL and SQL-like 
databases. The base implementation automatically includes functionality such as connection pooling and transaction 
support. Database drivers adhering to this interface automatically inherit these capabilities and are essentially 
interchangeable, as the API remains consistent between drivers. The function calls and implementation in your code are 
identical whether you’re using Postgres, MSSQL, MySQL, or some other driver. This makes it convenient to switch backend 
databases with minimal code change on the client. Of course, the drivers can implement database-specific capabilities 
and use different SQL syntax, but the function calls are nearly identical.


#### _Querying SQL Databases_
We will keep our coding exercises in [db](db) subfolder. For this chapter, please refer to [db/mysql-connect/main.go](db/mysql-connect/main.go)

You start by installing the driver with the following command:
```
$ go get github.com/go-sql-driver/mysql
```

Then, you can create a basic client that connects to the database and retrieves the information from your transactions 
table—using the script in [db/mysql-connect/main.go](db/mysql-connect/main.go).

```go
package main

import (
    "database/sql" ❶
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql" ❷
)

func main() {
    db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/store")❸
    if err != nil {
        log.Panicln(err)
    }  
    defer db.Close()

    var (
        ccnum, date, cvv, exp string
        amount                float32
    )  
    rows, err := db.Query("SELECT ccnum, date, amount, cvv, exp FROM transactions") ❹
    if err != nil {
        log.Panicln(err)
    }  
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&ccnum, &date, &amount, &cvv, &exp)❺
        if err != nil {
            log.Panicln(err)
        }
        fmt.Println(ccnum, date, amount, cvv, exp)
    }  
    if rows.Err() != nil {
        log.Panicln(err)
    }
}

```
The code begins by importing Go’s `database/sql` package ❶. This allows you to utilize Go’s awesome standard SQL 
library interface to interact with the database. You also import your MySQL database driver ❷. The leading underscore 
indicates that it’s imported anonymously, which means its exported types aren’t included, but the driver registers 
itself with the sql package so that the MySQL driver itself handles the function calls.

Next, you call `sql.Open()` to establish a connection to our database ❸. The first parameter specifies which driver 
should be used—in this case, the driver is mysql—and the second parameter specifies your connection string. You then 
query your database, passing an SQL statement to select all rows from your transactions table ❹, and then loop over the 
rows, subsequently reading the data into your variables and printing the values ❺.

That’s all you need to do to query a MySQL database. Using a different backend database requires only the following 
minor changes to the code:
  - Import the correct database driver.
  - Change the parameters passed to sql.Open().
  - Tweak the SQL syntax to the flavor required by your backend database.
 
Among the several database drivers available, many are pure Go, while a handful of others use cgo for some underlying interaction. Check out the list of available drivers at https://github.com/golang/go/wiki/SQLDrivers/.

### Building A Database Miner
In this section, you will create a tool that inspects the database schema (for example, column names) to determine whether 
the data within is worth pilfering. For instance, say you want to find passwords, hashes, social security numbers, and 
credit card numbers. Rather than building one monolithic utility that mines various backend databases, you’ll create 
separate utilities—one for each database—and implement a defined interface to ensure consistency between the 
implementations. This flexibility may be somewhat overkill for this example, but it gives you the opportunity to 
create reusable and portable code.

The interface should be minimal, consisting of a few basic types and functions, and it should require the implementation 
of a single method to retrieve database schema. [db/dbminer/dbminer.go](db/dbminer/dbminer.go), defines the database 
miner’s interface.

The code begins by defining an interface named `DatabaseMiner` ❶. A single method, called `GetSchema()`, is required 
for any types that implement the interface. Because each backend database may have specific logic to retrieve the 
database schema, the expectation is that each specific utility can implement the logic in a way that’s unique to the 
backend database and driver in use.

Next, you define a `Schema` type, which is composed of a few subtypes also defined here ❷. You’ll use the Schema type 
to logically represent the database schema—that is, databases, tables, and columns. You might have noticed that your 
`GetSchema()` function, within the interface definition, expects implementations to return a `*Schema`.

Now, you define a single function, called `Search()`, which contains the bulk of the logic. The `Search()` function 
expects a `DatabaseMiner` instance to be passed to it during the function call, and stores the miner value in a 
variable named `m` ❸. The function starts by calling `m.GetSchema()` to retrieve the schema ❹. The function then 
loops through the entire schema, searching against a list of regular expression (regex) values for column names that 
match ❺. If it finds a match, the database schema and matching field are printed to the screen.

Lastly, define a function named `getRegex()` ❻. This function compiles regex strings by using Go’s `regexp` package 
and returns a slice of these values. The regex list consists of case-insensitive strings that match against common 
or interesting field names such as ccnum, ssn, and password.

With your database miner’s interface in hand, you can create utility-specific implementations. Let’s start with the 
MongoDB database miner.

#### _Implementing a MongoDB Database Miner_
The MongoDB utility program in [db/mongo/main.go](db/mongo/main.go) implements the interface defined in [db/dbminer/dbminer.go](db/dbminer/dbminer.go) 
while also integrating the database connectivity code you built in [db/mongo-connect/main.go](db/mongo-connect/main.go).

```go
   package main

   import (
       "os"

    ❶ "github.com/blackhatgo/bhg/ch-7/db/dbminer"
       "gopkg.in/mgo.v2"
       "gopkg.in/mgo.v2/bson"
   )

❷ type MongoMiner struct {
       Host    string
       session *mgo.Session
   }

❸ func New(host string) (*MongoMiner, error) {
       m := MongoMiner{Host: host}
       err := m.connect()
       if err != nil {
           return nil, err
       }  
       return &m, nil
   }

❹ func (m *MongoMiner) connect() error {
       s, err := mgo.Dial(m.Host)
       if err != nil {
           return err
       }  
       m.session = s
       return nil
   }

❺ func (m *MongoMiner) GetSchema() (*dbminer.Schema, error) {
       var s = new(dbminer.Schema)

       dbnames, err := m.session.DatabaseNames()❻
       if err != nil {
           return nil, err
       }

       for _, dbname := range dbnames {
           db := dbminer.Database{Name: dbname, Tables: []dbminer.Table{}}
           collections, err := m.session.DB(dbname).CollectionNames()❼
           if err != nil {
               return nil, err
           }
           for _, collection := range collections {
               table := dbminer.Table{Name: collection, Columns: []string{}}

               var docRaw bson.Raw
               err := m.session.DB(dbname).C(collection).Find(nil).One(&docRaw)❽
               if err != nil {
                   return nil, err
               }

               var doc bson.RawD
               if err := docRaw.Unmarshal(&doc); err != nil {❾
                   if err != nil {
                       return nil, err
                   }
               }

               for _, f := range doc {
                   table.Columns = append(table.Columns, f.Name)
               }
               db.Tables = append(db.Tables, table)
           }
           s.Databases = append(s.Databases, db)
       }  
       return s, nil
   }

   func main() {

       mm, err := New(os.Args[1])
       if err != nil {
           panic(err)
       }  
    ❿ if err := dbminer.Search(mm); err != nil {
           panic(err)
       }
   }
```

You start by importing the `dbminer` package that defines your `DatabaseMiner interface` ❶. Then you define a 
`MongoMiner` type that will be used to implement the interface ❷. For convenience, you define a `New()` function that 
creates a new instance of your `MongoMiner` type ❸, calling a method named `connect()` that establishes a connection 
to the database ❹. The entirety of this logic essentially bootstraps your code, connecting to the database in a 
fashion similar to that discussed in [db/mongo-connect/main.go](db/mongo-connect/main.go).

The most interesting portion of the code is your implementation of the `GetSchema()` interface method ❺. Unlike in 
the previous MongoDB sample code in [db/mongo-connect/main.go](db/mongo-connect/main.go), you are now inspecting the 
MongoDB metadata, first retrieving database names ❻ and then looping over those databases to retrieve each database’s 
collection names ❼. Lastly, the function retrieves the raw document that, unlike a typical MongoDB query, uses lazy 
unmarshaling ❽. This allows you to explicitly unmarshal the record into a generic structure so that you can inspect 
field names ❾. If not for lazy unmarshaling, you would have to define an explicit type, likely using bson tag 
attributes, in order to instruct your code how to unmarshal the data into a struct you defined. In this case, you don’t 
know (or care) about the field types or structure—you just want the field names (not the data)—so this is how you can 
unmarshal structured data without needing to know the structure of that data beforehand.

Your main() function expects the IP address of your MongoDB instance as its lone argument, calls your New() function to 
bootstrap everything, and then calls `dbminer.Search()`, passing to it your MongoMiner instance ❿. Recall that 
`dbminer.Search()` calls GetSchema() on the received DatabaseMiner instance; this calls your MongoMiner implementation 
of the function, which results in the creation of dbminer.Schema that is then searched against the regex list in 
[db/dbminer/dbminer.go](db/dbminer/dbminer.go).

When you run your utility, you are blessed with the following output:
```
$ go run main.go 127.0.0.1
[DB] = store
    [TABLE] = transactions
       [COL] = _id
       [COL] = ccnum
       [COL] = date
       [COL] = amount
       [COL] = cvv
       [COL] = exp
[+] HIT: ccnum
```

You found a match! It may not look pretty, but it gets the job done—successfully locating the database collection that has a field named ccnum.

With your MongoDB implementation built, in the next section, you’ll do the same for a MySQL backend database.

#### _Implementing a MySQL Database Miner_
To make your MySQL implementation work, you’ll inspect the `information_schema.columns table`. This table maintains 
metadata about all the databases and their structures, including table and column names. To make the data the simplest 
to consume, use the following SQL query, which removes information about some of the built-in MySQL databases that are 
of no consequence to your pillaging efforts:
```sql
SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME FROM columns
    WHERE TABLE_SCHEMA NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
    ORDER BY TABLE_SCHEMA, TABLE_NAME
```

The query produces results resembling the following:
```sql
+--------------+--------------+-------------+
| TABLE_SCHEMA | TABLE_NAME   | COLUMN_NAME |
+--------------+--------------+-------------+
| store        | transactions | ccnum       |
| store        | transactions | date        |
| store        | transactions | amount      |
| store        | transactions | cvv         |
| store        | transactions | exp         |
--snip--
```

Although using that query to retrieve schema information is pretty straightforward, the complexity in your code comes 
from logically trying to differentiate and categorize each row while defining your `GetSchema()` function. For 
example, consecutive rows of output may or may not belong to the same database or table, so associating the rows to 
the correct `dbminer.Database` and `dbminer.Table` instances becomes a somewhat tricky endeavor.

[db/mysql/main.go](db/mysql/main.go) defines the implementation.
```go
type MySQLMiner struct {
    Host string
    Db   sql.DB
}

func New(host string) (*MySQLMiner, error) {
    m := MySQLMiner{Host: host}
    err := m.connect()
    if err != nil {
        return nil, err
    }
    return &m, nil
}

func (m *MySQLMiner) connect() error {

    db, err := sql.Open(
        "mysql",
     ❶ fmt.Sprintf("root:password@tcp(%s:3306)/information_schema", m.Host))
    if err != nil {
        log.Panicln(err)
    }
    m.Db = *db
    return nil
}

func (m *MySQLMiner) GetSchema() (*dbminer.Schema, error) {
    var s = new(dbminer.Schema)
 ❷ sql := `SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME FROM columns
    WHERE TABLE_SCHEMA NOT IN
    ('mysql', 'information_schema', 'performance_schema', 'sys')
    ORDER BY TABLE_SCHEMA, TABLE_NAME`
    schemarows, err := m.Db.Query(sql)
    if err != nil {
        return nil, err
    }
    defer schemarows.Close()

    var prevschema, prevtable string
    var db dbminer.Database
    var table dbminer.Table
 ❸ for schemarows.Next() {
        var currschema, currtable, currcol string
        if err := schemarows.Scan(&currschema, &currtable, &currcol); err != nil {
            return nil, err
        }

     ❹ if currschema != prevschema {
            if prevschema != "" {
                db.Tables = append(db.Tables, table)
                s.Databases = append(s.Databases, db)
            }
            db = dbminer.Database{Name: currschema, Tables: []dbminer.Table{}}
            prevschema = currschema
            prevtable = ""
        }

     ❺ if currtable != prevtable {
            if prevtable != "" {
                db.Tables = append(db.Tables, table)
            }
            table = dbminer.Table{Name: currtable, Columns: []string{}}
            prevtable = currtable
        }
     ❻ table.Columns = append(table.Columns, currcol)
    }
    db.Tables = append(db.Tables, table)
    s.Databases = append(s.Databases, db)
    if err := schemarows.Err(); err != nil {
        return nil, err
    }

    return s, nil
}

func main() {
    mm, err := New(os.Args[1])
    if err != nil {
        panic(err)
    }
    defer mm.Db.Close()
    if err := dbminer.Search(mm); err != nil {
        panic(err)
    }
}
```

A quick glance at the code and you’ll probably realize that much of it is very, very similar to the MongoDB example in 
the preceding section. As a matter of fact, the main() function is identical.

The bootstrapping functions are also similar—you just change the logic to interact with MySQL rather than MongoDB. 
Notice that this logic connects to your `information_schema database` ❶, so that you can inspect the database schema.

Much of the code’s complexity resides within the `GetSchema()` implementation. Although you are able to retrieve the 
schema information by using a single database query ❷, you then have to loop over the results ❸, inspecting each row 
so you can determine what databases exist, what tables exist in each database, and what columns exist for each 
table. Unlike in your MongoDB implementation, you don’t have the luxury of JSON/BSON with attribute tags to marshal 
and unmarshal data into complex structures; you maintain variables to track the information in your current row and 
compare it with the data from the previous row, in order to determine whether you’ve encountered a new database or 
table. Not the most elegant solution, but it gets the job done.

Next, you check whether the database name for your current row differs from your previous row ❹. If so, you create a 
new miner.Database instance. If it isn’t your first iteration of the loop, add the table and database to your miner.Schema 
instance. You use similar logic to track and add miner.Table instances to your current miner.Database ❺. Lastly, add 
each of the columns to our miner.Table ❻.

Now, run the program against your Docker MySQL instance to confirm that it works properly, as shown here:
```shell script
$ go run main.go 127.0.0.1
[DB] = store
    [TABLE] = transactions
       [COL] = ccnum
       [COL] = date
       [COL] = amount
       [COL] = cvv
       [COL] = exp
[+] HIT: ccnum
```

The output should be almost indiscernible from your MongoDB output. This is because your dbminer.Schema isn’t producing 
any output—the dbminer.Search() function is. This is the power of using interfaces. You can have specific implementations 
of key features, yet still utilize a single, standard function to process your data in a predictable, usable manner.

### Pillaging a Filesystem
In this section, you’ll build a utility that walks a user-supplied filesystem path recursively, matching against a list 
of interesting filenames that you would deem useful as part of a post-exploitation exercise. These files may contain, 
among other things, personally identifiable information, usernames, passwords, system logins, and password database 
files.

The utility looks specifically at filenames rather than file contents, and the script is made much simpler by the fact 
that Go contains standard functionality in its path/filepath package that you can use to easily walk a directory structure. 
You can see the utility in [filesystem/main.go](filesystem/main.go).

```go
   package main

   import (
       "fmt"
       "log"
       "os"
       "path/filepath"
       "regexp"
   )

❶ var regexes = []*regexp.Regexp{
       regexp.MustCompile(`(?i)user`),
       regexp.MustCompile(`(?i)password`),
       regexp.MustCompile(`(?i)kdb`),
       regexp.MustCompile(`(?i)login`),
   }

❷ func walkFn(path string, f os.FileInfo, err error) error {
       for _, r := range regexes {
        ❸ if r.MatchString(path) {
               fmt.Printf("[+] HIT: %s\n", path)
           }  
       }  
       return nil
   }

   func main() {
       root := os.Args[1]
    ❹ if err := filepath.Walk(root, walkFn); err != nil {
           log.Panicln(err)
       }  
   }
```

In contrast to your database-mining implementations, the filesystem pillaging setup and logic might seem a little too 
simple. Similar to the way you created your database implementations, you define a regex list for identifying interesting 
filenames ❶. To keep the code minimal, we limited the list to just a handful of items, but you can expand the list to 
accommodate more practical usage.


Next, you define a function, named `walkFn()`, that accepts a file path and some additional parameters ❷. The function 
loops over your regular expression list and checks for matches ❸, displaying them to stdout. The walkFn() function ❹ 
is used in the main() function, and passed as a parameter to filepath.Walk(). The Walk() function expects two 
parameters—a root path and a function (in this case, walkFn())—and recursively walks the directory structure starting 
at the value supplied as the root path, calling walkFn() for every directory and file it encounters.

With your utility complete, navigate to your desktop and create the following directory structure:
```
$ tree targetpath/
targetpath/
--- anotherpath
-   --- nothing.txt
-   --- users.csv
--- file1.txt
--- yetanotherpath
    --- nada.txt
    --- passwords.xlsx

2 directories, 5 files
```

Running your utility against this same targetpath directory produces the following output, confirming that your code 
works splendidly:
```
$ go run main.go ./somepath
[+] HIT: somepath/anotherpath/users.csv
[+] HIT: somepath/yetanotherpath/passwords.xlsx
```

That’s just about all there is to it. You can improve the sample code through the inclusion of additional or more-specific 
regular expressions. Further, we encourage you to improve the code by applying the regular expression check only to 
filenames, not directories. Another enhancement we encourage you to make is to locate and flag specific files with a 
recent modified or access time. This metadata can lead you to more important content, including files used as part of 
critical business processes.

### SUMMARY
In this chapter, we dove into database interactions and filesystem walking, using both Go’s native packages and 
third-party libraries to inspect database metadata and filenames. For an attacker, these resources often contain 
valuable information, and we created various utilities that allow us to search for this juicy information.

In the next chapter, you’ll take a look at practical packet processing. Specifically, you’ll learn how to sniff 
and manipulate network packets.