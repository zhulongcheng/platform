option now = () => 2018-05-22T19:54:16Z

from(db: "test")
    |> range(start:-40s, stop: -10s)
    |> window(every: 10s)
    |> sum() 