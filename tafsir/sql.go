package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func dump(file os.FileInfo) (SQLFile string, err error) {

	err = os.Setenv("MDB_JET3_CHARSET", "cp1256")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("mdb-tables", "bok/"+file.Name())
	output, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	//r := bufio.NewReader(output)
	r, err := ioutil.ReadAll(output)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	output.Close()

	l := string(r)
	tables := strings.Fields(l)

	var f *os.File
	defer f.Close()

	fn := strings.Split(file.Name(), ".")[0]

	err = os.MkdirAll("sql", 0755)
	if err != nil {
		return "", err
	}

	f, err = os.Create("sql/" + fn + ".sql")
	if err != nil {
		return "", err
	}

	if err = dumpSchema(f, file); err != nil {
		return "", err
	}

	f.Write([]byte("BEGIN;"))

	for _, table := range tables {

		export := exec.Command("mdb-export", "-I", "mysql", "bok/"+file.Name(), table)

		//export.Stdout = os.Stdout
		//export.Stderr = os.Stderr

		out, err := export.StdoutPipe()
		if err != nil {
			return "", err
		}

		if err = export.Start(); err != nil {
			return "", err
		}

		//rd := bufio.NewReader(out)

		r, err := ioutil.ReadAll(out)
		if err != nil {
			return "", err
		}

		if err = export.Wait(); err != nil {
			return "", err
		}

		_, err = f.Write(r)
		if err != nil {
			return "", err
		}

		//_, err = rd.WriteTo(f)
		//if err != nil {
		//	return "", err
		//}

		out.Close()

	}

	f.Write([]byte("COMMIT;\n"))
	return f.Name(), nil
}

func dumpSchema(f *os.File, bok os.FileInfo) error {

	cmd := exec.Command("mdb-schema", filepath.Join("bok", bok.Name()), "mysql")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	r, err := ioutil.ReadAll(stdout)
	if err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	_, err = f.Write(r)
	if err != nil {
		return err
	}

	defer stdout.Close()

	return nil
}
