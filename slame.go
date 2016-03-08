package main

import (
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"os"
	"path"
	"os/user"
	"github.com/fatih/color"
	"fmt"
	"strings"
	"os/exec"
)

var (
	pathToMe string
	db *bolt.DB
	bucketName = []byte("slame")
)

func main() {

	usr, err := user.Current()
	check(err)

	pathToMe = usr.HomeDir;

	dbPath := path.Join(pathToMe, ".slame.db");
	db, err = bolt.Open(dbPath, 0600, nil)

	check(err);
	defer db.Close()

	InitBucket();

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
				println("RUN!!!");
				if (len(c.Args()) > 0) {
					Run(c.Args());
				} else {

				}
			},
		},
	}

	app.Run(os.Args)
}

func InitBucket() {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName);
		check(err);
		return err
	})
}

func SetPartition(p string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte("partition"), []byte(p))
	})
	check(err);
}
func GetPartition() string {
	var p []byte;
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		p = b.Get([]byte("partition"))
		return nil
	})
	check(err)
	return string(p);
}

func SetMemory(m string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		b.Put([]byte("memory"), []byte(m))
		return nil
	})
	check(err);
}
func GetMemory() string {
	var m []byte;
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		m = b.Get([]byte("memory"))
		return nil
	})
	check(err);
	return string(m);
}

func Run(args []string) {

	argString := strings.Join(args, " ")

	println("RUN COMMAND:", argString)

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

		batch := fmt.Sprintf("sbatch %s %s -n 1 --mail-type=END,FAIL --mail-user=%s@nbi.ac.uk --wrap=\"%s\"", partition, memory, username, argString);
		println("want to run", batch);
		parts := strings.Fields(batch);
		head := parts[0];
		parts = parts[1:len(parts)];

		out, err := exec.Command(head, parts...).Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
		fmt.Printf("%s", out)
		//wg.Done() // Need to signal to waitgroup that this goroutine is done
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
		//panic(e);
	}
}
