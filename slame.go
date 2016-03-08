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
//"strconv"
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
	app.Authors = []cli.Author{cli.Author{Name: "Martin Page", Email: "martin.page@tsl.ac.uk"}, cli.Author{Name: "Ghanasyam.Rallapalli", Email:"ghanasyam.rallapalli@tsl.ac.uk"}}

	app.Commands = []cli.Command{
		{
			Name:      "partition",
			Aliases:     []string{"p"},
			Usage:     "get and set the partition to run on",
			Action: func(c *cli.Context) {
				if (len(c.Args()) > 0) {
					SetPartition(c.Args().First())
					PrintSuccess("Partition set to:", GetPartition())
				} else {
					PrintSuccess("Current partition:", GetPartition())
				}
			},
		},
		{
			Name:      "memory",
			Aliases:     []string{"m"},
			Usage:     "get and set the memory allocation",
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
			Aliases:     []string{"r"},
			Usage:     "run command",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "partition",
					Value: "tsl-short",
					Usage: "partion to run job on. Overwrites global partition selection",
				},
				cli.StringFlag{
					Name: "memory",
					Value: "1000",
					Usage: "memory to use for job. Overwrites global memory selection",
				},
			},
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

		head := "sbatch";
		cmd := exec.Command(head, "-vvv", "-p", partition, "--mem=" + memory, "-n", "1", "--mail-type=END,FAIL", "--mail-user=" + username + "@nbi.ac.uk", "--wrap=\"" + argString + "\"")
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
