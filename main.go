package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/machinebox/graphql"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

const accessAPI = "access.mainnet.nodes.onflow.org:9000"
const flowscanAPI = "https://query.flowgraph.co/?token=7e11c53ae1f9cb4654408ebd2ba1fc4067613f3a"
const epochsCSV = "https://raw.githubusercontent.com/psiemens/stakeout/main/epochs.csv"

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

	address := flow.HexToAddress(args[0])

	year := *yearPtr

	// Fetch the epoch list from GitHub
	epochs, err := getEpochs()
	if err != nil {
		panic(err)
	}

	// Default start date is the first epoch defined above
	defaultStartDate := epochs[0].Time

	// Default end date is the current date
	defaultEndDate := time.Now()

	start, err := parseDate(*startPtr, defaultStartDate)
	if err != nil {
		panic("Invalid start date")
	}

	end, err := parseDate(*endPtr, defaultEndDate)
	if err != nil {
		panic("Invalid end date")
	}

	// If year flag is passed, set start and end dates to beginning and end of the year
	if year != 0 {
		start = newTimestamp(year, time.January, 1, 0, 0, 0)
		end = newTimestamp(year+1, time.January, 1, 0, 0, 0)
	}

	flowClient, err := client.New(accessAPI, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	flowscanClient := graphql.NewClient(flowscanAPI)

	ctx := context.Background()

	delegationRecords, err := getDelegationRecords(ctx, flowscanClient, address)
	if err != nil {
		panic(err)
	}

	fmt.Printf("You have delegated to:\n")

	for _, record := range delegationRecords {
		fmt.Printf("- Node: %s (Delegator: %d, Start Date: %s)\n", record.NodeID, record.DelegatorID, record.Time.Format("2006-01-02"))
	}

	fmt.Printf("\nRewards received:\n")
	fmt.Println("Epoch time (UTC), Transaction ID, Rewards (FLOW)")

	grandTotal := uint64(0)

	for _, epoch := range filterEpochs(epochs, start, end) {
		epochRewards, err := getRewardsForEpoch(ctx, flowClient, epoch.TxID, delegationRecords)
		if err != nil {
			panic(err)
		}

		total := uint64(0)

		for _, reward := range epochRewards {
			total += reward
		}

		grandTotal += total

		fmt.Printf(
			"%s,%s,%s\n",
			epoch.Time.Format("2006-01-02 15:04:05"),
			epoch.TxID,
			cadence.UFix64(total),
		)
	}

	fmt.Println()
	fmt.Printf("Grand total: %s FLOW\n", cadence.UFix64(grandTotal))
}

type Epoch struct {
	Time time.Time
	TxID string
}

func newTimestamp(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func parseTimestamp(v string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", v)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func parseDate(v string, defaultDate time.Time) (time.Time, error) {
	if v == "" {
		return defaultDate, nil
	}

	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func getEpochs() ([]Epoch, error) {
	resp, err := http.Get(epochsCSV)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(resp.Body)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	epochs := make([]Epoch, len(rows))

	for i, row := range rows {
		timestamp, err := parseTimestamp(row[0])
		if err != nil {
			return nil, err
		}

		epochs[i] = Epoch{
			Time: timestamp,
			TxID: row[1],
		}
	}

	return epochs, nil
}

func filterEpochs(epochs []Epoch, start, end time.Time) []Epoch {
	results := make([]Epoch, 0)

	for _, epoch := range epochs {
		if (epoch.Time.After(start) || epoch.Time.Equal(start)) && (epoch.Time.Before(end) || epoch.Time.Equal(end)) {
			results = append(results, epoch)
		}
	}

	return results
}

type DelegationRecord struct {
	NodeID      string
	DelegatorID uint32
	Time        time.Time
}

func getRewardsForEpoch(
	ctx context.Context,
	c *client.Client,
	txID string,
	delegationRecords []DelegationRecord,
) ([]uint64, error) {
	tx, err := getTransactionResult(ctx, c, txID)
	if err != nil {
		return nil, err
	}

	return getDelegationRewards(tx.Events, delegationRecords), nil
}

func getTransactionResult(
	ctx context.Context,
	c *client.Client,
	txID string,
) (*flow.TransactionResult, error) {
	var tx *flow.TransactionResult

	getTransactionResult := func() error {
		var err error

		tx, err = c.GetTransactionResult(
			ctx,
			flow.HexToID(txID),
			grpc.MaxCallRecvMsgSize(20000000),
		)
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(getTransactionResult, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	return tx, nil
}

const delegatorRewardsPaidEventType = "A.8624b52f9ddcd04a.FlowIDTableStaking.DelegatorRewardsPaid"

type DelegatorRewardsPaidEvent flow.Event

func (event DelegatorRewardsPaidEvent) NodeID() string {
	return event.Value.Fields[0].ToGoValue().(string)
}

func (event DelegatorRewardsPaidEvent) DelegatorID() uint32 {
	return event.Value.Fields[1].ToGoValue().(uint32)
}

func (event DelegatorRewardsPaidEvent) Amount() uint64 {
	return event.Value.Fields[2].ToGoValue().(uint64)
}

func getDelegationRewards(
	events []flow.Event,
	delegationRecords []DelegationRecord,
) []uint64 {
	rewards := make([]uint64, 0)

	for _, event := range events {
		if event.Type == delegatorRewardsPaidEventType {
			rewardsPaidEvent := DelegatorRewardsPaidEvent(event)

			for _, record := range delegationRecords {
				if rewardsPaidEvent.NodeID() == record.NodeID &&
					rewardsPaidEvent.DelegatorID() == record.DelegatorID {

					rewards = append(
						rewards,
						rewardsPaidEvent.Amount(),
					)
				}
			}
		}
	}

	return rewards
}

type FlowscanPageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type DelegationTransactionResponse struct {
	Account struct {
		Transactions struct {
			PageInfo FlowscanPageInfo
			Edges    []struct {
				Node struct {
					Time   string
					Events struct {
						Edges []struct {
							Node struct {
								Fields []struct {
									Type  string
									Value string
								}
							}
						}
					}
				}
			}
		}
	}
}

func (r DelegationTransactionResponse) DelegationRecords() ([]DelegationRecord, error) {
	records := make([]DelegationRecord, 0)

	for _, tx := range r.Account.Transactions.Edges {
		timestamp, err := time.Parse(time.RFC3339, tx.Node.Time)
		if err != nil {
			return nil, err
		}

		for _, event := range tx.Node.Events.Edges {
			fields := event.Node.Fields

			delegatorID, err := strconv.ParseUint(fields[1].Value, 10, 32)
			if err != nil {
				return nil, err
			}

			record := DelegationRecord{
				NodeID:      fields[0].Value,
				DelegatorID: uint32(delegatorID),
				Time:        timestamp,
			}

			records = append(records, record)
		}
	}

	return records, nil
}

func (r DelegationTransactionResponse) GetPageInfo() FlowscanPageInfo {
	return r.Account.Transactions.PageInfo
}

func getDelegationRecords(
	ctx context.Context,
	client *graphql.Client,
	address flow.Address,
) ([]DelegationRecord, error) {
	records := make([]DelegationRecord, 0)

	pageInfo := FlowscanPageInfo{
		HasNextPage: true,
	}

	for pageInfo.HasNextPage {
		var newRecords []DelegationRecord
		var err error

		newRecords, pageInfo, err = getDelegationRecordsPage(ctx, client, address, pageInfo.EndCursor)
		if err != nil {
			return nil, err
		}

		records = append(records, newRecords...)
	}

	return records, nil
}

func getDelegationRecordsPage(
	ctx context.Context,
	client *graphql.Client,
	address flow.Address,
	afterCursor string,
) ([]DelegationRecord, FlowscanPageInfo, error) {

	req := graphql.NewRequest(`
		query AccountTransactionsQuery($address: ID!, $role: TransactionRole, $first: Int!, $after: ID) {
			account(id: $address) {
				transactions(first: $first, after: $after, role: $role) {
					...AccountTransactionTableFragment
				}
			}
		}

		fragment AccountTransactionTableFragment on TransactionConnection {
			pageInfo {
				hasNextPage
				endCursor
			}
			edges {
				node {
					hash
					time
					events(typeId: "A.8624b52f9ddcd04a.FlowIDTableStaking.NewDelegatorCreated") {
						edges {
							node {
								type {
									id
								}
								fields
							}
						}
					}
					status
				}
			}
		}
	`)

	req.Var("address", "0x"+address.Hex())

	// 50 is the maximum page size
	req.Var("first", 50)

	if afterCursor != "" {
		req.Var("after", afterCursor)
	}

	var response DelegationTransactionResponse

	err := client.Run(ctx, req, &response)
	if err != nil {
		return nil, FlowscanPageInfo{}, err
	}

	delegationRecords, err := response.DelegationRecords()
	if err != nil {
		return nil, FlowscanPageInfo{}, err
	}

	return delegationRecords, response.GetPageInfo(), nil
}
