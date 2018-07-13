regexFunc = (regLiteral) => 
   from(db:"testdb")
     |>  range(start:-5m)
     |>  filter(fn: (r) => r._field =~ regLiteral)
     |>  max() 

regexFunc(regLiteral: /inodes*/)