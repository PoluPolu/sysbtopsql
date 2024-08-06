package main

import (
    "database/sql"
    "fmt"
    "log"
    "regexp"
    "strconv"

    _ "github.com/lib/pq"
)

const (
    connStr = "user=postgres dbname=postgres sslmode=disable password=postgres port=5433" // Update with your credentials
)

var text = `
sysbench 1.0.18 (using system LuaJIT 2.1.0-beta3)

Running the test with following options:
Number of threads: 8
Initializing random number generator from current time

Prime numbers limit: 10000

Initializing worker threads...

Threads started!

CPU speed:
    events per second: 22884.34

General statistics:
    total time:                          10.0004s
    total number of events:              229004

Latency (ms):
         min:                                    0.35
         avg:                                    0.35
         max:                                    1.62
         95th percentile:                        0.35
         sum:                                79964.91

Threads fairness:
    events (avg/stddev):           28625.5000/17.12
    execution time (avg/stddev):   9.9956/0.00
`

func main() {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // // Create table if not exists
    // createTableQuery := `
    // CREATE TABLE IF NOT EXISTS sysbench_results (
    //     id SERIAL PRIMARY KEY,
    //     num_threads INT,
    //     prime_limit INT,
    //     events_per_second FLOAT,
    //     total_time FLOAT,
    //     total_events INT,
    //     latency_min FLOAT,
    //     latency_avg FLOAT,
    //     latency_max FLOAT,
    //     latency_95th FLOAT,
    //     latency_sum FLOAT,
    //     events_avg FLOAT,
    //     events_stddev FLOAT,
    //     exec_time_avg FLOAT,
    //     exec_time_stddev FLOAT
    // );`
    // _, err = db.Exec(createTableQuery)
    // if err != nil {
    //     log.Fatal(err)
    // }

    // Extract data using regex
    numThreads := extractInt(`Number of threads: (\d+)`, text)
    primeLimit := extractInt(`Prime numbers limit: (\d+)`, text)
    eventsPerSecond := extractFloat(`events per second: ([\d.]+)`, text)
    totalTime := extractFloat(`total time:\s+([\d.]+)s`, text)
    totalEvents := extractInt(`total number of events:\s+(\d+)`, text)
    latencyMin := extractFloat(`min:\s+([\d.]+)`, text)
    latencyAvg := extractFloat(`avg:\s+([\d.]+)`, text)
    latencyMax := extractFloat(`max:\s+([\d.]+)`, text)
    latency95th := extractFloat(`95th percentile:\s+([\d.]+)`, text)
    latencySum := extractFloat(`sum:\s+([\d.]+)`, text)
    eventsAvg, eventsStddev := extractTwoFloats(`events \(avg/stddev\):\s+([\d.]+)/([\d.]+)`, text)
    execTimeAvg, execTimeStddev := extractTwoFloats(`execution time \(avg/stddev\):\s+([\d.]+)/([\d.]+)`, text)

    // Insert data into the database
    insertQuery := `
    INSERT INTO sysbench_results (
        num_threads, prime_limit, cpu_eventspsec, total_time, total_events, latency_min, latency_avg, latency_max, latency_95th, latency_sum,
        threvents_avg, threvents_stddev, threxec_time_avg, threxec_time_stddev
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
    _, err = db.Exec(insertQuery, numThreads, primeLimit, eventsPerSecond, totalTime, totalEvents, latencyMin, latencyAvg, latencyMax, latency95th, latencySum,
        eventsAvg, eventsStddev, execTimeAvg, execTimeStddev)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Data inserted successfully")
}

func extractInt(pattern, text string) int {
    re := regexp.MustCompile(pattern)
    match := re.FindStringSubmatch(text)
    if len(match) > 1 {
        value, _ := strconv.Atoi(match[1])
        return value
    }
    return 0
}

func extractFloat(pattern, text string) float64 {
    re := regexp.MustCompile(pattern)
    match := re.FindStringSubmatch(text)
    if len(match) > 1 {
        value, _ := strconv.ParseFloat(match[1], 64)
        return value
    }
    return 0
}

func extractTwoFloats(pattern, text string) (float64, float64) {
    re := regexp.MustCompile(pattern)
    match := re.FindStringSubmatch(text)
    if len(match) > 2 {
        value1, _ := strconv.ParseFloat(match[1], 64)
        value2, _ := strconv.ParseFloat(match[2], 64)
        return value1, value2
    }
    return 0, 0
}