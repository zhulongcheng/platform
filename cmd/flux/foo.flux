
/*
// test script "node_cpu{cpu=\"cpu1\",mode=\"system\"}"
 fromProm(host:"http://localhost:9090", matcher:`node_disk_written_bytes_total`)
  //  |> filter(fn: (r) => r.__name__ == "node_disk_written_bytes_total")
  |> range(start:-1h)
  |> derivative()
  |> window(every:10m)
  |> max()
  |> window(every:inf)
  |> map(fn: (r) => r._value / 1024.0)
  |> yield(name:"foo")
*/

// fromProm(host:"http://localhost:9090", matcher:`node_cpu{mode="idle"}`)


/*
fromProm(host:"http://localhost:9090")
  |> range(start:-1m)
  |> filter(fn: (r) => r.__name__ == "node_cpu" and r.cpu in ["cpu0", "cpu1"])
  |> yield(name:"foo")
*/

fromProm(host:"http://localhost:9090")
  |> range(start:-1m)
  |> filter(fn: (r) => r.__name__ == "node_cpu" and r.mode == "idle")
  |> max()
  |> group()
  |> sort(cols:["_value"])
  |> limit(n:3)
  |> yield(name:"foo")
