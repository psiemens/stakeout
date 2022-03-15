package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

const accessAPI = "access.mainnet.nodes.onflow.org:9000"
const flowscanAPI = "https://flowscan.org/query"

func main() {

	args := os.Args[1:]

	if len(args) != 1 {
		fmt.Println("Pass your Flow address as an argument.\n\nExample:\n\nstakeout 0xe467b9dd11fa00df")
		return
	}

	address := flow.HexToAddress(args[0])

	c, err := client.New(accessAPI, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	delegationRecords := getDelegationRecords(ctx, address)

	fmt.Printf("You have delegated to:\n")

	for _, record := range delegationRecords {
		fmt.Printf("- Node: %s (Delegator: %d)\n", record.NodeID, record.DelegatorID)
	}

	fmt.Println()

	fmt.Printf("Rewards received:\n")

	grandTotal := uint64(0)

	for _, epoch := range epochs {
		epochRewards := getRewardsForEpoch(ctx, c, epoch.TxID, delegationRecords)

		total := uint64(0)

		for _, reward := range epochRewards {
			total += reward
		}

		grandTotal += total

		fmt.Printf(
			"- %s: %s FLOW\n",
			epoch.Date.Format("2006-01-02"),
			cadence.UFix64(total),
		)
	}

	fmt.Println()
	fmt.Printf("Grand total: %s FLOW\n", cadence.UFix64(grandTotal))
}

type Epoch struct {
	Date time.Time
	TxID string
}

var epochs = []Epoch{
	{
		newDate(2021, time.October, 12),
		"8f2d439ba31c7824989977b4883a1f5bd59adc347ab9d2f62d07a6639f59bd67",
	},
	{
		newDate(2021, time.October, 19),
		"ab8380881604ceae332783fa283b925b49f1c071e3e0eab7da1298570ed44c90",
	},
	{
		newDate(2021, time.October, 26),
		"f6fafde942e8b9538f1b92163e31f9752655fad6e69673cc1232829df81d5256",
	},
	{
		newDate(2021, time.November, 2),
		"cd9ad3758e9a08a1e9eb7dc0e7028de96ead624419403a30464c481476165d2a",
	},
	{
		newDate(2021, time.November, 9),
		"1cb1cf82f850d8f35d3e0114d9c5729dcfd7e9555e037ba4327bf7613263cf62",
	},
	{
		newDate(2021, time.November, 16),
		"993e78c383dc071d7d1bc10f4b387aed83a243115a77c3e1ff7f69cf0503cefe",
	},
	{
		newDate(2021, time.November, 23),
		"be839d4aab6a6443e1c6f16c77f7ee9134b3d923893c6ba8aa90b08647762bfe",
	},
	{
		newDate(2021, time.November, 30),
		"8bf3cb26e1d0996811c5855f4b99e1cea0c5244400f4d4e69fcfea2ec6f847ae",
	},
	{
		newDate(2021, time.December, 7),
		"044f8191fad43dd8b40554ed65781a0862e420d9c2c5640aa3166f8be791c84d",
	},
	{
		newDate(2021, time.December, 14),
		"df669cd5b615708e54d5589761906ca5137d4860f1f78a11ea3bed48ff458e82",
	},
	{
		newDate(2021, time.December, 22),
		"13ca79fc2fcb8adfb79cb7576bba5e7475e108f5c924083380b6958f5f56f58f",
	},
	{
		newDate(2021, time.December, 29),
		"9dcc4fff71e99b94a3dd90a953ef024d0cd9928e76fff69394134b6ef841ae21",
	},
}

func newDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

type DelegationRecord struct {
	NodeID      string
	DelegatorID uint32
}

func getRewardsForEpoch(
	ctx context.Context,
	c *client.Client,
	txID string,
	delegationRecords []DelegationRecord,
) []uint64 {
	tx, err := c.GetTransactionResult(
		ctx,
		flow.HexToID(txID),
		grpc.MaxCallRecvMsgSize(15000000),
	)
	if err != nil {
		panic(err)
	}

	return getDelegationRewards(tx.Events, delegationRecords)
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

type FlowscanTransactionResponse struct {
	Account struct {
		QueryResult struct {
			Count int
			Edges []struct {
				Node struct {
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

func (r FlowscanTransactionResponse) DelegationRecords() []DelegationRecord {
	records := make([]DelegationRecord, 0)

	for _, tx := range r.Account.QueryResult.Edges {
		for _, event := range tx.Node.Events.Edges {
			fields := event.Node.Fields

			delegatorID, err := strconv.ParseUint(fields[1].Value, 10, 32)
			if err != nil {
				panic(err)
			}

			record := DelegationRecord{
				NodeID:      fields[0].Value,
				DelegatorID: uint32(delegatorID),
			}

			records = append(records, record)
		}
	}

	return records
}

func getDelegationRecords(ctx context.Context, address flow.Address) []DelegationRecord {
	client := graphql.NewClient(flowscanAPI)

	req := graphql.NewRequest(`
		query AccountTransactionsQuery($address: ID!, $role: TransactionRole, $limit: Int!, $offset: Int) {
			account(id: $address) {
				queryResult: transactions(first: $limit, skip: $offset, role: $role) {
					count
					...AccountTransactionTableFragment
				}
			}
		}
		
		fragment AccountTransactionTableFragment on TransactionConnection {
			edges {
				node {
					hash
					time
					events(first: 10, skip: 0, type: ["A.8624b52f9ddcd04a.FlowIDTableStaking.NewDelegatorCreated"]) {
						edges {
							node {
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

	// TODO: implement pagination
	// currently it only scans the last 100 transactions on an account
	req.Var("limit", 100)
	req.Var("offset", 0)

	var response FlowscanTransactionResponse

	err := client.Run(ctx, req, &response)
	if err != nil {
		panic(err)
	}

	return response.DelegationRecords()
}
