package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

// This function is used to get a directory path that will store all the stats of the user's git code
func GetDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gogitlocalstats"
	return dotFile
}

// This function opens a file of the given filename. If the file does not exist, it creates the file
func OpenFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	return f
}

func ParseFileLinesToSlice(filePath string) []string {
	f := OpenFile(filePath)
	defer f.Close()
	var lines []string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return lines
}

// This functions return true if an array contains the value you need
func SliceContins(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func JoinSlices(new []string, existing []string) []string {
	for _, n := range new {
		if !SliceContins(existing, n) {
			existing = append(existing, n)
		}
	}
	return existing
}

func DumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	ioutil.WriteFile(filePath, []byte(content), 0755)
}

func AddNewSLiceElementsToFile(filePath string, newRepos []string) {
	existingRepos := ParseFileLinesToSlice(filePath)
	repos := JoinSlices(newRepos, existingRepos)
	DumpStringsSliceToFile(repos, filePath)
}

func RecursiveScanFolder(folder string) []string{
	return ScanGitFolders(make([]string,0),folder)
}

func Scan (folder string){
	fmt.Printf("Found Folders\n")
	repositories:=RecursiveScanFolder(folder)
	filePath:= GetDotFilePath()
	AddNewSLiceElementsToFile(filePath, repositories)
	fmt.Printf("\n\nSuccessfully Added \n\n")
}

func ScanGitFolders (folders []string, folder string)[]string{
	//Trim the last "/"
	folder = strings.TrimSuffix(folder,"/")

	f,err:= os.Open(folder)

	if err!=nil{
		log.Fatal(err)
	}
	files,err := f.Readdir(-1)
	f.Close()
	if err != nil{
		log.Fatal(err)
	}

	var path string

	for _,file:= range files{
		if file.IsDir(){
			path = folder + "/" + file.Name()
			if file.Name() == ".git"{
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules"{
				continue
			}
			folders = ScanGitFolders(folders,path)
		}
	}
	return folders
}