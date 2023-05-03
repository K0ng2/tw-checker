package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var version string

type TwitterAPI []struct {
	ScreenName string `json:"screen_name"`
	Protected  bool   `json:"protected"`
}

var rootCmd = &cobra.Command{
	Use:     "tw-checker",
	Version: version,
}

var userCmd = &cobra.Command{
	Use:  "user [username]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showResults(args)
	},
}

var fileCmd = &cobra.Command{
	Use:  "file [file]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return readFile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(userCmd, fileCmd)
}

var status = map[int]string{
	1: "Active",
	2: "Suspended or Not found",
	3: "Protected",
}

func readFile(filePath string) error {
	var lines []string

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err = sendListInBatches(lines)
	if err != nil {
		return err
	}

	return nil
}

func sendListInBatches(list []string) error {
	for i := 0; i < len(list); i += 100 {
		end := i + 100
		if end > len(list) {
			end = len(list)
		}
		err := showResults(list[i:end])
		if err != nil {
			return err
		}
	}
	return nil
}

func showResults(username []string) error {
	queries := map[string]string{
		"screen_name": strings.Join(username, ","),
	}
	headers := map[string]string{
		"authorization": "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
	}

	info := new(TwitterAPI)
	err := request("https://api.twitter.com/1.1/users/lookup.json", queries, headers, info)
	if err != nil {
		return err
	}

	var ss int
	for _, s := range username {
		ss = 2
		for _, r := range *info {
			if s == r.ScreenName {
				ss = 1
				if r.Protected {
					ss = 3
				}
				break
			}
		}

		var output string
		switch ss {
		case 1:
			output = color.GreenString("%s", status[ss])
		case 2:
			output = color.RedString("%s", status[ss])
		case 3:
			output = color.YellowString("%s", status[ss])
		}
		fmt.Printf("%-35s%s\n", output, s)
	}

	return nil
}

func request(url string, query map[string]string, headers map[string]string, t interface{}) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	h := req.Header
	for j, s := range headers {
		h.Add(j, s)
	}

	q := req.URL.Query()
	for j, s := range query {
		q.Add(j, s)
	}

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	rootCmd.Execute()
}
