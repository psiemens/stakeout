package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/onflow/cadence"
)

func main() {

	yearPtr := flag.Int("year", 0, "Filter by epochs in this year (e.g. 2022)")
	startPtr := flag.String("start", "", "Filter by epochs after this date (e.g. 2021-04-27)")
	endPtr := flag.String("end", "", "Filter by epochs before this date (e.g. 2021-04-27)")

	flag.Parse()

	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("Pass your Flow address as an argument.\n\nExample:\n\nstakeout 0xe467b9dd11fa00df")
		return
	}

	address := args[0]

	year := *yearPtr
	start := *startPtr
	end := *endPtr

	// If year flag is passed, set start and end dates to beginning and end of the year
	if year != 0 {
		start = newTimestamp(year, time.January, 1, 0, 0, 0).Format("2006-01-02")
		end = newTimestamp(year+1, time.January, 1, 0, 0, 0).Format("2006-01-02")
	}

	// Start from the first epoch by default
	if start == "" {
		start = "2020-12-22"
	}

	// End at the current date by default
	if end == "" {
		end = time.Now().Format("2006-01-02")
	}

	rewards, err := getRewards(address, start, end)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println("Epoch time (UTC),Rewards (FLOW),Node ID,Delegator ID")

	for _, reward := range rewards {
		fmt.Printf(
			"%s,%s,%s,%s\n",
			reward.Timestamp.Format("2006-01-02 15:04:05"),
			reward.Amount,
			reward.NodeID,
			reward.DelegatorID,
		)
	}
}

const findLabsRewardsEndpoint = "https://api.findlabs.io/historical/api/rest/rewards"

func getRewards(address, start, end string) ([]RewardPayment, error) {
	client := resty.New()

	page := 0
	pageSize := 100

	rewards := make([]RewardPayment, 0)

	for {
		pageResult := &DelegationRewardsResult{}

		offset := strconv.Itoa(page * pageSize)

		_, err := client.R().
			SetQueryParams(map[string]string{
				"user":   address,
				"from":   start,
				"to":     end,
				"offset": offset,
			}).
			SetResult(pageResult).
			SetHeader("Accept", "application/json").
			Get(findLabsRewardsEndpoint)

		if err != nil {
			return nil, err
		}

		pageRewards := pageResult.DelegationRewards

		rewards = append(rewards, pageRewards...)

		if len(pageRewards) < pageSize {
			break
		}

		page += 1
	}

	// Sort rewards by date in increasing order
	sort.Slice(rewards, func(i, j int) bool {
		return rewards[i].Timestamp.Before(rewards[j].Timestamp)
	})

	return rewards, nil
}

type DelegationRewardsResult struct {
	DelegationRewards []RewardPayment `json:"delegation_rewards"`
}

type RewardPayment struct {
	NodeID      string
	DelegatorID string
	Height      uint64
	Amount      UFix64
	Timestamp   time.Time
}

type UFix64 struct {
	cadence.UFix64
}

func (u *UFix64) UnmarshalJSON(data []byte) error {
	s := string(data)

	n, err := cadence.NewUFix64(s)
	if err != nil {
		return err
	}

	*u = UFix64{n}

	return nil
}

func newTimestamp(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}
