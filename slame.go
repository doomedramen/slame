package main

import (
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"os"
	"path"
	"github.com/fatih/color"
	"fmt"
	"strings"
	"os/exec"
	"bytes"
	"strconv"
	"errors"
)

const (
	MartinName = "Martin Page"
	MartinEmail = "martin.page@tsl.ac.uk"

	ShyamName = "Ghanasyam Rallapalli"
	ShyamEmail = "ghanasyam.rallapalli@tsl.ac.uk"

	SlameFileName = ".slame.db"

	AppName = "Slame"
	AppUsage = "A tools to make SLURM less of a nightmare"
	AppVersion = "0.1.0"

	CommandPartitionName = "partition"
	CommandPartitionAlias = "p"
	CommandPartitionUsage = "Get and set the partition to run on"

	CommandMemoryName = "memory"
	CommandMemoryAlias = "m"
	CommandMemoryUsage = "Get and set the memory allocation"

	CommandRunName = "run"
	CommandRunAlias = "r"
	CommandRunUsage = "Run a command via slurm"

	ParamParitionName = "partition"
	ParamPartitionValue = "tsl-short"
	ParamPartitionUsage = "partion to run job on. Overwrites global partition selection"

	ParamMemoryName = "memory"
	ParamMemoryValue = "1000"
	ParamMemoryUsage = "Memory to use for job. Overwrites global memory selection"

	SetPartitionMessage = "Partition set to:"
	GetPartitionMessage = "Current partition:"

	SetMemoryMessage = "Memory allocation set to:"
	GetMemoryMessage = "Current memory allocation:"

	MB = "mb"
	GB = "gb"
	TB = "tb"

	Error1 = "We could not detect your username"
	Error2 = "You have set your memory requirement"
	Error3 = "You have not set your partion requirement"
	Error4 = "No arguments received after command"
	Error5 = "Count not parse the value given"
)

var (
	db *bolt.DB
	bucketName = []byte("slame")
)

func main() {

	pathToMe := os.Getenv("HOME")

	var e error

	dbPath := path.Join(pathToMe, SlameFileName);
	db, e = bolt.Open(dbPath, 0600, nil)
	check(e);

	InitBucket();

	//defer db.Close()

	app := cli.NewApp()
	app.Name = AppName
	app.Usage = AppUsage
	app.Version = AppVersion
	app.Authors = []cli.Author{
		cli.Author{Name: MartinName, Email: MartinEmail},
		cli.Author{Name: ShyamName, Email:ShyamEmail},
	}

	app.Commands = []cli.Command{
		{
			Name:CommandPartitionName,
			Aliases:[]string{CommandPartitionAlias},
			Usage: CommandPartitionUsage,
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					SetPartition(c.Args().First())
					PrintSuccess(SetPartitionMessage, GetPartition())
				} else {
					PrintSuccess(GetPartitionMessage, GetPartition())
				}
			},
		},
		{
			Name:      CommandMemoryName,
			Aliases:     []string{CommandMemoryAlias},
			Usage:     CommandMemoryUsage,
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					SetMemory(c.Args().First())
					PrintSuccess(SetMemoryMessage, GetMemory())
				} else {
					PrintSuccess(GetMemoryMessage, GetMemory())
				}
			},
		},
		{
			Name:      CommandRunName,
			Aliases:     []string{CommandRunAlias},
			Usage:     CommandRunUsage,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: ParamParitionName,
					Value: ParamPartitionValue,
					Usage: ParamPartitionUsage,
				},
				cli.StringFlag{
					Name: ParamMemoryName,
					Value: ParamMemoryValue,
					Usage: ParamMemoryUsage,
				},
			},
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					Run(c.Args());
				} else {
					PrintError(Error4)
					cli.ShowAppHelp(c)
				}
			},
		},
	}

	app.Run(os.Args)
}

func MemoryConv(amount string) (string, error) {
	_, err := strconv.Atoi(amount);
	if (err == nil) {
		return amount, nil
	} else {
		amountLC := strings.ToLower(amount)

		if (strings.HasSuffix(amountLC, MB)) {
			withoutSuffix := strings.Split(amountLC, MB)[0]
			number, err := strconv.Atoi(withoutSuffix);
			check(err);
			inMB := number
			return strconv.Itoa(inMB), nil

		} else if (strings.HasSuffix(amountLC, GB)) {
			withoutSuffix := strings.Split(amountLC, GB)[0]
			number, err := strconv.Atoi(withoutSuffix);
			check(err);
			inMB := number * 1024
			return strconv.Itoa(inMB), nil
		} else if (strings.HasSuffix(amountLC, TB)) {
			withoutSuffix := strings.Split(amountLC, TB)[0]
			number, err := strconv.Atoi(withoutSuffix);
			check(err);
			inMB := number * 1024 * 1024
			return strconv.Itoa(inMB), nil
		} else {
			return "", errors.New(Error5)
		}
	}
}

func Get(key string) (string, error) {
	var p []byte;
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		p = b.Get([]byte(key))
		return nil
	})
	return string(p), err;
}
func Put(key string, value string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(key), []byte(value))
	})
}

func InitBucket() {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName);
		check(err);
		return err
	})
	check(err)
}

func SetPartition(p string) {
	err := Put(CommandPartitionName, p);
	check(err)
}
func GetPartition() string {
	out, err := Get(CommandPartitionName);
	check(err)
	return out;
}

func SetMemory(m string) {
	mb, err := MemoryConv(m)
	check(err)
	err = Put(CommandMemoryName, mb);
	check(err)
}
func GetMemory() string {
	out, err := Get(CommandMemoryName);
	check(err)
	return out;
}

func Run(args []string) {

	argString := strings.Join(args, " ")

	memory := GetMemory();
	partition := GetPartition();
	username := os.Getenv("USER");

	if (username == "") {
		PrintError(Error1)
	} else if (memory == "") {
		PrintError(Error2)
	} else if (partition == "") {
		PrintError(Error3)
	} else {

		cmd := SBatch(partition, memory, username, argString);

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
		fmt.Println(out.String())
	}
}

func SBatch(partition string, memory string, username string, argString string) *exec.Cmd {
	sbatch := "sbatch";
	ug := "-vvv -p " + partition + " --mem=" + memory + " -n 1 --mail-type=END,FAIL --mail-user=" + username + "@nbi.ac.uk --wrap=\"" + argString + "\""

	PrintSuccess("going to run","sbatch", ug)
	strings.Split(ug, " ")
	return exec.Command(sbatch, ug)
}

func PrintError(s ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintln(os.Stderr, s...)
	color.Unset()
}

func PrintSuccess(s ...interface{}) {
	color.Set(color.FgGreen)
	fmt.Fprintln(os.Stdout, s...)
	color.Unset()
}

func check(e error) {
	if e != nil {
		PrintError(e)
		os.Exit(1)
	}
}
