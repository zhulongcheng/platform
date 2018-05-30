package functions

import "github.com/influxdata/platform/query"

func init() {
	query.RegisterBuiltIn("meta", metaBuiltins)
}

var metaBuiltins = `
tagValues = (db, key, predicate=(r) => true) =>
	from(db:db)
	  |> range(start:-24h)
	  // TODO: The default true predicate is not working
	  //|> filter(fn: predicate)
	  |> group(by:[key])
	  |> distinct(column:key)
	  |> group(by:["_stop","_start"])

tagKeys = (db, measurement) =>
	from(db:db)
		|> range(start:-24h)
		|> filter(fn:(r) => r._measurement == measurement)
		|> keys()

measurements = (db) =>
	tagValues(db:db, key:"_measurement")
`
