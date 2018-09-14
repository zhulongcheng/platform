# Getting Started with Flux

## Installation

### InfluxDB

Install and run InfluxDB:

    cd path/to/influxdb
    git checkout sgc-flux
    go build -o influxd ./cmd/influxd
    ./influxd

Install the influx command download the nightly:

Mac OS X

    https://dl.influxdata.com/flux/nightlies/fluxd_nightly_darwin_amd64.tar.gz
    tar zxvf fluxd_nightly_darwin_amd64.tar.gz
    cd fluxd_nightly_darwin_amd64

Linux

    wget https://dl.influxdata.com/flux/nightlies/fluxd_nightly_linux_amd64.tar.gz
    tar xvfz fluxd_nightly_linux_amd64.tar.gz
    cd fluxd_nightly_linux_amd64


### REPL

The REPL is a Run Evaluate Print Loop for Flux. You can use it to execute Flux queries in an interactive session. From the directory of the nightly build run these commands:

    export INFLUX_HOST=http://localhost:8086
    ./influx repl


## Learning Flux

Flux is a data scripting language. To learn Flux we walk through a handful of simple expressions and what they mean in Flux.

    > // Flux is a scripting language and so basic expressions are supported.
    > // For example here is a simple addition, the result is printed after hitting ENTER.
    > 1 + 1
    2
    >
    > // you can assign expressions to a variable
    > s = "this is a string"
    > i = 1 // an integer
    > f = 2.0 // a floating point number
    >
    > // typing in the name of a variable will print its value
    > s
    this is a string
    > i
    1
    > f
    2
    >
    > // Flux also has objects, each value can be a different type.
    > o = {name:"Jim", age: 42}
    >
    > // Access a propery of an object using dot notation.
    > o.name
    Jim
    > // Flux has lists too, list values must be the same type.
    > l = [1,2,3,i]
    > l
    [1,2,3,1]
    >
    > // Flux also has primitive values for durations and times.
    > d = 3h5m // 3 hours and 5 minutes
    > t = 2018-09-14T09:00:00-07:00
    >
    > // Flux uses functions for most of its heavy lifting
    > // here is a simple function
    > square = (n) => n * n
    > // it is called like this, using the name of the argument as a keyword argument
    > square(n:3)
    9
    > // Flux does not have positional arugments, the argument name must always be used.

Now that we have an idea of the basic types and structures in a Flux script lets use Flux to perform a query. Flux has a function called `from` that reads data from the connected databse server. The `from` function returns data as a sequence of tables. Other functions in Flux process sequences of tables and transform the data within the tables. This way you can chain together a sequence of transforms on the data producing a final result.

To build our first query we need to learn the `from`, `range`, and `filter` functions. The `from` function returns tables for a specified bucket. A bucket is named like `db/rp` for example `telegraf/autogen`. The `range` function only allows data within a specific time range to be passed through it. The `filter` function only allows data that matches a predicate function to be passed through it.

By chaining these functions together we can get a small subset of the data from our `telegraf` database.

    > from(bucket:"telegraf/autogen") |> range(start:-20s) |> filter(fn:(r) => r._measurement == "cpu")
    ...
    > // The |> pipe forward operator means, pass the result of the previous expression into the next.
    > // By using |> we can easily create a chain of operations.
    >
    > // Notice that no data was returned, this is because the REPL lets you build up your query over multiple lines.
    > // When you want to run the query use the `run` function which executes the most recently entered query.
    > run()
    ...
    >
    > // We can use variables when we are building queries
    > // Lets assign the above query to the `cpu` var to make quick an easy access to the cpu data.
    > cpu = from(bucket:"telegraf/autogen") |> range(start:-1m) |> filter(fn:(r) => r._measurement == "cpu")
    > // Now we can reference `cpu` any time we want the cpu data,
    > // for example we can count the number of rows for each table in our cpu data.
    > cpu |> count()
    > run()
    >
    > // Notice that the `from` function places each series in its own table.
    > // Functions like `count` operate on a per table basis.
    > // After applying the `count` operation each table only has one row, the count of the rows in the input tables.
    > // What if we want to see the counts of each cpu in on big table instead of a table for each series?
    > cpu |> count() |> group(none: true)
    > run()
    >
    > // Grouping the data changes which rows belong in which tables,
    > // the `from` function groups by all columns except for the _time and _value columns.
    > // that means that all data for a single series is in its own table.
    > // Using `group(none:true)` we tell Flux to not group by any columns, so all rows belong in the same table.
    > // We could instead group by the _field column to get our counts organized by field instead.
    > cpu |> count() |> group(by: ["_field"])
    > run()
    > // Or we could group by the _field and cpu columns to get our counts organized by field and cpus.
    > cpu |> count() |> group(by: ["_field","cpu"])
    > run()
    >
    > // Using groups is a powerful way to organize your data so you can perform operations along the dimensions that are important to you.
    > // How would you get the cpu that had the most user cpu over the last minute?
    >
    > // Since we only care about the user cpu lets add a new filter to drop all other fields
    > userCpu = cpu |> filter(fn:(r) => r._field == "usage_user")
    >
    > // Now we can use the `max` function to select the row that has the highest user cpu
    > // The max function picks only one row from each table it processes.
    > userCpu |> max()
    > run()
    > // That gives us the max user cpu per cpu and what ever tags we have, we are after the max cpu over all cpus.
    > // We can change the grouping before we perform the `max` operation to change how the tables are grouped so it searches across all cpus.
    > userCpu |> group(none:true) |> max()
    > run()
    >

Now let's look into windowing data into time buckets for processing. The `window` function is just a special kind of grouping operation that changes the table grouping to be aligned with windows. At this point you may have noticed that the data has two columns: _start and _stop. The convention is that the _start column contains a time value that represents the lower time bound for the row. Similarly the _stop column contains a time value that represents the exclusive upper time bound for the row. The range function added these columns to the data based on the range you passed it as well as grouping by these columns. The result is that by default rows for the same range of time belong in the same table. The window function provides a way to change grouping based on time by chaining the _start and _stop columns.

Let's see an example and let's use a different dataset to change things up a bit.

    > used_percent = from(bucket:"telegraf/autogen") |> range(start:-1m) |> filter(fn:(r) => r._measurement == "mem" and r._field == "used_percent")
    > used_percent |> window(every: 20s)
    > // The window data can be returned directly because its just a grouping of the data into tables.
    > // Can you see how the 20s windows became organized into table groups?
    > run()
    > 
    > // Now we can compute the mean of each window
    > used_percent |> window(every: 20s) |> mean()
    > run()
    > // We still have the same number of tables as we did before the mean operation,
    > // this is because most transformations do not change the grouping, but simply transform each table independently. 
    >
    > // If we want to combine all results into a single table we can use group again.
    > used_percent |> window(every: 20s) |> mean() |> group(none:true)
    > run()
    > // This result could be easily graphed to display the mean RAM usage over the last minute grouped into 20s windows.


At this point we have learned how to use various transformations:

* from
* range
* filter
* count
* max
* group
* window
* mean

The process of learning Flux is about learning each of the available transformations. This allows you to start with the bits that you need and then incrementally add to your knowledge as your needs grow.


Before we end, here is one last example to learn how to use two more operations `join` and `map`.  The `join` transformation take two table streams and combines them into a single stream of tables. The `map` transformation applies a function to each row of the table, producing a new row in the output table.

For this example we are going to join our memory used percent with the system uptime. Maybe we will find a correlation that indicates we have some kind of memory leak?

    > // First we need to get each sequence of tables
    > used_percent = from(bucket:"telegraf/autogen") |> range(start:-1m) |> filter(fn:(r) => r._measurement == "mem" and r._field == "used_percent")
    > run()
    > uptime = from(bucket:"telegraf/autogen") |> range(start:-1m) |> filter(fn:(r) => r._measurement == "system" and r._field == "uptime")
    > run()
    >
    > // Now for the join, we need to specify the tables to join and which columns to use when joining.
    > // Here we only care to join on _time.
    > joined = join(tables:{used_percent: used_percent, uptime:uptime}, on:["_time"])
    > run()
    > // Notice the result now has multiple _value columns, one for used_percent and uptime.
    > // At this point we could graph these results and visually inspect the data for some kind of correlation.
    >
    > // Perhaps the ratio of the two values is a meaningfull metric to visualize as well?
    > // Let's explore how we could do that using the `map` transformation.
    > joined |> map(fn:(r) => ({_value: r._value_used_percent / r._value_uptime}))
    > // We use an anonymous function to compute the ratio between used_percent and uptime.
    > // The columns of the table are the properties of the object `r` passed into the function.
    > // Finally, the result of this computation is stored as the _value column on the output table.
    > run()


Great! We now have the foundational transformations under our belt, `from`, `range`, `filter`, `group`, `window`, and `join`.
By composing those transformations we can transform and shape our data to get the results we need.

