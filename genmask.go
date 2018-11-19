package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type params struct {
	URL      string
	Region   string
	Operator string
	Comment  bool
	Prefix   string
	Suffix   string
	Group    bool
}

func main() {
	p := readArgs()

	var values [][]string
	if p.Region != "" {
		values = filterRegion(parse(getCodes(p.URL)), p.Region)
	} else {
		values = parse(getCodes(p.URL))
	}

	if p.Operator != "" {
		values = filterOperator(values, p.Operator)
	}

	if p.Group {
		sort.Slice(values, func(i, j int) bool { return values[i][4] < values[j][4] })
	}
	op := ""
	for _, v := range values {
		_, min, max, dif := convert(v)
		if !validate(min, max, dif) {
			fmt.Printf("wrong interval: from %d to %d != %d\n", min, max, dif)
			continue
		}
		if p.Comment {
			fmt.Printf("; %v, %v, %v, %v, %v, %v\n", v[0], v[1], v[2], v[3], v[4], v[5])
		}
		if p.Group {
			if v[4] != op {
				fmt.Printf("; %s\n", v[4])
				op = v[4]
			}
		}
		if len([]rune(v[0])) != 3 || len([]rune(v[1])) != 7 || len([]rune(v[2])) != 7 {
			fmt.Printf("wrong interval: from %d to %d != %d\n", min, max, dif)
			continue
		}
		compute(p.Prefix+v[0], v[1], v[2], ""+p.Suffix)
	}
}

func getCodes(url string) string {
	var client http.Client
	res, err := client.Get(url)
	if err != nil {
		log.Panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Panic(err)
		}
		dec := charmap.Windows1251.NewDecoder()
		body := make([]byte, len(b)*2)
		n, _, err := dec.Transform(body, []byte(b), false)

		if err != nil {
			log.Println(err)
		}

		return string(body[:n])
	}
	log.Panic("bad response")
	return ""
}

// returns slice without row names
// element 0: code
// 1: min value of interval
// 2: max value of interval
// 3: interval length
// 4: cellular operator
// 5: region
func parse(data string) [][]string {
	codes := csv.NewReader(strings.NewReader(fixCodes(data, 6)))

	codes.LazyQuotes = true
	codes.Comma = ';'

	ext, err := codes.ReadAll()
	if err != nil {
		log.Println(err)
	}
	return ext[1:]
}

// I found three records with excess separators
// in https://rossvyaz.gov.ru/docs/articles/DEF-9x.csv
// (start at 955;5550000;5559999;10000 ...)
func fixCodes(data string, f int) string {
	c := 0
	var x []rune
	for _, v := range []rune(data) {
		if v == '\n' {
			c = 0
		}

		// replace excess separators with spaces
		if v == ';' {
			c++
			if c > f-1 {
				v = ' '
			}
		}
		x = append(x, v)
	}
	return string(x)
}

func filterRegion(values [][]string, region string) [][]string {
	var res [][]string
	for _, v := range values {
		if strings.Contains(strings.ToLower(v[5]), strings.ToLower(region)) {
			res = append(res, v)
		}
	}
	return res
}

func filterOperator(values [][]string, operator string) [][]string {
	var res [][]string
	for _, v := range values {
		if strings.Contains(strings.ToLower(v[4]), strings.ToLower(operator)) {
			res = append(res, v)
		}
	}
	return res
}

func validate(min, max, dif int) bool {
	if max-min == dif-1 {
		return true
	}
	return false
}

func convert(v []string) (pre, min, max, dif int) {
	pre, err := strconv.Atoi(v[0])
	if err != nil {
		log.Panic(err)
	}
	min, err = strconv.Atoi(v[1])
	if err != nil {
		log.Panic(err)
	}
	max, err = strconv.Atoi(v[2])
	if err != nil {
		log.Panic(err)
	}
	dif, err = strconv.Atoi(v[3])
	if err != nil {
		log.Panic(err)
	}
	return
}

type runes []rune

func increment(r runes) runes {
	if len(r) <= 1 {
		return r
	}
	var res runes
	i, err := strconv.Atoi(string(r))
	if err != nil {
		log.Panic(err)
	}
	i++
	res = runes(strconv.Itoa(i))
	if len(res) == len(r) {
		return res
	}

	res = res.reverse()
	l := len(res)

	for i := 0; i < len(r)-l; i++ {
		res = append(res, '0')

	}

	return res.reverse()
}

func (r runes) reverse() (rev []rune) {
	for i := len(r) - 1; i >= 0; i-- {
		rev = append(rev, r[i])
	}
	return
}

func hi(r runes) runes {
	var res runes
	for i := len(r) - 1; i >= 1; i-- {
		res = append(res, '9')
		if r[i] == '0' {
			continue
		}
		break
	}
	l := len(res)
	r = r.reverse()
	for i := l; i < len(r); i++ {
		res = append(res, r[i])
	}

	return res.reverse()
}

func decrement(r runes) runes {
	if len(r) <= 1 {
		return r
	}
	var res runes
	i, err := strconv.Atoi(string(r))
	if err != nil {
		log.Panic(err)
	}
	i--
	res = runes(strconv.Itoa(i))

	if len(res) == len(r) {
		return res
	}

	res = res.reverse()
	l := len(res)

	for i := 0; i < len(r)-l; i++ {
		res = append(res, '0')

	}

	return res.reverse()
}

func low(r runes) runes {
	var res runes
	res = append(res, r[0])
	for i := 2; i <= len(r); i++ {
		res = append(res, '0')
	}
	return res
}

func compute(pre, min, max, suf string) {

	// mask found if min and max are equal
	if min == max {
		fmt.Printf("%v%v%v\n", pre, min, suf)
		return
	}

	var prefix []rune
	mi := runes(min)
	ma := runes(max)

	if len(mi) != len(ma) {
		log.Panic("the length of min and max values is not equal")
	}

	for k, v := range ma {
		if v == mi[k] {
			prefix = append(prefix, v)
			continue
		}
		break
	}

	if l := len(prefix); l != 0 {
		compute(pre+string(prefix), string(mi[l:]), string(ma[l:]), suf)
		return
	}

	var suffix runes
	for k, v := range mi.reverse() {
		if ma.reverse()[k]-v == 9 {
			suffix = append(suffix, 'X')
			continue
		}
		break
	}

	if l := len(suffix); l != 0 {
		compute(pre,
			string(mi)[:len(mi)-l],
			string(ma)[:len(ma)-l],
			string(suffix)+suf)
		return
	}
	if len(mi) == 1 {
		compute(pre+"["+string(mi)+"-"+string(ma)+"]", "", "", suf)
		return
	}

	zc, err := strconv.Atoi(min)
	if err != nil {
		log.Panic(err)
	}

	if zc == 0 {
		compute(pre, string(mi), string(decrement(low(ma))), suf)
		compute(pre, string(low(ma)), max, suf)
		return
	}

	compute(pre, string(mi), string(hi(mi)), suf)

	if increment(hi(mi))[0] == mi[0] {
		compute(pre, string(increment(hi(mi))), max, suf)
	} else {
		if increment(hi(mi))[0] == ma[0] {
			compute(pre, string(increment(hi(mi))), max, suf)
		} else {
			compute(pre, string(increment(hi(mi))), string(decrement(low(ma))), suf)
			compute(pre, string(low(ma)), max, suf)
		}
	}
}

func readArgs() params {
	p := params{
		URL:      "https://rossvyaz.gov.ru/docs/articles/DEF-9x.csv",
		Region:   "",
		Operator: "",
		Comment:  false,
		Prefix:   "",
		Suffix:   "",
		Group:    false,
	}

	wait := false
	key := ""

	for _, v := range os.Args[1:] {
		if wait {
			switch key {
			case "-u":
				p.URL = v
			case "-r":
				p.Region = v
			case "-o":
				p.Operator = v
			case "-p":
				p.Prefix = v
			case "-s":
				p.Suffix = v
			}
			wait = false
		} else {
			if v == "-c" {
				p.Comment = true
				continue
			}
			if v == "-g" {
				p.Group = true
				continue
			}

			switch v {
			case "-u", "-r", "-o", "-p", "-s":
				key = v
				wait = true
			case "-h":
				help()
				os.Exit(0)
			default:
				fmt.Println("unknown option:", v)
				fmt.Println("show help: genmask -h")
				os.Exit(1)
			}
		}
	}
	if wait == true {
		fmt.Printf("missing value of %s argument\n", key)
		os.Exit(1)
	}
	return p
}

func help() {
	fmt.Println("usage: genmask [-u <url>] [-r <region filter>] [-c] [-p <prefix>] [-s <suffix>]")
	fmt.Println("\t-u <value>: url to csv file. Default is https://rossvyaz.gov.ru/docs/articles/DEF-9x.csv")
	fmt.Println("\t-r <value>: find entries in the csv file that contain the value in region field.")
	fmt.Println("\t            It's better to use short masks, because errors and typos are possible in the csv file.")
	fmt.Println("\t-o <value>: find entries in the csv file that contain the value in operator field.")
	fmt.Println("\t            It's better to use short masks, because errors and typos are possible in the csv file.")
	fmt.Println("\t-c         Print a comment: <; code, min, max, length, cellular operator, region> before each interval")
	fmt.Println("\t-p <value>: Print a prefix for each mask")
	fmt.Println("\t-s <value>: Print a suffix for each mask")
	fmt.Println("\t-g <value>: Group output by cellular operator")
	fmt.Println("show this help: genmask -h")
}
