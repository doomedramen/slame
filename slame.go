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
)

var (
	db *bolt.DB
	bucketName = []byte("slame")
)

func main() {

	pathToMe := os.Getenv("HOME")

	var e error

	dbPath := path.Join(pathToMe, ".slame.db");
	db, e = bolt.Open(dbPath, 0600, nil)
	check(e);

	InitBucket();

	//defer db.Close()

	app := cli.NewApp()
	app.Name = "slame"
	app.Usage = "slurm util"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:      "partition",
			Aliases:     []string{"p"},
			Usage:     "set/get partition",
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					SetPartition(c.Args().First())
					PrintSuccess("partition set to:", GetPartition())
				} else {
					PrintSuccess("Current partition:", GetPartition())
				}
			},
		},
		{
			Name:      "memory",
			Aliases:     []string{"m"},
			Usage:     "set/get memory allocation in mb",
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					SetMemory(c.Args().First())
					PrintSuccess("Memory allocation set to:", GetMemory())
				} else {
					PrintSuccess("Current memory allocation:", GetMemory())
				}
			},
		},
		{
			Name:      "run",
			Aliases:     []string{"m"},
			Usage:     "run command",
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					Run(c.Args());
				} else {

				}
			},
		},
	}

	app.Run(os.Args)
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
	err := Put("partition", p);
	check(err)
}
func GetPartition() string {
	out, err := Get("partition");
	check(err)
	return out;
}

func SetMemory(m string) {
	err := Put("memory", m);
	check(err)
}
func GetMemory() string {
	out, err := Get("memory");
	check(err)
	return out;
}

func Run(args []string) {

	argString := strings.Join(args, " ")

	memory := GetMemory();
	partition := GetPartition();
	username := os.Getenv("USER");

	if (username == "") {
		PrintError("we could not detect your username //TODO")
	} else if (memory == "") {
		PrintError("you have set your memory requirement")
	} else if (partition == "") {
		PrintError("you have not set your partion requirement")
	} else {

		parsedMem, err := strconv.Atoi(memory)
		check(err)

		//%x[sbatch #{partition} #{memory} -n 1 --mail-type=END,FAIL --mail-user=${USER}@nbi.ac.uk --wrap="#{cmd}"]

		batch := fmt.Sprintf("sbatch -vvv --partition=%s --mem=%d -n 1 --mail-type=END,FAIL --mail-user=%s@nbi.ac.uk --wrap=%q", partition, parsedMem, username, argString);
		parts := strings.Fields(batch);
		head := parts[0];
		parts = parts[1:len(parts)];

		cmd := exec.Command(head, parts...)
		fmt.Println("args", cmd.Args)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
		fmt.Println("Result: " + out.String())
	}
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
		//panic(e);
	}
}
