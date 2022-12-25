package main

import (
	"context"
	"flag"
	"fmt"
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

func main() {

	yearPtr := flag.Int("year", 0, "Filter results by this year")

	flag.Parse()

	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("Pass your Flow address as an argument.\n\nExample:\n\nstakeout 0xe467b9dd11fa00df")
		return
	}

	address := flow.HexToAddress(args[0])

	year := *yearPtr

	c, err := client.New(accessAPI, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	delegationRecords := getDelegationRecords(ctx, address)

	fmt.Printf("You have delegated to:\n")

	for _, record := range delegationRecords {
		fmt.Printf("- Node: %s (Delegator: %d, Start Date: %s)\n", record.NodeID, record.DelegatorID, record.Time.Format("2006-01-02"))
	}

	fmt.Println()

	fmt.Printf("Rewards received:\n")

	grandTotal := uint64(0)

	for _, epoch := range filterEpochsByYear(epochs, year) {
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
		newDate(2020, time.December, 22),
		"1ab22491777fb3c75333135c82b90a9f56ede8e0d18d43ecf6c5589d06353f95",
	},
	{
		newDate(2020, time.December, 29),
		"d7a8daca55dcadddd21dc9b63e71c515032584e9cf126b74744f676a1196e3c8",
	},
	{
		newDate(2021, time.January, 5),
		"9e377e6818d19b7a26d65b076120b2681d512111d397c617cac0843060437fc9",
	},
	{
		newDate(2021, time.January, 12),
		"29886f7cdbbb47b96dfa05ad277fcb3c8876c4ee32954269a926061aa4157af7",
	},
	{
		newDate(2021, time.January, 19),
		"83d585f3fa1368258cf317690e4bf82d4f841bfe2ccbc6c191838b73371b6e66",
	},
	{
		newDate(2021, time.January, 26),
		"487759099a09fc088943aac45909cca37d6b5bed418feab502ff37bd1b93346f",
	},
	{
		newDate(2021, time.February, 2),
		"c2f29629040ae6fca12e103f676bf112200f24f0fc8b4cb99c92f9287ce6195f",
	},
	{
		newDate(2021, time.February, 9),
		"e29a22e29297fcd8f67a030ae641405e14a889167a639f54b05f48fe1142bd5b",
	},
	{
		newDate(2021, time.February, 16),
		"d775735a1d9a8574faaa2b82b9dea0833d5bf48e43845c9c6bf5a42eb96414f2",
	},
	{
		newDate(2021, time.February, 23),
		"eed4f38fe4fec2bf9b60eba7708ff7d41fa654942ccc20cb1fc86872813f5b93",
	},
	{
		newDate(2021, time.March, 2),
		"3650b50994719480b8358a507afa30d58c232a30a6cfd947a94b62a78d25c849",
	},
	{
		newDate(2021, time.March, 9),
		"e4db187672cf545c865850840ff66d34c43cb298d8aae687c209c7f2fef44ea1",
	},
	{
		newDate(2021, time.March, 16),
		"2791557e2bdf4be66f69eeb919c5c3da94f2af4baf9c1e6953aedd887715bd7e",
	},
	{
		newDate(2021, time.March, 23),
		"15ac2d8ee73eae3483da10370dcc6f432d928d4631be22f8c7f0f85a8a62e1fd",
	},
	{
		newDate(2021, time.March, 30),
		"a60e812b005100290a4657f7022f1e2d3dc4745e25a0373b82d807c77471e069",
	},
	{
		newDate(2021, time.April, 6),
		"7674a088faa2b30ace3c798c2273face4e2b41970ca1990ad9d96a1bb7e2c2b2",
	},
	{
		newDate(2021, time.April, 13),
		"4addbcc17f8a5e9ecf1eb0d78a5c59332cde9d11ce347da676009284247565dd",
	},
	{
		newDate(2021, time.April, 20),
		"86917dcf3055d1aa21d898ccd4b8d1f55fd0924c7f5b7ce708dd92e69094e83d",
	},
	{
		newDate(2021, time.April, 27),
		"5c46cc3ff35a522d9439f26023e11a0f8697bf04b418e9e019a95ab3f49231ce",
	},
	{
		newDate(2021, time.May, 4),
		"7c79febe209178851e4c6b0c1d9bf89955f89783ea8f9ace65e4c442ddd16e7d",
	},
	{
		newDate(2021, time.May, 11),
		"df60fda0991410b871b80437840e461ecd5b35decc6da201a1e5d62ae23b0cec",
	},
	{
		newDate(2021, time.May, 18),
		"a0ea12fe8166b3d3b10cb221cbc3dd236dba5990274e9a90c32e6a904f72d96f",
	},
	{
		newDate(2021, time.May, 25),
		"5a3093f7e577ba8e6dc3c6241c7b3a0c09d3068a91a684a477277d4f792e19da",
	},
	{
		newDate(2021, time.June, 1),
		"6e97bf58f0261bbf06733c7b225f56f32032dd974e73aa8c2c72e2fddd6ab4f9",
	},
	{
		newDate(2021, time.June, 8),
		"8d06b6d75fd315cdbe0caf035fb42b8503aef5f12b875fb18c38b9329cf10b83",
	},
	{
		newDate(2021, time.June, 15),
		"9b86013070876210c5ff04c5518d7698d21be2c3d4525cf985a82b311e7bb217",
	},
	{
		newDate(2021, time.June, 22),
		"207f908a7e949c1094dee63f71bf9f8487e07d8a2d904265b358ee03dc59441a",
	},
	{
		newDate(2021, time.June, 29),
		"f9cc19931aa131511951ff339dd3c454cdc7d956611cf0271e68c0ed50f6fca2",
	},
	{
		newDate(2021, time.July, 6),
		"406c3fe36fb4eb05cb3ba830cb0a8b66b6fc6d6f3d56e237eef201fe3b7b9821",
	},
	{
		newDate(2021, time.July, 13),
		"ade3ad62894c99798876b0c8fb9db84e978a14e23095944abcf76d614e5c018b",
	},
	{
		newDate(2021, time.July, 20),
		"51dab505c0bd843af3c3e43d5502cdc21c2f6dbb5765383b0a285f0578f859d4",
	},
	{
		newDate(2021, time.July, 27),
		"8730f4ff3552ca1e47777e51eda0169f49880bf49da19106839a8e20ae6bbec6",
	},
	{
		newDate(2021, time.August, 3),
		"d8a85bdd767f56a63bcb627bd74b1fe47f6e0d50b9c19b32c47895ca8373d208",
	},
	{
		newDate(2021, time.August, 10),
		"80366137eb2665d52480b6d9592decb11c98dcb38657633673173503d94db711",
	},
	{
		newDate(2021, time.August, 17),
		"38364caf7310ebac175d40fe99b9e3ac636eb0d370e9f647f92bb3fe938915ff",
	},
	{
		newDate(2021, time.August, 24),
		"4a6a3c043f22661d32a1b894e89744d2427f35dfe590d7f563665502afd03075",
	},
	{
		newDate(2021, time.August, 31),
		"109838b1910c72484e2789b5e60cfbab2e2a75da354b5e378397dae72f97e73e",
	},
	{
		newDate(2021, time.September, 7),
		"4bd3dc0284bfd077eb5eac7d39afd01df3e8125c45b07e2d3ea7e7e8512e09ef",
	},
	{
		newDate(2021, time.September, 14),
		"1746c7cf8e39e292c779a7891d788b4b8848b8ce5c8432007b7787fa262e286d",
	},
	{
		newDate(2021, time.September, 21),
		"b5f214c3284e15b8e13f178582db069350ace98a81e086461da5977050a70962",
	},
	{
		newDate(2021, time.September, 28),
		"a819c6ebd9bf8c969be0d43ce1f57a803c4fb4cba1e92ebd1de6403da73d426e",
	},
	{
		newDate(2021, time.October, 5),
		"10fd2d7642ec87e61a8222491297c9462c6fdc3d331662ff7115a0e1becb3e21",
	},
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
	{
		newDate(2022, time.January, 5),
		"ae25c41e18a798e565c9cc1d0ed4afe46bd75cad5b7ce04d92ec8515314d85b3",
	},
	{
		newDate(2022, time.January, 12),
		"c1541b18ddd377983be4eda812100fb4b4baf5face603e6616e0f7d0391f8ddb",
	},
	{
		newDate(2022, time.January, 19),
		"96e5ecf9943aecdacc40c5b99fd0163bc917997d9da341d62364593c0ea729b3",
	},
	{
		newDate(2022, time.January, 26),
		"27c4302f4ff4ea9d50ec8f368705f77e897d20f83212ce5e6e1c58ecb784e853",
	},
	{
		newDate(2022, time.February, 2),
		"e8d6f108901dfe67f89ea43e588e77e379bc465925dfb725bc07db5e665b0ab2",
	},
	{
		newDate(2022, time.February, 9),
		"ba593876491dc3a53a8a33dd2709c25ce474e84a9d4d228d4309e1a7945a2e71",
	},
	{
		newDate(2022, time.February, 16),
		"adc2c23f4a6e3c1bf2ba93d92adfc3dbbc7f9821226abb8576abf8f195d5a64e",
	},
	{
		newDate(2022, time.February, 23),
		"b8329b2c39141ec9ee2a0fc3ba1ab8372fabcac8864ba54bd39355df8df0c3ef",
	},
	{
		newDate(2022, time.March, 2),
		"ce5c402de62dbd6ccbe764739a6dae1187e751faf9a9179854eae31e9c9e4d26",
	},
	{
		newDate(2022, time.March, 9),
		"64a8781b0c6873d98ec30d5ef6ee296dcdddf93a8c2ec2e4378a6cfaea6b2631",
	},
	{
		newDate(2022, time.March, 16),
		"1f82f2b99296348450c9c7bd5da0a28b7d9b4d9b382c317aef1b782e22b324ca",
	},
	{
		newDate(2022, time.March, 23),
		"7b341be43ba69889f2ef37477c52b15c1ce18350815208a21f17728631dda1e0",
	},
	{
		newDate(2022, time.March, 30),
		"9941aac5fd7d280901fdda258388f455a9ec3d75bf910d1e161a7657f73e1c01",
	},
	{
		newDate(2022, time.April, 6),
		"1fd4282c4ba3194f6bcc6f90661649cf356da25186f4990f43e3ff7732486e7e",
	},
	{
		newDate(2022, time.April, 13),
		"99f684594fcb73a91b37ea5fcd9de5c2c51d4feac5a0a08a6d799ed7933928bc",
	},
	{
		newDate(2022, time.April, 20),
		"c0a96fe2f03088dafd209ef35965c67965d0c5a323ee28eb0c91965535f4546e",
	},
	{
		newDate(2022, time.April, 27),
		"1e75bcd5b38b498b3eb622d11e6e65d2f315f1584ed8e6eb3164f85784f979d7",
	},
	{
		newDate(2022, time.May, 4),
		"0b6a8f6b23949659338b7f68ee19aa4342ec104dbcf4148f7341e77cb192d0d7",
	},
	{
		newDate(2022, time.May, 11),
		"34e207b07960c467213136b1d3b7e6a00e404b2ff312d2d2e764b3ec57e4e4fe",
	},
	{
		newDate(2022, time.May, 18),
		"9a0aaa02902cf71da5f62d5a19be2fc65e1987a4c265da8167ef56deac948d18",
	},
	{
		newDate(2022, time.May, 25),
		"327a0a26aa75107189da320655fc9978f96109917c1c3ade571dbada1d759adf",
	},
	{
		newDate(2022, time.June, 1),
		"8c5b29259f707c05e75018724d91ea644263567e8bc9f90b2a30d4bf6edb67fc",
	},
	{
		newDate(2022, time.June, 8),
		"5227c6b14ec518e96c78825eba41ca1b69a43de09e24e606ee95ea984d4db51b",
	},
	{
		newDate(2022, time.June, 15),
		"33e091751bcedf2b562d9f86e47602dcb5307a0e2ab7c08699d7a35592a2e78b",
	},
	{
		newDate(2022, time.June, 22),
		"fd044369db1be80ad9dcf8ba18d27ecdd85c0ec7728f142d875fb1620e711273",
	},
	{
		newDate(2022, time.June, 29),
		"69ee71740019a5a48b74880cc91ff3748adf1b732445ecdde0564ac939d81d93",
	},
	{
		newDate(2022, time.July, 6),
		"d7861b9657fd5bd2190bf963b4db1099e49a184efc64e12af96974679cc93773",
	},
	{
		newDate(2022, time.July, 13),
		"704df55da29d2e9c07159d95297541bbc3da5f6df4053271249db85a12751a35",
	},
	{
		newDate(2022, time.July, 20),
		"f27b1e86b5e38452ca5fb2afa339d58e3f3ad868439fb530280b19155e5db97e",
	},
	{
		newDate(2022, time.July, 27),
		"cbe00a365c296b6e4e970ab456a24902deb04e58d83274e725ee12a867f632cc",
	},
	{
		newDate(2022, time.August, 3),
		"e5d24eed4a1dc4f1a402dd418b01c82b4e4191cbfe4175caf4d0d60d16e29287",
	},
	{
		newDate(2022, time.August, 10),
		"1e1388869b3353418a47479bca467d5dfa536e3ceb2ee1d4d35c81db1247c513",
	},
	{
		newDate(2022, time.August, 17),
		"6a08f3a94f39bcee4dceced42ebad476cae08670d259842afb0b362a75c2d57b",
	},
	{
		newDate(2022, time.August, 24),
		"a62e18063fe03e4b155256703c850288dd83b631787f944f241de696d8f77d7f",
	},
	{
		newDate(2022, time.August, 31),
		"1fa3deaad2140ed5f0a6c7803061a8108b86faaa17176c2950bd61533c0b362b",
	},
	{
		newDate(2022, time.September, 7),
		"ddbfd2881dc2d868cda2a9915150a3b557879ad941b03e87c86e5af59a63045f",
	},
	{
		newDate(2022, time.September, 14),
		"4b7607da3004caafce8aa6934794d42de37dd9c9483110a78f6549f8490bf850",
	},
	{
		newDate(2022, time.September, 21),
		"3cd1c1057c3540a43d0bf953c618aa7719fc281ccf0c67be0c350cdbca19b960",
	},
	{
		newDate(2022, time.September, 28),
		"84314980f7f9ae4c5575113929b7ad5e43629b0b50aef9f973fbcfb11b6c6687",
	},
	{
		newDate(2022, time.October, 5),
		"6d3a9b2805ad038dafff23bf89dc8eeac94179f82d02eb7fd5efdc964ee4c7e6",
	},
	{
		newDate(2022, time.October, 12),
		"e8fe14c75f3c2edc7bfb95a7b0f83dff6e2c6135b3859ea01389797a61a72b52",
	},
	{
		newDate(2022, time.October, 19),
		"1ee173613f2ede3e9ebc451ee8d3023c793262483c1eb7dd6db69afa2bcc5c94",
	},
	{
		newDate(2022, time.October, 26),
		"343352434d18800b70985a585278fff9320569e4fa17c5416434986fc0d70ded",
	},
	{
		newDate(2022, time.November, 2),
		"3c04497d0a9632b6a180f5bdcb23552482001d87d8a56e97c195fcb0949a16cd",
	},
	{
		newDate(2022, time.November, 9),
		"2cf69ad8f735e98c5f2db62243d61538771616068b9a15cf2b53189a023485a8",
	},
	{
		newDate(2022, time.November, 16),
		"02f7ea39c662d4a6de0eb6c5a966f40d18e0a8cc7819faa62084b00bd9a1c14b",
	},
	{
		newDate(2022, time.November, 23),
		"9a81fbea1fe22e0908a116f514be7d65045f4eaab28bee2e98eb41a749e11d4a",
	},
	{
		newDate(2022, time.November, 30),
		"aa890ff09415b12005a0c233d09282abec26aaa8c42990259f3b4fc30d50f0d2",
	},
	{
		newDate(2022, time.December, 7),
		"ce338608fcbaeac42c2d394ba75d8544d4e82bfaff5c1972935c46fa23e08d52",
	},
	{
		newDate(2022, time.December, 14),
		"8724b273af8edf541ef886ec9a82a89ae48f9eb68f7ac6b82e52b8764138b912",
	},
	{
		newDate(2022, time.December, 21),
		"13398360d9064ca13b07fd3f737637e9eea17e2c8f720b6d7044eaee835dddaa",
	}
}

func newDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func filterEpochsByYear(epochs []Epoch, year int) []Epoch {
	if year == 0 {
		return epochs
	}

	results := make([]Epoch, 0)

	for _, epoch := range epochs {
		if epoch.Date.Year() == year {
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
) []uint64 {
	tx, err := getTransactionResult(ctx, c, txID)
	if err != nil {
		panic(err)
	}

	return getDelegationRewards(tx.Events, delegationRecords)
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

type FlowscanTransactionResponse struct {
	Account struct {
		QueryResult struct {
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

func (r FlowscanTransactionResponse) DelegationRecords() []DelegationRecord {
	records := make([]DelegationRecord, 0)

	for _, tx := range r.Account.QueryResult.Edges {
		timestamp, err := time.Parse(time.RFC3339, tx.Node.Time)
		if err != nil {
			panic(err)
		}

		for _, event := range tx.Node.Events.Edges {
			fields := event.Node.Fields

			delegatorID, err := strconv.ParseUint(fields[1].Value, 10, 32)
			if err != nil {
				panic(err)
			}

			record := DelegationRecord{
				NodeID:      fields[0].Value,
				DelegatorID: uint32(delegatorID),
				Time:        timestamp,
			}

			records = append(records, record)
		}
	}

	return records
}

func (r FlowscanTransactionResponse) GetPageInfo() FlowscanPageInfo {
	return r.Account.QueryResult.PageInfo
}

func getDelegationRecords(ctx context.Context, address flow.Address) []DelegationRecord {
	client := graphql.NewClient(flowscanAPI)

	records := make([]DelegationRecord, 0)

	pageInfo := FlowscanPageInfo{
		HasNextPage: true,
	}

	for pageInfo.HasNextPage {
		var newRecords []DelegationRecord
		newRecords, pageInfo = getDelegationRecordsPage(ctx, client, address, pageInfo.EndCursor)

		records = append(records, newRecords...)
	}

	return records
}

func getDelegationRecordsPage(
	ctx context.Context,
	client *graphql.Client,
	address flow.Address,
	afterCursor string,
) ([]DelegationRecord, FlowscanPageInfo) {

	req := graphql.NewRequest(`
		query AccountTransactionsQuery($address: ID!, $role: TransactionRole, $first: Int!, $after: ID) {
			account(id: $address) {
				queryResult: transactions(first: $first, after: $after, role: $role) {
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

	var response FlowscanTransactionResponse

	err := client.Run(ctx, req, &response)
	if err != nil {
		panic(err)
	}

	return response.DelegationRecords(), response.GetPageInfo()
}
