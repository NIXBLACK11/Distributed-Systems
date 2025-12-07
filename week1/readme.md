# Week 1

## Progress
20%

## References
Youtube videos followed:
- [Introduction](https://www.youtube.com/watch?v=cQP8WApzIQQ&list=PLrw6a1wE39_tb2fErI4-WkMbsvGQk9_UB)
- []

Papers:
- [Map Reduce](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf)

## Notes:
### Basic
- Performance
- Fault Tolerence

Challenges
- Partial failure
- Concurrency

Lab
- Map Reduce
- Raft
- K/V Server
- Sharded K/V Service

Infrastructure
- Storage
- Communication
- Computation

### Map Reduce

MAp reduce for simple word count(Video):

```
Map(k, v)
    Split v into words
    for each word w
        emit(w, "1")

Reduce(k, v)
    emit(len(v))
```

#### Paper

The main aim i beleive is to achieve abstraction, i.e hide the messy details of parallelixation, fault-tolerance, data distribution and load balancing.
And representing them as simple map and reduce functions.


pseudo-code:
```
map(String key, String value):
    // key: document name
    // value: document contents
    for each word w in value:
        EmitIntermediate(w, "1");

reduce(String key, Iterator values):
    // key: a word
    // values: a list of counts
    int result = 0;
    for each v in values:
        result += ParseInt(v);
    Emit(AsString(result));
```

A good understanding we can get by this 
```
map (k1,v1) → list(k2,v2)
reduce (k2,list(v2)) → list(v2)
```

we see what values are passed and which of them are of same types


Steps to replicate:
1. Divide the data into M parts
2. There are M map tasks and R reduce tasks.
3. A map worker reads it's respective input slit and performs the user defined map task on it.
4. The data is then saved to a buffer, and then divided into R buckets.
5. When these buckets reach a threshold the data is saved n disks.
5. The master manages the map and reduce workers and knows where the M and R data is saved.
6. The reduce worker reads its respective partition data, sorts it so that same keys are together.
7. And then the reduce function is performed on them.
8. The master has multiple tables to handle the data and where it is stored.

Some points about failures in deterministicfunctions:
✔ Only one map output is accepted
✔ Only one reduce output is committed
✔ Failures don’t change the final result
✔ Output == what a sequential single-machine run would have produced

https://chatgpt.com/c/692ed453-47b4-8324-8255-4adff0b8762a

### RPCS and Threads

This is basic notes + a worker poll implementation and hopefully stress tested.
