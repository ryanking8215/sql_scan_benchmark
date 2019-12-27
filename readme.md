# SQL Scan struct benchmark
Supported implements:
* sql
* sqlx
* gorm
* xorm 
* QueryToString -  query reuslt to string, convert string to each field of struct.

# result
Benchmark result on my desktop machine.
```
goos: darwin
goarch: amd64
pkg: sqlscan
BenchmarkSqlScan-4         	    1672	    714776 ns/op	   35584 B/op	    1905 allocs/op
BenchmarkSqlxScan-4        	    1579	    758367 ns/op	   36048 B/op	    1907 allocs/op
BenchmarkGormScan-4        	     705	   1660215 ns/op	  371559 B/op	    8209 allocs/op
BenchmarkXormScan-4        	     678	   1793351 ns/op	  200911 B/op	    7910 allocs/op
BenchmarkQueryToString-4   	    1437	    871171 ns/op	  137200 B/op	    4691 allocs/op
PASS
coverage: [no statements]
ok  	sqlscan	7.019s
Success: Benchmarks passed.
```