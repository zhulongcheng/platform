filterRegex = /inodes*/

from(db:"test")
    |> range(start:-5m)
    |> filter(fn: (r) => r._field =~ filterRegex)
    |> max() 