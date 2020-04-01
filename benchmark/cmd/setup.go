package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/relab/bbchain-dapp/src/core/course"
	"github.com/relab/bbchain-dapp/src/core/course/contract"
	"github.com/spf13/cobra"

	pb "github.com/relab/bbchain-dapp/benchmark/proto"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup benchmark",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := setupClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = setupDB(dbPath, dbFile)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		clientConn.Close()
		db.Close()
	},
}

var deployCoursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "Deploy N courses contract",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln(err.Error())
		}
		e, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalln(err.Error())
		}
		s, err := strconv.Atoi(args[2])
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = createCourses(c, e, s)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func createCourses(c, e, s int) error {
	as := database.NewAccountStore(db, []string{"eth_accounts"})
	for i := 0; i < c; i++ {
		evaluators, err := as.GetAndSelect(as.Path, e, pb.Type_EVALUATOR)
		sender := evaluators[0]
		cAddr, err := deployCourse(sender.GetHexKey(), database.HexAddresses(evaluators), int64(len(evaluators)))
		if err != nil {
			return err
		}

		cs, err := database.CreateCourseStore(db, cAddr.Hex())
		if err != nil {
			return err
		}

		a := database.NewAccountStore(db, cs.Path)
		err = a.AddAccounts(append(cs.Path, "evaluators"), evaluators)
		if err != nil {
			return err
		}

		course, err := getCourse(cAddr)
		if err != nil {
			return err
		}

		students, err := as.GetAndSelect(as.Path, s, pb.Type_STUDENT)
		err = addStudents(sender.GetHexKey(), course, database.HexAddresses(students))
		if err != nil {
			return err
		}

		err = a.AddAccounts(append(cs.Path, "students"), students)
		if err != nil {
			return err
		}
	}
	return nil
}

func deployCourse(senderHexKey string, ownersList []string, quorum int64) (common.Address, error) {
	owners := database.HexSliceToAddress(ownersList)
	backend, _ := clientConn.Backend()
	opts, err := GetTxOpts(senderHexKey)
	if err != nil {
		return common.Address{}, err
	}

	now := time.Now()
	startingTime := now.Unix()
	endingTime := now.Add(time.Hour).Unix()

	cAddr, _, _, err := contract.DeployCourse(opts, backend, owners, big.NewInt(quorum), big.NewInt(startingTime), big.NewInt(endingTime))
	if err != nil || cAddr.Hex() == "0x0000000000000000000000000000000000000000" {
		return common.Address{}, fmt.Errorf("failed to deploy the contract: %v", err)
	}
	fmt.Printf("Contract %x successfully deployed\n", cAddr)
	return cAddr, nil
}

func getCourse(courseAddress common.Address) (*course.Course, error) {
	backend, _ := clientConn.Backend()
	course, err := course.NewCourse(courseAddress, backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course: %v", err)
	}
	return course, nil
}

func addStudents(senderHexKey string, course *course.Course, addresses []string) error {
	students := database.HexSliceToAddress(addresses)
	opts, err := GetTxOpts(senderHexKey)
	if err != nil {
		return err
	}
	// TODO: modify contract to add a list of addresses
	for _, student := range students {
		_, err := course.AddStudent(opts, student)
		if err != nil {
			return fmt.Errorf("Failed to add student: %v", err)
		}
	}
	return nil
}

func getBalance(hexAddress string) *big.Float {
	backend, _ := clientConn.Backend()
	address := common.HexToAddress(hexAddress)
	balance, err := backend.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	// eth = wei / 10^18
	return new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(deployCoursesCmd)
}
