package instant

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// PrimeType is an answer Type
const PrimeType Type = "prime"

// Prime is an instant answer
type Prime struct {
	Answer
}

var rePrime *regexp.Regexp

func (p *Prime) setQuery(r *http.Request, qv string) Answerer {
	p.Answer.setQuery(r, qv)
	return p
}

func (p *Prime) setUserAgent(r *http.Request) Answerer {
	return p
}

func (p *Prime) setLanguage(lang language.Tag) Answerer {
	p.language = lang
	return p
}

func (p *Prime) setType() Answerer {
	p.Type = PrimeType
	return p
}

func (p *Prime) setRegex() Answerer {
	triggers := []string{
		"prime numbers", "prime number", "prime",
	}

	t := strings.Join(triggers, "|")
	p.regex = append(p.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	p.regex = append(p.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return p
}

func (p *Prime) solve(r *http.Request) Answerer {
	var start, end int

	matches := rePrime.FindStringSubmatch(p.remainder)
	if len(matches) == 3 {
		start, _ = strconv.Atoi(matches[1])
		end, _ = strconv.Atoi(matches[2])
		if end < start {
			start, end = end, start
		}

		primes := p.calculatePrimes(start, end)
		if len(primes) > 0 {
			p.Solution = strings.Join(primes, ", ")
		}
	}

	return p
}

func (p *Prime) tests() []test {
	tests := []test{
		{
			query: "prime numbers between 5 and 121",
			expected: []Data{
				{
					Type:      PrimeType,
					Triggered: true,
					Solution:  "5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113",
				},
			},
		},
		{
			query: "prime number between 614 and 537",
			expected: []Data{
				{
					Type:      PrimeType,
					Triggered: true,
					Solution:  "541, 547, 557, 563, 569, 571, 577, 587, 593, 599, 601, 607",
				},
			},
		},
		{
			query: "prime between -484 and 87",
			expected: []Data{
				{
					Type:      PrimeType,
					Triggered: true,
					Solution:  "2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83",
				},
			},
		},
		/*
			{
				query: "prime between 999764 and 1000351", // tests our max
				expected: []Data{
					{
						Type:      PrimeType,
						Triggered: true,
						Solution:  "999769, 999773, 999809, 999853, 999863, 999883, 999907, 999917, 999931, 999953, 999959, 999961, 999979, 999983",
						Err:       fmt.Errorf("Prime numbers greater than %d not returned", max),
						Cache:     true,
					},
				},
			},
		*/
	}

	return tests
}

var max = 1000000 // maybe a way to increase this? Perhaps focus on a range of 1M numbers, not a max???

// stolen from http://stackoverflow.com/questions/21854191/generating-prime-numbers-in-go
func (p *Prime) calculatePrimes(start, end int) []string {
	var x, y, n int

	if end > max {
		end = max
		p.Err = fmt.Errorf("Prime numbers greater than %d not returned", max)
	}
	min := 1
	if start < min { // Prime numbers are usually considered to be positive
		start = min
	}
	nsqrt := math.Sqrt(float64(end))
	isPrime := make([]bool, end)
	for x = 1; float64(x) <= nsqrt; x++ {
		for y = 1; float64(y) <= nsqrt; y++ {
			n = 4*(x*x) + y*y
			if n <= end && (n%12 == 1 || n%12 == 5) {
				isPrime[n] = !isPrime[n]
			}
			n = 3*(x*x) + y*y
			if n <= end && n%12 == 7 {
				isPrime[n] = !isPrime[n]
			}
			n = 3*(x*x) - y*y
			if x > y && n <= end && n%12 == 11 {
				isPrime[n] = !isPrime[n]
			}
		}
	}

	for n = 5; float64(n) <= nsqrt; n++ {
		if isPrime[n] {
			for y = n * n; y < end; y += n * n {
				isPrime[y] = false
			}
		}
	}

	// necessary
	isPrime[2] = true
	isPrime[3] = true

	primes := []string{}
	for x = start; x < len(isPrime)-1; x++ {
		if isPrime[x] {
			primes = append(primes, strconv.Itoa(x))
		}
	}

	return primes
}

func init() {
	rePrime = regexp.MustCompile(`^between (-?[0-9]+) and (-?[0-9]+)`)
}
