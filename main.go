package main

import (
	"fmt"
	"strings"
	"log"
	"flag"
	"os"
	"wattpad-to-ebook/ebook"
	"wattpad-to-ebook/wattpad_stories"
	"github.com/yosssi/gohtml"
)


func download_wattpad(url string) error {
	chapters, metadata, err := wattpadstories.Get_Chapters(url)
	
	// fmt.Println(metadata)
	
	epubName := fmt.Sprintf("%s - %s.epub", metadata.Name, metadata.Author)
	// fmt.Println(epubName)
	if err != nil {
		return err
	}

	tempDir, err := ebook.Setup_temp()

	if err != nil {
		return err
	}

	err = ebook.Setup_container(tempDir)

	if err != nil {
		return err
	}

	err = ebook.Setup_content(tempDir, len(chapters), metadata.Name, metadata.Author, metadata.Description, metadata.CoverImageType)

	if err != nil {
		return err
	}

	err = ebook.Setup_CSS(tempDir)

	if err != nil {
		return err
	}

	var nav_chapters []ebook.ChapterNavItem

	for i, chap := range chapters {
		nav_chapters = append(nav_chapters, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}

	err = ebook.Setup_Nav(tempDir, nav_chapters, metadata.Name)
	
	if err != nil {
		return err
	}
	
	for _, chapter := range chapters {
		bodyBytes, err := wattpadstories.Get_Chapter_Text(chapter.URL)
		
		if err != nil {
		return err
		}
		pretty := gohtml.Format(string(bodyBytes))
		err = ebook.AddChapters(pretty, chapter.Index, tempDir, chapter.Title)

		if err != nil {
			return err
		}

	}
	
	var chap_list = []ebook.ChapterNavItem{}

	for i, chap := range chapters {
		chap_list = append(chap_list, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}

	err = ebook.SetupToc(tempDir, metadata.Name, chap_list)
	
	if err != nil {
		return err
	}


	ebook.Make_Ebook(tempDir, epubName, metadata.CoverImage, metadata.CoverImageType)
	
	os.RemoveAll(tempDir)	
	
	return nil
}




func main(){
	url := flag.String("u", "", "URL of the story (required)")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "Error: -u is required.")
		flag.Usage()
		os.Exit(1)
	}

	// Proceed normally
	if strings.Contains(*url, "wattpad.com/story"){
		fmt.Println("Generating EPUB for:", *url)
		err := download_wattpad(*url)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Epub Generated Successfully")
	} else {
		log.Fatalf("A url '%s' não é válida do wattpad", *url)
	}
}