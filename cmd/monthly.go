/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type Response struct {
	Data   []Data `json:"data"`
	Paging Paging `json:"paging"`
}

type Data struct {
	StartedOnWorkerAt time.Time `json:"started_on_worker_at"`
	FinishedAt        time.Time `json:"finished_at"`
	Status            int       `json:"status"`
}

type Paging struct {
	TotalItemCount int `json:"total_item_count"`
	PageItemLimit  int `json:"page_item_limit"`
}

// monthlyCmd represents the get command
var monthlyCmd = &cobra.Command{
	Use:   "monthly",
	Short: "月あたりのビルドアナリティクス",
	Long: `1ヶ月あたりの
・ビルド回数
・ビルド合計時間
・ビルド1回あたりの平均時間
を算出します。`,
	Run: func(cmd *cobra.Command, args []string) {
		ac := newAPIClient()

		now := time.Now()
		afterDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		beforeDate := afterDate.AddDate(0, 0, -1)

		rawURL := fmt.Sprintf("https://api.bitrise.io/v0.1/apps/%s/builds?before=%s&after=%s", os.Getenv("APP_SLUG_ID"), beforeDate.Unix(), afterDate.Unix())
		u, _ := url.Parse(rawURL)

		params := &apiParams{
			method: "GET",
			url:    u,
			header: os.Getenv("ACCESS_TOKEN"),
		}

		_, str, _ := ac.doRequest(params)

		var data Response
		err := json.Unmarshal([]byte(str), &data)
		if err != nil {
			log.Fatal(err)
		}

		var buildSumDuration time.Duration
		var buildTimesStatusOK int
		var buildTimesStatusError int
		var buildTimesStatusAborted int

		fmt.Println("--------- analytics ---------")
		fmt.Printf("ビルド回数(total)：%v回\n", data.Paging.TotalItemCount)
		for _, buildData := range data.Data {
			switch buildData.Status {
			case 1:
				buildTimesStatusOK++
				buildSumDuration += buildData.FinishedAt.Sub(buildData.StartedOnWorkerAt)
			case 2:
				buildTimesStatusError++
				buildSumDuration += buildData.FinishedAt.Sub(buildData.StartedOnWorkerAt)
			case 3:
				buildTimesStatusAborted++
			}
		}
		fmt.Printf("  > OK：%v回\n", buildTimesStatusOK)
		fmt.Printf("  > Error：%v回\n", buildTimesStatusError)
		fmt.Printf("  > Aborted：%v回\n", buildTimesStatusAborted)
		fmt.Printf("ビルド合計時間：%v\n", buildSumDuration)
		avarage := int(buildSumDuration) / (buildTimesStatusOK + buildTimesStatusError)
		fmt.Printf("ビルド平均タイム：%v\n", time.Duration(avarage))
		fmt.Println("--------- analytics ---------")
	},
}

func init() {
	rootCmd.AddCommand(monthlyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
