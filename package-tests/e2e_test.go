package packagetests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"wattpad-to-ebook/ebook"
	"wattpad-to-ebook/wattpad_stories"
	"github.com/stretchr/testify/require"
	"github.com/yosssi/gohtml"
)



func Test_wattpad_noImage(t *testing.T) {
	url := "https://www.wattpad.com/story/389173089-manager%27s-duties"
	chapters, metadata, err := wattpadstories.Get_Chapters(url)
	
	require.NotEmpty(t, chapters, "Era para ter os capítulos aqui, mas não tem")
	require.NotEmpty(t, metadata, "Era para ter os metadados da história aqui, mas não tem")
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	
	epubName := fmt.Sprintf("%s - %s.epub", metadata.Name, metadata.Author)
	
	tempDir, err := ebook.Setup_temp()
	
	require.DirExists(t, tempDir, "era para o diretório temporário existir, mas não existe")
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)

	err = ebook.Setup_container(tempDir)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "META-INF", "container.xml"), "era para o container.xml existir, mas não existe")
	
hasImages := false

for _, chapter := range chapters {
    bodyBytes, err := wattpadstories.Get_Chapter_Text(chapter.URL)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)

    modifiedBody, foundImage, err := wattpadstories.DownloadAndRewriteImages(bodyBytes, tempDir, chapter.Index)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
    require.NotEmpty(t, modifiedBody)

    if foundImage {
        hasImages = true
    }

    pretty := gohtml.Format(modifiedBody)
    err = ebook.AddChapters(pretty, chapter.Index, tempDir, chapter.Title)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
}

imgPath := filepath.Join(tempDir, "images")

var imgDir []os.DirEntry
if hasImages {
    require.DirExists(t, imgPath, "era para o diretório 'images' existir, mas não existe")
    imgDir, err = os.ReadDir(imgPath)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
    require.NotEmpty(t, imgDir, "essa histto 'imgDir' não era para estar vazio, mas está")
} else {
    require.NoDirExists(t, imgPath, "era para o diretório 'images' não existir, mas existe")
}

	err = ebook.Setup_content(tempDir, len(chapters), metadata.Name, metadata.Author, metadata.Description, metadata.CoverImageType, imgDir)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "OEBPS", "content.opf"), "era para o 'content.opf' existir, mas não existe")
	
	
	err = ebook.Setup_CSS(tempDir)
	
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "style", "main.css"), "era para o 'main.css' existir, mas não existe")
	
	var nav_chapters []ebook.ChapterNavItem
	
	for i, chap := range chapters {
		nav_chapters = append(nav_chapters, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}
	
	require.NotEmpty(t, nav_chapters, "era para ter os capítulos na variável, mas não tem")
	
	err = ebook.Setup_Nav(tempDir, nav_chapters, metadata.Name)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "style", "nav.css"), "era para o 'nav.css' existir, mas não existe")

	for i, _ := range chapters {
		chapter_file := filepath.Join(tempDir, "OEBPS", fmt.Sprintf("chapter_%d.xhtml", i+1))
		require.FileExistsf(t, chapter_file, "era para o arquivo: '%s' existir, mas não existe", chapter_file)
	}
	
	var chap_list = []ebook.ChapterNavItem{}
	
	for i, chap := range chapters {
		chap_list = append(chap_list, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}
	require.NotEmpty(t, chap_list, "era para ter o corpo da navegação de capítulos, mas não tem")

	err = ebook.SetupToc(tempDir, metadata.Name, chap_list)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)


	err = ebook.Make_Ebook(tempDir, epubName, metadata.CoverImage, metadata.CoverImageType, hasImages)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, epubName, "era para o epub ter sido criado, mas não foi")
	os.RemoveAll(tempDir)	
	require.NoDirExists(t, tempDir, "o diretório temporário não era para existir mais até esse ponto, mas ainda existe")
	os.Remove(epubName)
}



func Test_wattpad_withImage(t *testing.T) {
	url := "https://www.wattpad.com/story/388706112-sole-elite-disclosed-classroom-of-the-elite"
	chapters, metadata, err := wattpadstories.Get_Chapters(url)
	
	require.NotEmpty(t, chapters, "Era para ter os capítulos aqui, mas não tem")
	require.NotEmpty(t, metadata, "Era para ter os metadados da história aqui, mas não tem")
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	
	epubName := fmt.Sprintf("%s - %s.epub", metadata.Name, metadata.Author)
	
	tempDir, err := ebook.Setup_temp()
	
	require.DirExists(t, tempDir, "era para o diretório temporário existir, mas não existe")
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)

	err = ebook.Setup_container(tempDir)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "META-INF", "container.xml"), "era para o container.xml existir, mas não existe")
	
hasImages := false

for _, chapter := range chapters {
    bodyBytes, err := wattpadstories.Get_Chapter_Text(chapter.URL)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)

    modifiedBody, foundImage, err := wattpadstories.DownloadAndRewriteImages(bodyBytes, tempDir, chapter.Index)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
    require.NotEmpty(t, modifiedBody)

    if foundImage {
        hasImages = true
    }

    pretty := gohtml.Format(modifiedBody)
    err = ebook.AddChapters(pretty, chapter.Index, tempDir, chapter.Title)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
}

imgPath := filepath.Join(tempDir, "images")

var imgDir []os.DirEntry
if hasImages {
    require.DirExists(t, imgPath, "era para o diretório 'images' existir, mas não existe")
    imgDir, err = os.ReadDir(imgPath)
    require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
    require.NotEmpty(t, imgDir, "essa histto 'imgDir' não era para estar vazio, mas está")
} else {
    require.NoDirExists(t, imgPath, "era para o diretório 'images' não existir, mas existe")
}

	err = ebook.Setup_content(tempDir, len(chapters), metadata.Name, metadata.Author, metadata.Description, metadata.CoverImageType, imgDir)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "OEBPS", "content.opf"), "era para o 'content.opf' existir, mas não existe")
	
	
	err = ebook.Setup_CSS(tempDir)
	
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "style", "main.css"), "era para o 'main.css' existir, mas não existe")
	
	var nav_chapters []ebook.ChapterNavItem
	
	for i, chap := range chapters {
		nav_chapters = append(nav_chapters, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}
	
	require.NotEmpty(t, nav_chapters, "era para ter os capítulos na variável, mas não tem")
	
	err = ebook.Setup_Nav(tempDir, nav_chapters, metadata.Name)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, filepath.Join(tempDir, "style", "nav.css"), "era para o 'nav.css' existir, mas não existe")

	for i, _ := range chapters {
		chapter_file := filepath.Join(tempDir, "OEBPS", fmt.Sprintf("chapter_%d.xhtml", i+1))
		require.FileExistsf(t, chapter_file, "era para o arquivo: '%s' existir, mas não existe", chapter_file)
	}
	
	var chap_list = []ebook.ChapterNavItem{}
	
	for i, chap := range chapters {
		chap_list = append(chap_list, ebook.ChapterNavItem{Href: fmt.Sprintf("chapter_%d.xhtml", i+1), Title: chap.Title})
	}
	require.NotEmpty(t, chap_list, "era para ter o corpo da navegação de capítulos, mas não tem")

	err = ebook.SetupToc(tempDir, metadata.Name, chap_list)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)


	err = ebook.Make_Ebook(tempDir, epubName, metadata.CoverImage, metadata.CoverImageType, hasImages)
	require.Nilf(t, err, "Não era pra ter erro, mas tem\nErro: ", err)
	require.FileExists(t, epubName, "era para o epub ter sido criado, mas não foi")
	os.RemoveAll(tempDir)	
	require.NoDirExists(t, tempDir, "o diretório temporário não era para existir mais até esse ponto, mas ainda existe")
	os.Remove(epubName)
}