// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// docxBuilder writes a Word document (OOXML) with headings, paragraphs, tables, PNG images and page numbers in the
// footer, enough for a report without a document library
type docxBuilder struct {
	body   strings.Builder
	images [][]byte
}

type docxCell struct {
	Text  string
	Color string
	Bold  bool
}

const (
	docxColorPass = "15803D"
	docxColorWarn = "B45309"
	docxColorFail = "B91C1C"
	docxColorGray = "6B7280"
)

func docxEscape(text string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(text))
	return buf.String()
}

func docxRun(text string, bold bool, color string, size int) string {
	props := ""
	if bold {
		props += "<w:b/><w:bCs/>"
	}
	if color != "" {
		props += fmt.Sprintf(`<w:color w:val="%s"/>`, color)
	}
	if size > 0 {
		props += fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, size, size)
	}
	if props != "" {
		props = "<w:rPr>" + props + "</w:rPr>"
	}
	runs := []string{}
	for i, line := range strings.Split(text, "\n") {
		br := ""
		if i > 0 {
			br = "<w:br/>"
		}
		runs = append(runs, fmt.Sprintf(`<w:r>%s%s<w:t xml:space="preserve">%s</w:t></w:r>`, props, br, docxEscape(line)))
	}
	return strings.Join(runs, "")
}

func (d *docxBuilder) paragraph(style string, align string, content string) {
	props := ""
	if style != "" {
		props += fmt.Sprintf(`<w:pStyle w:val="%s"/>`, style)
	}
	if align != "" {
		props += fmt.Sprintf(`<w:jc w:val="%s"/>`, align)
	}
	if props != "" {
		props = "<w:pPr>" + props + "</w:pPr>"
	}
	d.body.WriteString("<w:p>" + props + content + "</w:p>")
}

func (d *docxBuilder) Title(text string) {
	d.paragraph("Title", "", docxRun(text, false, "", 0))
}

func (d *docxBuilder) Heading(level int, text string) {
	d.paragraph(fmt.Sprintf("Heading%d", level), "", docxRun(text, false, "", 0))
}

func (d *docxBuilder) Paragraph(text string) {
	d.paragraph("", "", docxRun(text, false, "", 0))
}

func (d *docxBuilder) Note(text string) {
	d.paragraph("Note", "", docxRun(text, false, "", 0))
}

// Rich writes a paragraph of runs, each one a text with its own bold and color
func (d *docxBuilder) Rich(cells ...docxCell) {
	content := ""
	for _, cell := range cells {
		content += docxRun(cell.Text, cell.Bold, cell.Color, 0)
	}
	d.paragraph("", "", content)
}

func (d *docxBuilder) Bullet(text string) {
	d.paragraph("ListBullet", "", docxRun("•\t"+text, false, "", 0))
}

func (d *docxBuilder) PageBreak() {
	d.body.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
}

// Table writes a table whose first row is the shaded header repeated on every page; widths are in percent
func (d *docxBuilder) Table(headers []string, widths []int, rows [][]docxCell) {
	const total = 9638 // the text width of an A4 page with 2 cm margins, in twentieths of a point
	d.body.WriteString(`<w:tbl><w:tblPr><w:tblStyle w:val="ReportTable"/><w:tblW w:w="5000" w:type="pct"/><w:tblLayout w:type="fixed"/></w:tblPr><w:tblGrid>`)
	for _, width := range widths {
		d.body.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, total*width/100))
	}
	d.body.WriteString(`</w:tblGrid>`)

	cell := func(width int, content string, shaded bool) string {
		shading := ""
		if shaded {
			shading = `<w:shd w:val="clear" w:color="auto" w:fill="DCE6F2"/>`
		}
		return fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>%s</w:tcPr><w:p><w:pPr><w:pStyle w:val="TableText"/></w:pPr>%s</w:p></w:tc>`, total*width/100, shading, content)
	}
	if len(headers) > 0 {
		d.body.WriteString(`<w:tr><w:trPr><w:tblHeader/><w:cantSplit/></w:trPr>`)
		for i, header := range headers {
			d.body.WriteString(cell(widths[i], docxRun(header, true, "", 0), true))
		}
		d.body.WriteString(`</w:tr>`)
	}
	for _, row := range rows {
		d.body.WriteString(`<w:tr><w:trPr><w:cantSplit/></w:trPr>`)
		for i := range widths {
			value := docxCell{}
			if i < len(row) {
				value = row[i]
			}
			d.body.WriteString(cell(widths[i], docxRun(value.Text, value.Bold, value.Color, 0), false))
		}
		d.body.WriteString(`</w:tr>`)
	}
	d.body.WriteString(`</w:tbl>`)
	d.paragraph("Spacer", "", "")
}

// Image places a PNG as wide as the given centimeters, keeping its aspect ratio
func (d *docxBuilder) Image(png []byte, pixelWidth int, pixelHeight int, widthCm float64) {
	d.images = append(d.images, png)
	index := len(d.images)
	cx := int64(widthCm * 360000)
	cy := cx * int64(pixelHeight) / int64(pixelWidth)
	drawing := fmt.Sprintf(`<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="%d" cy="%d"/><wp:docPr id="%d" name="Chart %d"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="%d" name="image%d.png"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="rIdImage%d"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`,
		cx, cy, index, index, index, index, index, cx, cy)
	d.paragraph("", "center", drawing)
}

const docxNamespaces = `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"`

// docxStyles sets 宋体 for Chinese body text and 微软雅黑 for headings, with Calibri and Arial for Latin text
func docxStyles() string {
	font := func(latin string, eastAsia string) string {
		return fmt.Sprintf(`<w:rFonts w:ascii="%s" w:hAnsi="%s" w:eastAsia="%s" w:cs="%s"/>`, latin, latin, eastAsia, latin)
	}
	heading := func(id string, name string, size int, before int, color string) string {
		return fmt.Sprintf(`<w:style w:type="paragraph" w:styleId="%s"><w:name w:val="%s"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:qFormat/><w:pPr><w:keepNext/><w:keepLines/><w:spacing w:before="%d" w:after="120"/><w:outlineLvl w:val="%d"/></w:pPr><w:rPr>%s<w:b/><w:bCs/><w:color w:val="%s"/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr></w:style>`,
			id, name, before, int(id[len(id)-1]-'1'), font("Arial", "微软雅黑"), color, size, size)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:docDefaults><w:rPrDefault><w:rPr>` + font("Calibri", "宋体") + `<w:sz w:val="21"/><w:szCs w:val="21"/><w:lang w:val="en-US" w:eastAsia="zh-CN" w:bidi="ar-SA"/></w:rPr></w:rPrDefault>` +
		`<w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="300" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults>` +
		`<w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style>` +
		`<w:style w:type="character" w:default="1" w:styleId="DefaultParagraphFont"><w:name w:val="Default Paragraph Font"/><w:uiPriority w:val="1"/><w:semiHidden/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:qFormat/><w:pPr><w:spacing w:before="1200" w:after="360"/><w:jc w:val="center"/></w:pPr><w:rPr>` + font("Arial", "微软雅黑") + `<w:b/><w:bCs/><w:color w:val="1F3864"/><w:sz w:val="44"/><w:szCs w:val="44"/></w:rPr></w:style>` +
		heading("Heading1", "heading 1", 32, 360, "1F3864") +
		heading("Heading2", "heading 2", 26, 240, "2E5597") +
		`<w:style w:type="paragraph" w:styleId="Note"><w:name w:val="Note"/><w:basedOn w:val="Normal"/><w:qFormat/><w:rPr><w:color w:val="6B7280"/><w:sz w:val="18"/><w:szCs w:val="18"/></w:rPr></w:style>` +
		`<w:style w:type="paragraph" w:styleId="ListBullet"><w:name w:val="List Bullet"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:tabs><w:tab w:val="left" w:pos="360"/></w:tabs><w:spacing w:after="80"/><w:ind w:left="360" w:hanging="360"/></w:pPr></w:style>` +
		`<w:style w:type="paragraph" w:styleId="TableText"><w:name w:val="Table Text"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:spacing w:before="20" w:after="20" w:line="240" w:lineRule="auto"/></w:pPr><w:rPr><w:sz w:val="18"/><w:szCs w:val="18"/></w:rPr></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Spacer"><w:name w:val="Spacer"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="60" w:line="120" w:lineRule="exact"/></w:pPr></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Footer"><w:name w:val="footer"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="0"/></w:pPr><w:rPr><w:color w:val="6B7280"/><w:sz w:val="18"/><w:szCs w:val="18"/></w:rPr></w:style>` +
		`<w:style w:type="table" w:default="1" w:styleId="TableNormal"><w:name w:val="Normal Table"/><w:uiPriority w:val="99"/><w:semiHidden/><w:tblPr><w:tblInd w:w="0" w:type="dxa"/><w:tblCellMar><w:top w:w="0" w:type="dxa"/><w:left w:w="108" w:type="dxa"/><w:bottom w:w="0" w:type="dxa"/><w:right w:w="108" w:type="dxa"/></w:tblCellMar></w:tblPr></w:style>` +
		`<w:style w:type="table" w:styleId="ReportTable"><w:name w:val="Report Table"/><w:basedOn w:val="TableNormal"/><w:tblPr><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/><w:left w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/><w:right w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/></w:tblBorders><w:tblCellMar><w:top w:w="40" w:type="dxa"/><w:left w:w="100" w:type="dxa"/><w:bottom w:w="40" w:type="dxa"/><w:right w:w="100" w:type="dxa"/></w:tblCellMar></w:tblPr></w:style>` +
		`</w:styles>`
}

// Bytes packs the document, the footer shows the given text and "page X of Y" in the language
func (d *docxBuilder) Bytes(title string, footer string, pageFormat string) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)
	add := func(name string, content string) error {
		f, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: time.Now()})
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(content))
		return err
	}

	imageRels := ""
	for i := range d.images {
		imageRels += fmt.Sprintf(`<Relationship Id="rIdImage%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image%d.png"/>`, i+1, i+1)
	}

	// "第 {PAGE} 页 共 {NUMPAGES} 页" or "Page {PAGE} of {NUMPAGES}"
	field := func(name string) string {
		return fmt.Sprintf(`<w:fldSimple w:instr=" %s "><w:r><w:t>1</w:t></w:r></w:fldSimple>`, name)
	}
	pageText := ""
	for i, part := range strings.Split(pageFormat, "%s") {
		if i == 1 {
			pageText += field("PAGE")
		} else if i == 2 {
			pageText += field("NUMPAGES")
		}
		if part != "" {
			pageText += docxRun(part, false, "", 0)
		}
	}
	footerXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:ftr ` + docxNamespaces + `><w:p><w:pPr><w:pStyle w:val="Footer"/></w:pPr>` + docxRun(footer+"    ", false, "", 0) + pageText + `</w:p></w:ftr>`

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	parts := []struct{ name, content string }{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/><Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`},
		{"docProps/core.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>` + docxEscape(title) + `</dc:title><dc:creator>Casdoor</dc:creator><cp:lastModifiedBy>Casdoor</cp:lastModifiedBy><dcterms:created xsi:type="dcterms:W3CDTF">` + now + `</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">` + now + `</dcterms:modified></cp:coreProperties>`},
		{"docProps/app.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>Casdoor</Application></Properties>`},
		{"word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rIdSettings" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/><Relationship Id="rIdFooter1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>` + imageRels + `</Relationships>`},
		{"word/settings.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:zoom w:percent="100"/><w:updateFields w:val="false"/><w:defaultTabStop w:val="420"/><w:characterSpacingControl w:val="compressPunctuation"/><w:compat><w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/></w:compat></w:settings>`},
		{"word/styles.xml", docxStyles()},
		{"word/footer1.xml", footerXml},
		{"word/document.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNamespaces + `><w:body>` + d.body.String() +
			`<w:sectPr><w:footerReference w:type="default" r:id="rIdFooter1"/><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1247" w:right="1134" w:bottom="1247" w:left="1134" w:header="567" w:footer="567" w:gutter="0"/></w:sectPr></w:body></w:document>`},
	}
	for _, part := range parts {
		if err := add(part.name, part.content); err != nil {
			return nil, err
		}
	}
	for i, image := range d.images {
		f, err := w.CreateHeader(&zip.FileHeader{Name: fmt.Sprintf("word/media/image%d.png", i+1), Method: zip.Store, Modified: time.Now()})
		if err != nil {
			return nil, err
		}
		if _, err = f.Write(image); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
