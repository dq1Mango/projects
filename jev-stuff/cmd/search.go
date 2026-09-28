package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dq1Mango/projects/jev-stuff"
)

const ENVIRONMENT_TOKEN = "JEV_TOKEN"

const QUESTION_ID = "relevance"

const COST_PER_TOKEN = 0.42 / 1e6

const Template = `
	{
    "state": %s,
    "model": "jev-latest",
    "questions": {
      "relevance": {
        "type": "noul",
        "instructions": {
					"query": "%s",
					"question": "is this message relevant to the query: ` + "`query`" + `?"
				}
      }
    }
  }
`

type Message struct {
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
}

var TestingMessage = Message{
	Content:   "we are javing a bbq next wednesday, do you want to come?",
	Sender:    "best friend",
	Timestamp: time.Now(),
}

func exampleQuery(jev *jev.Jev, query string) {
	state, err := json.Marshal(TestingMessage)
	if err != nil {
		panic(err)
	}

	rendered := fmt.Sprintf(Template, string(state), query)

	var pretty bytes.Buffer
	// json.Compact(&pretty, []byte(rendered))
	json.Indent(&pretty, []byte(rendered), "", "  ")
	fmt.Println(pretty.String())

	answers, err := jev.SystemOne([]byte(rendered))

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Printf("Noul: %f\n", answers.Nouls["relevance"].Noul)
	fmt.Printf("Input Tokens: %d\n", answers.Usage.InputTokens)
}

func sendQuery(jev *jev.Jev, message Message, query string) (*jev.UsefullResponse, error) {
	state, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	rendered := fmt.Sprintf(Template, string(state), query)

	return jev.SystemOne([]byte(rendered))
}

type workerResult struct {
	noul    float64
	message Message
	tokens  int
}

func searchWorker(
	jev *jev.Jev,
	query string,
	jobs <-chan Message,
	results chan<- workerResult,
) {
	for msg := range jobs {
		resp, err := sendQuery(jev, msg, query)

		if err != nil {
			fmt.Println("somethings gone wrong: ", err)
			continue
		}

		results <- workerResult{
			noul:    float64(resp.Nouls[QUESTION_ID].Noul),
			message: msg,
			tokens:  resp.Usage.InputTokens,
		}
	}
}

type match struct {
	Message Message `json:"message"`
	Noul    float64 `json:"noul"`
}

func search(
	j *jev.Jev,
	messages []Message,
	query string,
	threshold float64,
	workers int,
) ([]match, int, error) {
	jobs := make(chan Message)
	results := make(chan workerResult, len(messages))

	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() { searchWorker(j, query, jobs, results) })
	}

	for _, msg := range messages {
		jobs <- msg
	}

	close(jobs)

	wg.Wait()

	close(results)

	matches := make([]match, 0, len(results))
	var tokens int

	for res := range results {
		if res.noul >= threshold {
			matches = append(matches, match{Message: res.message, Noul: res.noul})
		}

		tokens += res.tokens
	}

	return matches, tokens, nil
}

func main() {
	query := flag.String("query", "", "the query to execute [required]")
	dataset := flag.String("dataset", "", "path to the data set to be queried")
	threshold := flag.Float64("threshold", 0.75, "coorelation needed for a match")
	workers := flag.Int("workers", 10, "how many worker threads to spawn for queries")
	out := flag.String("out", "", "file to write matches to (leave empty for stdout)")

	flag.Parse()

	if *query == "" {
		flag.Usage()
		os.Exit(1)
	}

	token, exists := os.LookupEnv(ENVIRONMENT_TOKEN)

	if !exists {
		fmt.Fprintf(os.Stderr, "Please set %s\n", ENVIRONMENT_TOKEN)
		os.Exit(1)
	}

	jev := jev.NewJev(token)

	if *dataset == "" {
		exampleQuery(jev, *query)
		return
	}

	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, "--workers must be > 0")
		os.Exit(1)
	}

	file, err := os.ReadFile(*dataset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read dataset, %s; %s", *dataset, err.Error())
		os.Exit(1)
	}

	var outfile *os.File
	if *out != "" {
		outfile, err = os.Create(*out)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create output file, %s; %s", *out, err.Error())
			os.Exit(1)
		}

		defer outfile.Close()

	} else {
		outfile = os.Stdout
	}

	var data []Message
	err = json.Unmarshal(file, &data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not unmarshal dataset; %s", err.Error())
		os.Exit(1)
	}

	start := time.Now()

	matches, tokens, err := search(jev, data, *query, *threshold, *workers)

	elapsed := time.Since(start)

	type Output struct {
		Tokens  int     `json:"tokens"`
		Cost    float32 `json:"cost"`
		Time    string  `json:"time"`
		Matches []match `json:"matches"`
	}

	cost := float32(tokens) * COST_PER_TOKEN

	output := Output{
		Tokens:  tokens,
		Cost:    cost,
		Time:    elapsed.String(),
		Matches: matches,
	}

	marshall_mathers, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, marshall_mathers, "", "\t")
	if err != nil {
		panic(err)
	}

	prettyResults(tokens, cost, elapsed)

	outfile.Write(pretty.Bytes())
}

func prettyResults(tokens int, cost float32, duration time.Duration) {
	fmt.Println()
	fmt.Println("---------- Results ----------")
	fmt.Printf("Tokens: %d\n", tokens)
	fmt.Printf("Cost: $%.4f\n", cost)
	fmt.Printf("Duration: %s\n", duration.String())
	fmt.Println()
}
