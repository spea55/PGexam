package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	filename         = "log.txt"
	text             string
	stop             = make(chan bool)
	timeout_count    int
	overloaded_ping  int
	overloaded_count int
)

type Log struct {
	dateAtime []time.Time
	ip        []string
	ping      []string
}

func main() {
	var log Log

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("enter timeout count, overloaded ping, overloaded count.")
	scanner.Scan()
	slices := strings.Split(scanner.Text(), ",")

	timeout_count, _ = strconv.Atoi(slices[0])
	overloaded_ping_str := slices[1]
	overloaded_count, _ = strconv.Atoi(slices[2])
	overloaded_ping, _ = strconv.Atoi(overloaded_ping_str)

	err := log.read_log_file()
	if err != nil {
		fmt.Println(os.Stderr)
		os.Exit(1)
	}

	log.server_check()
	<-stop
}

func (log *Log) server_check() {

	go func() {
		log.isOverloaded()
		stop <- true
	}()

	for i, ping := range log.ping {
		if ping == "-" {
			log.isBreakServer(i)
		}
	}
}

func (log *Log) isBreakServer(index int) {
	count := 1
	//ip := strings.Split(log.ip[index], "/")
	//net_prefix := ip[1]
	for j := index + 1; j < len(log.ip); j++ {
		// if strings.HasSuffix(log.ip[j], net_prefix) && log.ping[j] == "-" {

		// }else
		if log.ip[index] == log.ip[j] {
			if log.ping[j] == "-" {
				count += 1
			} else {
				if count == timeout_count {
					text = "ip:" + log.ip[j] + ",status:break, start:" + log.dateAtime[index].String() + ", end:" + log.dateAtime[j].String() + "\n"
					write_file()
				}
				break
			}
		}
	}
}

func (log *Log) isOverloaded() {
	var sum, ping, i, end, start int
	log_ip_diffs := make([]Log, 0)
	var slices Log
	var ip string

	for i, ip = range log.ip {
		index := find_index(log_ip_diffs, ip)
		log_add := Log{}
		if index == -1 {
			log_add.dateAtime = append(log_add.dateAtime, log.dateAtime[i])
			log_add.ip = append(log_add.ip, log.ip[i])
			log_add.ping = append(log_add.ping, log.ping[i])
			log_ip_diffs = append(log_ip_diffs, log_add)
		} else {
			log_ip_diffs[index].dateAtime = append(log_ip_diffs[index].dateAtime, log.dateAtime[i])
			log_ip_diffs[index].ip = append(log_ip_diffs[index].ip, log.ip[i])
			log_ip_diffs[index].ping = append(log_ip_diffs[index].ping, log.ping[i])
		}
	}

	for i := range log_ip_diffs {
		fmt.Println(i)
		for j := range log_ip_diffs[i].ping {
			fmt.Println(log_ip_diffs[i].dateAtime[j], log_ip_diffs[i].ip[j], log_ip_diffs[i].ping[j])
		}
	}

	for _, slices = range log_ip_diffs {
		end = 0
		sum = 0
		ping = 0
		length := len(slices.ip)
		for i = length - 1; i >= length-overloaded_count-1; i-- {
			if slices.ping[i] != "-" {
				ping, _ = strconv.Atoi(slices.ping[i])
			} else {
				ping = 1000
			}
			sum += ping
		}

		if sum/overloaded_count >= overloaded_ping {
			for i = length - 1; i >= 0; i-- {
				if slices.ping[i] != "-" {
					ping, _ = strconv.Atoi(slices.ping[i])
				}

				if ping >= overloaded_ping || slices.ping[i] == "-" {
					if end == 0 {
						end = i
						fmt.Println(end, slices.ip[end], slices.ping[end])
					}
				}
			}

			ping, _ = strconv.Atoi(slices.ping[length-overloaded_count])
			if ping >= overloaded_ping {
				for i = length - overloaded_count - 1; i >= 0; i-- {
					if ping <= overloaded_ping || slices.ping[i] != "-" {
						start = i + 1
					}
				}
			} else {
				for i = length - overloaded_count + 1; i <= length - 1; i++ {
					if ping >= overloaded_ping || slices.ping[i] == "-" {
						start = i
					}
				}
			}

			text = "ip:" + slices.ip[0] + ",status:overloaded, start:" + slices.dateAtime[start].String() + ", end:" + slices.dateAtime[end].String() + "\n"
			write_file()
		}
	}
}

func (log *Log) read_log_file() error {
	var date time.Time
	fp, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error: %s", "filename is empty.")
	}
	defer fp.Close()

	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		slice := strings.Split(sc.Text(), ",")
		date = string_to_date(slice[0])
		log.dateAtime = append(log.dateAtime, date)
		log.ip = append(log.ip, slice[1])
		log.ping = append(log.ping, slice[2])
	}

	return nil
}

func write_file() {
	file, err := os.OpenFile("output.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		panic(err)
	}
}

func string_to_date(date_s string) time.Time {
	const format = "2006-01-02 15:04:05"
	year, _ := strconv.Atoi(date_s[:4])
	month, _ := strconv.Atoi(date_s[4:6])
	day, _ := strconv.Atoi(date_s[6:8])
	hour, _ := strconv.Atoi(date_s[8:10])
	min, _ := strconv.Atoi(date_s[10:12])
	seconds, _ := strconv.Atoi(date_s[12:])

	date := time.Date(year, time.Month(month), day, hour, min, seconds, 0, time.UTC)
	return date
}

func find_index(query []Log, str string) int {
	for i, _ := range query {
		for _, ip := range query[i].ip {
			if ip == str {
				return i
			} else {
				break
			}
		}
	}

	return -1
}
