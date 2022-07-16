package shelve

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JordWyatt/shelve/pkg/shelve"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "shelve",
	Short: "shelve - a simple CLI to managed music on my NAS",
	Long:  `shelve is used to import music into an 'Organised' folder and an 'Unorganised' folder.`,
	Run:   importDirectories,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops, an error occured while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func importDirectories(cmd *cobra.Command, args []string) {
	directories := getDirectoriesInStagingDirectory()
	directoryNames := []string{}
	directoryNamesToImport := []string{}

	for _, directory := range directories {
		directoryNames = append(directoryNames, directory.Name)
	}

	prompt := &survey.MultiSelect{
		Message: "Select directories to import:",
		Options: directoryNames,
	}

	survey.AskOne(prompt, &directoryNamesToImport)

	directoriesNamesToImportMap := make(map[string]struct{})

	for _, directoryName := range directoryNamesToImport {
		directoriesNamesToImportMap[directoryName] = struct{}{}
	}

	for _, directory := range directories {
		if _, ok := directoriesNamesToImportMap[directory.Name]; ok {
			copyDirectoryToTargetDirectory(directory)
		}
	}

	triggerBeetsImport()

	fmt.Println("Done!")
}

func getDirectoriesInStagingDirectory() []shelve.Directory {
	files, err := ioutil.ReadDir(shelve.STAGING_DIRECTORY)

	if err != nil {
		log.Fatalf("Error reading from %s: %s", shelve.STAGING_DIRECTORY, err)
	}

	var directories = []shelve.Directory{}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, file := range files {
		if file.IsDir() {
			directory := shelve.Directory{
				Name: file.Name(),
				Path: filepath.Join(shelve.STAGING_DIRECTORY, file.Name()),
			}
			directories = append(directories, directory)
		}
	}

	return directories
}

func copyDirectoryToTargetDirectory(directory shelve.Directory) {

	fmt.Printf("Copying %s to %s \n", directory.Path, shelve.TARGET_DIRECTORY)

	app := "cp"
	flags := "-al"
	cmd := exec.Command(app, flags, directory.Path, shelve.TARGET_DIRECTORY)
	stdout, err := cmd.Output()

	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println(string(stdout))
}

func triggerBeetsImport() {
	fmt.Println("Running beets import in docker container...")

	app := "docker"
	args := []string{"exec", "-it", "beets", "/bin/bash", "-c", "beet import /downloads"}
	cmd := exec.Command(app, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf(err.Error())
	}
}
