package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/elbuo8/gost"
	"github.com/russross/blackfriday"
)

func main() {
	var filepath = flag.String("f", "", "Specify relative filepath")
	var description = flag.String("d", "Gost Generated Gist", "Gist File Description")
	var public = flag.Bool("p", false, "Visibility of Gist")
	var filename = flag.String("n", "GostGenerated", "Name of Gist File")
	var outputFile = flag.String("o", "output.html", "Output filename")
	flag.Parse()
	if *filepath == "" {
		flag.PrintDefaults()
		log.Fatal("Filepath missing")
	}
	file, err := os.Open(*filepath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	gist := gost.New(os.Getenv("GOST_TOKEN"))
	savedGist := &gost.Gist{
		Description: *description,
		Public:      *public,
		Filename:    *filename,
		Files:       make(map[string]gost.GistFile),
	}

	var gistId string
	var transformedGists bytes.Buffer
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		currentLn := scanner.Text()

		if strings.HasPrefix(currentLn, "```") {

			extension := currentLn[3:]
			var currentGist bytes.Buffer

			for scanner.Scan() {

				currentGistLn := scanner.Text()

				if strings.HasPrefix(currentGistLn, "```") {

					if gistId == "" && len(savedGist.Files) == 0 { // Post
						gistFile := &gost.GistFile{
							Content:  currentGist.String(),
							Filename: "0." + extension,
						}
						savedGist.Files[gistFile.Filename] = *gistFile
						resp, err := gist.Create(savedGist.Description, savedGist.Public, gistFile)
						if err != nil {
							log.Fatal(err)
						}
						var gistResp map[string]interface{}
						err = json.Unmarshal(resp, &gistResp)
						if err != nil {
							log.Fatal(err)
						}
						gistId = gistResp["id"].(string)
						transformedGists.WriteString("[gist id=" + gistId + " file=" + gistFile.Filename + "]")

					} else {
						gistFile := &gost.GistFile{
							Content:  currentGist.String(),
							Filename: strconv.Itoa(len(savedGist.Files)) + "." + extension,
						}
						savedGist.Files[gistFile.Filename] = *gistFile
						_, err := gist.Edit(gistId, savedGist)
						if err != nil {
							log.Fatal(gistId)
						}
						transformedGists.WriteString("[gist id=" + gistId + " file=" + gistFile.Filename + "]")
					}

					break // Done with Current Gist, Keep Parsing regular

				} else {
					currentGist.WriteString(currentGistLn + "\n")
				}
			}
		} else {
			transformedGists.WriteString(currentLn + "\n")
		}
	}

	renderer := blackfriday.HtmlRenderer(0, "", "")
	HTMLRender := blackfriday.Markdown(transformedGists.Bytes(), renderer, 0)
	newFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatal(err)
	}
	_, err = newFile.Write(HTMLRender)
	if err != nil {
		log.Fatal(err)
	}
}
