package ebook

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
)


type NCX struct {
	XMLName xml.Name `xml:"ncx"`
	Xmlns   string   `xml:"xmlns,attr"`
	Version string   `xml:"version,attr"`
	Head    NCXHead  `xml:"head"`
	Title   NCXTitle `xml:"docTitle"`
	NavMap  NCXNavMap `xml:"navMap"`
}

type NCXHead struct {
	Metas []NCXMeta `xml:"meta"`
}

type NCXMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type NCXTitle struct {
	Text string `xml:"text"`
}

type NCXNavMap struct {
	NavPoints []NCXNavPoint `xml:"navPoint"`
}

type NCXNavPoint struct {
	ID        string       `xml:"id,attr"`
	PlayOrder int          `xml:"playOrder,attr"`
	NavLabel  NCXLabel     `xml:"navLabel"`
	Content   NCXContent   `xml:"content"`
}

type NCXLabel struct {
	Text string `xml:"text"`
}

type NCXContent struct {
	Src string `xml:"src,attr"`
}

func GenerateTOCNCX(title string, uid string, chapters []ChapterNavItem) ([]byte, error) {
	toc := NCX{
		Xmlns:   "http://www.daisy.org/z3986/2005/ncx/",
		Version: "2005-1",
		Head: NCXHead{
			Metas: []NCXMeta{
				{Name: "dtb:uid", Content: uid},
				{Name: "dtb:depth", Content: "1"},
				{Name: "dtb:totalPageCount", Content: "0"},
				{Name: "dtb:maxPageNumber", Content: "0"},
			},
		},
		Title: NCXTitle{Text: title},
	}

	for i, chap := range chapters {
		navPoint := NCXNavPoint{
			ID:        fmt.Sprintf("chapter_%d", i+1),
			PlayOrder: i + 1,
			NavLabel:  NCXLabel{Text: chap.Title},
			Content:   NCXContent{Src: chap.Href},
		}
		toc.NavMap.NavPoints = append(toc.NavMap.NavPoints, navPoint)
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version='1.0' encoding='utf-8'?>` + "\n")

	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(toc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


// 1. Defina as structs que correspondem à hierarquia XML
type Container struct {
  XMLName   xml.Name   `xml:"container"`
  XMLNS     string     `xml:"xmlns,attr"`
  Version   string     `xml:"version,attr"`
  Rootfiles Rootfiles  `xml:"rootfiles"`
}

type Rootfiles struct {
  Rootfile Rootfile `xml:"rootfile"`
}

type Rootfile struct {
  FullPath  string `xml:"full-path,attr"`
  MediaType string `xml:"media-type,attr"`
}

// 2. Função para gerar o XML completo com cabeçalho
func BuildContainerXML(fullPath, mediaType string) ([]byte, error) {
  c := Container{
    XMLNS:   "urn:oasis:names:tc:opendocument:xmlns:container",
    Version: "1.0",
    Rootfiles: Rootfiles{
      Rootfile: Rootfile{
        FullPath:  fullPath,
        MediaType: mediaType,
      },
    },
  }

  // Gerar XML legível
  out, err := xml.MarshalIndent(c, "", "  ")
  if err != nil {
    return nil, err
  }

  // Acrescentar declaração XML no início
  header := []byte(xml.Header)
  return append(header, out...), nil
}



type Package struct {
	XMLName          xml.Name  `xml:"package"`
	Xmlns            string    `xml:"xmlns,attr"`
	UniqueIdentifier string    `xml:"unique-identifier,attr"`
	Version          string    `xml:"version,attr"`
	Prefix           string    `xml:"prefix,attr"`
	Metadata         Metadata  `xml:"metadata"`
	Manifest         Manifest  `xml:"manifest"`
	Spine            Spine     `xml:"spine"`
}

type Metadata struct {
	XMLNSDC     string      `xml:"xmlns:dc,attr"`
	XMLNSOPF    string      `xml:"xmlns:opf,attr"`
	Modified    Meta        `xml:"meta"`
	// Generator   MetaSimple  `xml:"meta"`
	Identifier  Identifier  `xml:"dc:identifier"`
	Title       string      `xml:"dc:title"`
	Language    string      `xml:"dc:language"`
	Creator     Creator     `xml:"dc:creator"`
	Description string      `xml:"dc:description"`
}

type Meta struct {
	Property string `xml:"property,attr"`
	Content  string `xml:",chardata"`
}

type MetaSimple struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type Identifier struct {
	ID   string `xml:"id,attr"`
	Body string `xml:",chardata"`
}

type Creator struct {
	ID   string `xml:"id,attr"`
	Body string `xml:",chardata"`
}

type Manifest struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Href       string `xml:"href,attr"`
	ID         string `xml:"id,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr,omitempty"`
}

type Spine struct {
	Toc      string     `xml:"toc,attr"`
	Itemrefs []Itemref  `xml:"itemref"`
}

type Itemref struct {
	IDRef string `xml:"idref,attr"`
}

type ChapterNavItem struct {
	Href  string
	Title string
}

func GenerateNavXHTML(bookTitle string, chapters []ChapterNavItem) (string, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version='1.0' encoding='utf-8'`)
	// doc.CreateDocType("html", "", "", "")

	html := doc.CreateElement("html")
	html.CreateAttr("xmlns", "http://www.w3.org/1999/xhtml")
	html.CreateAttr("xmlns:epub", "http://www.idpf.org/2007/ops")
	html.CreateAttr("lang", "en")
	html.CreateAttr("xml:lang", "en")

	head := html.CreateElement("head")
	head.CreateElement("title").SetText(bookTitle)
	link := head.CreateElement("link")
	link.CreateAttr("href", "../style/nav.css")
	link.CreateAttr("rel", "stylesheet")
	link.CreateAttr("type", "text/css")

	body := html.CreateElement("body")
	nav := body.CreateElement("nav")
	nav.CreateAttr("epub:type", "toc")
	nav.CreateAttr("id", "id")
	nav.CreateAttr("role", "doc-toc")

	nav.CreateElement("h2").SetText(bookTitle)

	ol := nav.CreateElement("ol")
	for _, chap := range chapters {
		li := ol.CreateElement("li")
		a := li.CreateElement("a")
		a.CreateAttr("href", chap.Href)
		a.SetText(chap.Title)
	}

	doc.Indent(2)
	return doc.WriteToString()
}



func GenerateContentOPF(title, author, description string, chapterCount int, img_type string) ([]byte, error) {
	chapters := make([]Item, 0)
	var refs []Itemref

	// refs = append(refs, Itemref{IDRef: "style_nav"})

	for i := range chapterCount {
		refs = append(refs, 
		Itemref{IDRef: fmt.Sprintf("chapter_%d", i+1)})
	}
	// // Introduction
	// chapters = append(chapters, Item{Href: "introduction.xhtml", ID: "chapter_0", MediaType: "application/xhtml+xml"})
	// refs = append(refs, Itemref{IDRef: "chapter_0"})

	// Nav
	chapters = append(chapters, Item{Href: "nav.xhtml", ID: "nav", MediaType: "application/xhtml+xml", Properties: "nav"})
	// refs = append(refs, Itemref{IDRef: "nav"})

	img_ext := getImageExt(img_type)
	// Static entries
	staticItems := []Item{
		{Href: "../style/main.css", ID: "doc_style", MediaType: "text/css"},
		{Href: "../style/nav.css", ID: "style_nav", MediaType: "text/css"},
		{Href: fmt.Sprintf("../cover.%s", img_ext), ID: "cover", MediaType: img_type, Properties: "cover-image"},
		{Href: "toc.ncx", ID: "ncx", MediaType: "application/x-dtbncx+xml"},
	}

	for i := range chapterCount {
		staticItems = append(staticItems, 
			Item{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), ID: fmt.Sprintf("chapter_%d", i+1), MediaType: "application/xhtml+xml"},)
	}

	manifest := Manifest{Items: append(staticItems, chapters...)}

	pkg := Package{
		Xmlns:            "http://www.idpf.org/2007/opf",
		UniqueIdentifier: "id",
		Version:          "3.0",
		Prefix:           "rendition: http://www.idpf.org/vocab/rendition/#",
		Metadata: Metadata{
			XMLNSDC:     "http://purl.org/dc/elements/1.1/",
			XMLNSOPF:    "http://www.idpf.org/2007/opf",
			Modified:    Meta{Property: "dcterms:modified", Content: time.Now().UTC().Format(time.RFC3339)},
			// Generator:   MetaSimple{Name: "generator", Content: "YourGenerator 1.0"},
			Identifier:  Identifier{ID: "id", Body: "quikmcbu"},
			Title:       title,
			Language:    "en",
			Creator:     Creator{ID: "creator", Body: author},
			Description: description,
		},
		Manifest: manifest,
		Spine:    Spine{Toc: "ncx", Itemrefs: refs},
	}

	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)
	xmlEnc := xml.NewEncoder(buf)
	xmlEnc.Indent("", "  ")
	if err := xmlEnc.Encode(pkg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}




type Html struct {
  XMLName   xml.Name `xml:"html"`
  Xmlns     string   `xml:"xmlns,attr"`
  XmlnsEpub string   `xml:"xmlns:epub,attr"`
  EpubPrefix string  `xml:"epub:prefix,attr"`
  Lang      string   `xml:"lang,attr"`
  XmlLang   string   `xml:"xml:lang,attr"`
  Head      Head     `xml:"head"`
  Body      Body     `xml:"body"`
}

type Head struct {
  Title string `xml:"title"`
  Link  Link   `xml:"link"`
}

type Link struct {
  Href string `xml:"href,attr"`
  Rel  string `xml:"rel,attr"`
  Type string `xml:"type,attr"`
}

type Body struct {
  Content []byte `xml:",innerxml"`
}

// Extracts the inner HTML of the <body> tag
func getBodyNodeFromHTML(rawFormatted string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(rawFormatted))
	if err != nil {
		return nil, err
	}
	var ErrBodyNotFound = errors.New("no <body> tag found")
	var bodyNode *html.Node
	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyNode = n
			return
		}
		for c := n.FirstChild; c != nil && bodyNode == nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	if bodyNode == nil {
		return nil, ErrBodyNotFound
	}
	// fmt.Println(bodyNode)
	return bodyNode, nil
}



func GenerateXHTML(title string, bodyContent string) ([]byte, error) {
	// Step 1: Parse HTML5 body content

	doc, err := getBodyNodeFromHTML(bodyContent)
	if err != nil {
		return nil, err
	}

	// Step 2: Serialize valid XHTML
	bodyXHTML := xhtml.InnerHTML(doc)
	// var buffer bytes.Buffer
	// html.Render(&buffer, doc)

	// fmt.Println(buffer.String())
	// fmt.Println(doc)

	// Step 3: Build document
	docStruct := Html{
		Xmlns:      "http://www.w3.org/1999/xhtml",
		XmlnsEpub:  "http://www.idpf.org/2007/ops",
		EpubPrefix: "z3998: http://www.daisy.org/z3998/2012/vocab/structure/#",
		Lang:       "en",
		XmlLang:    "en",
		Head: Head{
			Title: title,
			Link: Link{
				Href: "../style/main.css",
				Rel:  "stylesheet",
				Type: "text/css",
			},
		},
		Body: Body{
			Content: []byte(bodyXHTML),
		},
	}

	// Step 4: Encode as XML
	buf := &bytes.Buffer{}
	buf.WriteString(`<?xml version='1.0' encoding='utf-8'?>` + "\n")
	buf.WriteString(`<!DOCTYPE html>` + "\n")

	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(docStruct); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Setup_temp()(string, error){
	tempDir, err := os.MkdirTemp("", "epub-story-")

	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func css_main() string{
	main := 
	`
	@namespace epub "http://www.idpf.org/2007/ops";

body {
    font-family: Verdana, Helvetica, Arial, sans-serif;
}

h1 {
    text-align: center;
}

h2 {
    text-align: left;
    font-weight: bold;
}

ol {
    list-style-type: none;
    margin: 0;
}

ol > li {
    margin-top: 0.3em;
}

ol > li > span {
    font-weight: bold;
}

ol > li > ol {
    margin-left: 0.5em;
}

.spoiler {
    padding-left: 0.4em;
    border-left: 0.2em solid #c7ccd1;
}
	`
	return main
}



func Setup_CSS(tempDir string) (error) {
	err := os.MkdirAll(filepath.Join(tempDir, "style"), os.ModePerm)
	if err != nil {
		return err
	}

	cssMain := filepath.Join(tempDir, "style", "main.css")
	navCSS := filepath.Join(tempDir, "style", "nav.css")

	file, err := os.Create(cssMain)

	if err != nil {
		return err
	}

	file.WriteString(css_main())
	file.Close()

	file, err = os.Create(navCSS)
	if err != nil {
		return err
	}

	file.WriteString("BODY {color: white;}")
	file.Close()
	return nil
}


func Setup_container(tempDir string) (error) {
	err := os.MkdirAll(filepath.Join(tempDir, "META-INF"), os.ModePerm)

	if err != nil {
		return err
	}
	containerPath := filepath.Join(tempDir, "META-INF", "container.xml")

	containerFile, err := os.Create(containerPath)

	if err != nil {
		return err
	}

	containerContent, err := BuildContainerXML("OEBPS/content.opf", "application/oebps-package+xml")
  	if err != nil {
    	return err
  	}
	containerFile.Write(containerContent)
	containerFile.Close()

	contentPath := filepath.Join(tempDir, "OEBPS")

	err = os.MkdirAll(contentPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func Setup_content(tempDir string, numb_of_chaps int, name string, author string, desc string, img_type string) (error){
	contentFilePath := filepath.Join(tempDir, "OEBPS", "content.opf")

	contentFile, err := os.Create(contentFilePath)
	
	if err != nil {
		return err
	}

	containerBytes, err := GenerateContentOPF(name, author, desc, numb_of_chaps, img_type)
  	if err != nil {
    	return err
  	}
	contentFile.Write(containerBytes)
	defer contentFile.Close()
	return nil
}

func Setup_Nav(tempDir string, chapters []ChapterNavItem, Title string) (error) {
	NavFilePath := filepath.Join(tempDir, "OEBPS", "nav.xhtml")
	NavFile, err := os.Create(NavFilePath)
	
	if err != nil {
		return err
	}

	NavString, err := GenerateNavXHTML(Title, chapters)

	if err != nil {
		return err
	}

	NavFile.WriteString(NavString)

	return nil
}


func SetupToc(tempDir string, Title string, chap_list []ChapterNavItem) error {
	TocFilePath, err := os.Create(filepath.Join(tempDir, "OEBPS", "toc.ncx"))

	if err != nil {
		return err
	}


	TocString, err := GenerateTOCNCX(Title, "quikmcbu", chap_list)

	if err != nil {
		return err
	}

	TocFilePath.Write(TocString)
	TocFilePath.Close()
	return nil
}


func createToc(w *zip.Writer, tempDir string) error {
	f, err := w.Create("OEBPS/toc.ncx")
    if err != nil {
        return err
    }
    content, err := os.ReadFile(filepath.Join(tempDir, "OEBPS", "toc.ncx"))

	if err != nil {
		return err
	}
    _, err = f.Write(content)
    if err != nil {
		return err
	}

	return nil
}

func AddChapters(chap_body string, chap_index int, tempDir string, title string) (error) {
	XHTMLPath := filepath.Join(tempDir, "OEBPS", fmt.Sprintf("chapter_%d.xhtml", chap_index))
	XHTMLFile, err := os.Create(XHTMLPath)

	if err != nil {
		return err
	}

	XHMTLBytes, err := GenerateXHTML(title, chap_body)

	if err != nil {
		return err
	}

	XHTMLFile.Write(XHMTLBytes)
	XHTMLFile.Close()

	return nil
}

func createMimetype(w *zip.Writer) error {
    f, err := w.CreateHeader(&zip.FileHeader{
        Name:   "mimetype",
        Method: zip.Store,
    })
    if err != nil {
        return err
    }
    _, err = f.Write([]byte("application/epub+zip"))
    return err
}

func createContainerXML(w *zip.Writer, tempFile string) error {
    f, err := w.Create("META-INF/container.xml")
    if err != nil {
        return err
    }
    content, err := os.ReadFile(tempFile)

	if err != nil {
		return err
	}
    _, err = f.Write(content)
    if err != nil {
		return err
	}
	return nil
}

func createContentOPF(w *zip.Writer, tempFile string) error {
    f, err := w.Create("OEBPS/content.opf")
    if err != nil {
        return err
    }
	
	content, err := os.ReadFile(tempFile)

	if err != nil {
		return err
	}

    _, err = f.Write(content)
	if err != nil {
		return err
	}

    return nil
}

func createMain(w *zip.Writer, tempDir string) error {
	f, err := w.Create("style/main.css")

	 if err != nil {
        return err
    }
	
	content, err := os.ReadFile(filepath.Join(tempDir, "style", "main.css"))

	if err != nil {
		return err
	}

    _, err = f.Write(content)
	if err != nil {
		return err
	}


	return nil
}


func getImageExt(mediaType string) string {
    switch mediaType {
    case "image/png":
        return "png"
    case "image/gif":
        return "gif"
    default:
        return "jpg" // fallback
    }
}

func AddCoverImage(zipWriter *zip.Writer, coverBytes []byte, mediaType string) error {
    f, err := zipWriter.Create("cover." + getImageExt(mediaType))
    if err != nil {
        return err
    }
    _, err = f.Write(coverBytes)
    return err
}

func createStyleNav(w *zip.Writer, tempDir string) error {
	f, err := w.Create("style/nav.css")

	 if err != nil {
        return err
    }
	
	content, err := os.ReadFile(filepath.Join(tempDir, "style", "nav.css"))

	if err != nil {
		return err
	}

    _, err = f.Write(content)
	if err != nil {
		return err
	}


	return nil
}

func createChapter(w *zip.Writer, tempDir string, tempFile string) error {
    f, err := w.Create(filepath.Join("OEBPS", tempFile))
    if err != nil {
        return err
    }
    content, err := os.ReadFile(filepath.Join(tempDir, "OEBPS", tempFile))
	if err != nil {
		return err
	}
    _, err = f.Write(content)
    
	if err != nil {
		return err
	}

	return nil
}

func createNav(w *zip.Writer, tempDir string) (error) {
	f, err := w.Create("OEBPS/nav.xhtml")
	
	if err != nil {
		return err
	}
	
	content, err := os.ReadFile(filepath.Join(tempDir, "OEBPS", "nav.xhtml"))
	if err != nil {
		return err
	}
    _, err = f.Write(content)
    
	if err != nil {
		return err
	}

	return nil
}


func Make_Ebook(tempDir string, filename string, img_bytes []byte, img_type string) (error) {
	// fmt.Println(filename)
	epub, err := os.Create(filename)
	
	if err != nil {
		return err
	}

	cont := filepath.Join(tempDir, "META-INF", "container.xml")
	content := filepath.Join(tempDir, "OEBPS", "content.opf")
	chapters, err := os.ReadDir(filepath.Join(tempDir, "OEBPS"), )

	if err != nil {
		return err
	}

    defer epub.Close()

    zipWriter := zip.NewWriter(epub)
    defer zipWriter.Close()

    // para criar os arquivos do EPUB
    if err := createMimetype(zipWriter); err != nil {
        return err
    }

	if err := createContainerXML(zipWriter, cont); err != nil {
		return err
	}
	

	if err := createContentOPF(zipWriter, content); err != nil {
		return err
	}
	for _, i := range chapters {
		
		if strings.HasSuffix(i.Name(), ".xhtml") && i.Name() != "nav.xhtml" {
			err := createChapter(zipWriter, tempDir, i.Name())
			if err != nil {
				return err
			}
		}
	}

	if err := createMain(zipWriter, tempDir); err != nil {
		return err
	}

	if err := createStyleNav(zipWriter, tempDir); err != nil {
		return err
	}
	
	if err := createNav(zipWriter, tempDir); err != nil {
		return err
	}

	if err := createToc(zipWriter, tempDir); err != nil {
		return err
	}


	if err := AddCoverImage(zipWriter, img_bytes, img_type); err != nil {
		return err
	}


	return nil

}