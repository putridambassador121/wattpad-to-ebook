package wattpadstories

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/gabriel-vasile/mimetype"
)


type Story_Metadata struct {
	Name string
	Author string
	Description string
	CoverImage []byte
	CoverImageType string
}

type Story_Chapters struct {
	Index int
	Title string
	URL string
}



func getReader(resp *http.Response) (io.ReadCloser, error) {
  enc := strings.ToLower(resp.Header.Get("Content-Encoding"))

  switch {
  case strings.Contains(enc, "gzip"):
    return gzip.NewReader(resp.Body)
  case strings.Contains(enc, "deflate"):
    return io.NopCloser(flate.NewReader(resp.Body)), nil
  case strings.Contains(enc, "br"):
    return io.NopCloser(brotli.NewReader(resp.Body)), nil
  case strings.Contains(enc, "zstd"):
    dec, err := zstd.NewReader(resp.Body)
    if err != nil {
      return nil, err
    }
    // NewReader já devolve um io.ReadCloser
    return dec.IOReadCloser(), nil
  default:
    return resp.Body, nil
  }
}

func get_Image(img_url string) ([]byte, string, error) {

	resp, err := http.Get(img_url)

	if err != nil {
		return nil, "", err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, "", err
	}

	return body, resp.Header.Get("content-type"),nil
}

func Get_Chapters(story_url string, ) ([]Story_Chapters, Story_Metadata, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", story_url, nil)

	if err != nil {
		return nil, Story_Metadata{}, err
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	resp, err := client.Do(req)

	if err != nil {
	return nil, Story_Metadata{}, err
	}

	
	if resp.StatusCode != 200 {
		return nil, Story_Metadata{}, fmt.Errorf("deu algum problema aqui: o código é %d", resp.StatusCode)
	}
	
	body, err := getReader(resp)
	
	if err != nil {
		return nil, Story_Metadata{}, err
	}
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, Story_Metadata{}, err
	}
	
	// chap_html := `div[data-testid="toc"] ul[aria-label="story-parts"]`
	var story_metadata Story_Metadata

	var chapter_list []Story_Chapters

	title := doc.Find(`div.gF-N5`).Text()
	author := doc.Find(`div[data-testid="story-badges"] a`).Text()
	description := doc.Find("pre.mpshL._6pPkw").Text()
	cover_img_url, _ := doc.Find("img.cover__BlyZa").Attr("src")
	cover_img_bytes, imgtype, err := get_Image(cover_img_url)
	
	if err != nil {
		return []Story_Chapters{}, Story_Metadata{}, err
	}

	story_metadata.Name = title
	story_metadata.Author = author
	story_metadata.Description = description
	story_metadata.CoverImage = cover_img_bytes
	story_metadata.CoverImageType = imgtype

	chapter_finder := doc.Find(`div[data-testid="toc"] ul[aria-label="story-parts"]`)

	chapter_finder.Find("a").Each(func(i int, s *goquery.Selection) {
    href, exists := s.Attr("href")
    if exists {
		chapter_list = append(chapter_list,
    Story_Chapters{Index: i+1, Title: s.Find("div.wpYp-").Text(), URL: href},
	)	
    }
	
	})

	return chapter_list, story_metadata, nil
}



func Get_Chapter_Text(chapter_url string) ([]byte, error) {

	client := &http.Client{}

	id := strings.Split(chapter_url[24:], "-")[0]
	
	
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.wattpad.com/apiv2/?m=storytext&id=%s&page=0", id), nil)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := getReader(resp)
	
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(body)
	
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func DownloadAndRewriteImages(htmlContent []byte, tempDir string, chapIndex int) (string, bool,error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return "", false, err
	}

	foundAnyImage := false

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
    src, exists := s.Attr("src")
    if !exists || strings.TrimSpace(src) == "" {
        return
    }

    // Download da imagem
    res, err := http.Get(src)
    if err != nil {
        return
    }
    defer res.Body.Close()

    // Detectar tipo e salvar imagem
    buf, err := io.ReadAll(res.Body)
    if err != nil {
        return
    }

    contentType := mimetype.Detect(buf)

    filename := fmt.Sprintf("chapter%d_img%d%s", chapIndex, i, contentType.Extension())
    path := filepath.Join(tempDir, "images", filename)

    os.MkdirAll(filepath.Join(tempDir, "images"), os.ModePerm)
    os.WriteFile(path, buf, 0644)

    s.SetAttr("src", fmt.Sprintf("../images/%s", filename))
    s.SetAttr("width", "100%")

    foundAnyImage = true
})

htmlBody, err := doc.Html()

if err != nil {
	return "", false, nil
}

return htmlBody, foundAnyImage, nil
}