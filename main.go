package main

import "fmt"

func main() {
	err := ToJsons("/usr/local/var/db/redis/dump.rdb", "dump.json")
	fmt.Printf("error: %v\n", err)
}
