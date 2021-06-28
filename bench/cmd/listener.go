package cmd

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/relab/credbench/pkg/course"
	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to blockchain events",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}

		wsbackend, err := ethclient.Dial("ws://127.0.0.1:8546")
		if err != nil {
			log.Fatal(err)
		}

		listenCourse(wsbackend, c)
	},
}

func listen(backend *ethclient.Client) {
	headers := make(chan *types.Header)
	sub, err := backend.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := backend.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}
			log.Infoln("BLOCK: ", block.Hash().Hex())
			log.Infoln("NUMBER: ", block.Number().Uint64())
			log.Infoln("TIME: ", block.Time())
			log.Infoln("DIFFICULTY: ", block.Difficulty().Uint64())
			log.Infoln("NUMBER OF TXS: ", len(block.Transactions()))
			for i, t := range block.Transactions() {
				log.Infof("tx-%d: %x \n", i, t.Hash())
			}
			log.Infoln("------------------------------")
		}
	}
}

// TODO: pass query
func listenCourse(backend *ethclient.Client, course *course.Course) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{course.Address()},
	}

	events := make(chan types.Log)
	sub, err := backend.SubscribeFilterLogs(context.Background(), query, events)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case e := <-events:
			log.Infof("%+v\n", e)
		}
	}
}
