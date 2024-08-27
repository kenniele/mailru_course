package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	for _, curJob := range jobs {
		wg.Add(1)
		go func(job job, in, out chan interface{}) {
			defer wg.Done()
			job(in, out)
			close(out)
		}(curJob, in, out)
		out, in = make(chan interface{}, MaxInputDataLen), out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for v := range in {
		fmt.Printf("%v SingleHash data %v\n", v, v)
		newV := strconv.Itoa(v.(int))
		md5 := DataSignerMd5(newV)
		fmt.Printf("%v SingleHash md5(data) %v\n", v, md5)
		wg.Add(1)
		wgHash := sync.WaitGroup{}
		go func(v string) {
			defer wg.Done()
			result := make([]string, 2)
			wgHash.Add(2)
			go func(v string) {
				defer wgHash.Done()
				result[0] = DataSignerCrc32(newV)
			}(newV)
			go func(v string) {
				defer wgHash.Done()
				result[1] = DataSignerCrc32(md5)
			}(newV)
			wgHash.Wait()
			hash := strings.Join(result, "~")
			fmt.Printf("%v SingleHash crc32(data) %v\n", v, strings.Split(hash, "~")[0])
			fmt.Printf("%v SingleHash crc32(md5(data)) %v\n", v, strings.Split(hash, "~")[1])
			fmt.Printf("%v SingleHash result %v\n", v, hash)
			out <- hash
		}(newV)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for v := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			var result [6]string
			d := data.(string)
			wgHash := &sync.WaitGroup{}
			for i := 0; i < 6; i++ {
				wgHash.Add(1)
				go func(i int) {
					defer wgHash.Done()
					result[i] = DataSignerCrc32(strconv.Itoa(i) + d)
					fmt.Printf("%v Multihash: crc32(th+step1)) %v %v\n", d, i, result[i])
				}(i)
			}
			wgHash.Wait()
			hash := strings.Join(result[:], "")
			fmt.Printf("%v Multihash result: %v\n", d, hash)
			out <- hash
		}(v)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for v := range in {
		results = append(results, v.(string))
	}
	sort.Strings(results)
	result := strings.Join(results, "_")
	fmt.Printf("\nCombineResults %v", result)
	out <- result
}
