# doccleaner

# goal

This library can clean fields for a big set of documents thanks to given paths and clean functions.
A document could be json, bson (from gopkg.in/mgo.v2/bson) and more generaly map[string]interface{}.

# use cases

Some interesting use cases:

* anonymize some fields (password, username, credit cards...)
* fix some typos on a set of documents
* update dates of data set for testing purpose
* at least, all these changes can be done during a oriented document database export like MongoDB or Elasticsearch

# howto

## configuration 

For associating an element in document and a clean method, we need a configuration file (toml format):

```properties["leaf1"]
method="constantCleaner"
args=[]
["node2.leaf2"]
method="constantCleaner"
args=[]
["node2.leaf4"]
method="constantCleaner"
args=[]
["node3.node31.node311.node3111.leaf32"]
method="constantCleaner"
args=[]
["leaf4"]
method="constantCleaner"
args=[]
["node5.node51.node511.leaf5"]
method="constantCleaner"
args=[]
```

For each _X.Y.Z_ path, these rules are applied:

*  _X_,_Y_ and _Z_ could be any field names
* If _X_ is an array, librairy looks for _Y_ field on each array elements
* clean method is applied to _Z_ with given _args_.
* in example above, only _constantCleaner_ method is defined but properties can contain different clean methods.

## clean function

```go
// define a struct type
type constantValueCleaner struct {
}

// this type must implement clean method
func (c *constantValueCleaner) Clean(value interface{}) (changed interface{}, err error) {
  changed = 1234
  err = nil
  return 
}

```

This clean method only returns _1234_ constant for any given values.

## all together

```go
// init
propertiesReader := ... // properties reader (from string, file, ...)
cleaners := make(map[string]doccleaner.ValueCleaner) // map cleaner type names given in properties file with a real cleaner instance
cleaners["constantCleaner"] = &constantValueCleaner{} // fill cleaners map 
jsonCleaner := doccleaner.NewDocCleaner(propertiesReader,cleaners) // initialize the json cleaner

// call
jsonCleaner.Clean(objJson) // clean an unmarshalled json object
```

More example in [node_test.go|node_test.go]




